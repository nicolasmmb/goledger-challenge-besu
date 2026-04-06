package sync

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"backend/internal/application/dto"
	"backend/internal/domain/shared"
	"backend/internal/domain/transaction"
	"backend/internal/domain/value"
)

type ReconcilePendingService struct {
	txRepo          transaction.Repository
	txReader        transaction.Reader
	projectionRepo  value.Repository
	confirmations   uint64
	contractAddress string
	workers         int
}

func NewReconcilePendingService(
	txRepo transaction.Repository,
	txReader transaction.Reader,
	projectionRepo value.Repository,
	confirmations uint64,
	contractAddress string,
	workers ...int,
) *ReconcilePendingService {
	workerCount := 5
	if len(workers) > 0 && workers[0] > 0 {
		workerCount = workers[0]
	}

	return &ReconcilePendingService{
		txRepo:          txRepo,
		txReader:        txReader,
		projectionRepo:  projectionRepo,
		confirmations:   confirmations,
		contractAddress: contractAddress,
		workers:         workerCount,
	}
}

func (s *ReconcilePendingService) Execute(ctx context.Context, limit int) (dto.SyncResult, error) {
	pending, err := s.txRepo.FindPending(ctx, limit)
	if err != nil {
		return dto.SyncResult{}, fmt.Errorf("find pending tx: %w", err)
	}
	if len(pending) == 0 {
		return dto.SyncResult{}, nil
	}

	currentBlock, err := s.txReader.CurrentBlock(ctx)
	if err != nil {
		return dto.SyncResult{}, fmt.Errorf("get current block: %w", err)
	}

	updated, err := s.processPendingConcurrently(ctx, pending, currentBlock)
	if err != nil {
		return dto.SyncResult{}, err
	}
	return dto.SyncResult{Processed: len(pending), Updated: updated}, nil
}

func (s *ReconcilePendingService) processPendingConcurrently(
	ctx context.Context,
	pending []*transaction.Transaction,
	currentBlock uint64,
) (int, error) {
	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	workerCount := min(s.workers, len(pending))
	if workerCount < 1 {
		workerCount = 1
	}

	jobs := make(chan *transaction.Transaction)
	errCh := make(chan error, 1)
	var wg sync.WaitGroup
	var updatedCount int32
	var cancelOnce sync.Once

	worker := func() {
		defer wg.Done()
		for {
			select {
			case <-runCtx.Done():
				return
			case tx, ok := <-jobs:
				if !ok {
					return
				}
				updated, err := s.processSingleTx(runCtx, tx, currentBlock)
				if err != nil {
					select {
					case errCh <- err:
					default:
					}
					cancelOnce.Do(cancel)
					return
				}
				if updated {
					atomic.AddInt32(&updatedCount, 1)
				}
			}
		}
	}

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go worker()
	}

producer:
	for _, tx := range pending {
		select {
		case <-runCtx.Done():
			break producer
		case jobs <- tx:
		}
	}
	close(jobs)
	wg.Wait()

	select {
	case err := <-errCh:
		return int(atomic.LoadInt32(&updatedCount)), err
	default:
	}
	if err := runCtx.Err(); err != nil && err != context.Canceled {
		return int(atomic.LoadInt32(&updatedCount)), err
	}
	return int(atomic.LoadInt32(&updatedCount)), nil
}

func (s *ReconcilePendingService) processSingleTx(ctx context.Context, tx *transaction.Transaction, currentBlock uint64) (bool, error) {
	select {
	case <-ctx.Done():
		return false, ctx.Err()
	default:
	}

	receipt, err := s.txReader.GetReceipt(ctx, tx.TxHash)
	if err != nil {
		return false, fmt.Errorf("get receipt %s: %w", tx.TxHash, err)
	}
	if !receipt.Exists {
		return false, nil
	}

	next, block, msg := reconcileStatus(receipt, currentBlock, s.confirmations)
	next = normalizeSyncStatus(tx.Status, next)
	if err := tx.TransitionTo(next, block, msg, shared.NowUTC()); err != nil {
		return false, err
	}

	if err := s.txRepo.UpdateStatus(ctx, transaction.UpdateStatusInput{
		TxHash:       tx.TxHash,
		Status:       tx.Status,
		BlockNumber:  tx.BlockNumber,
		ErrorMessage: tx.ErrorMessage,
	}); err != nil {
		return false, fmt.Errorf("update tx status %s: %w", tx.TxHash, err)
	}

	if tx.Status == transaction.StatusConfirmed && tx.BlockNumber != nil {
		if err := s.projectionRepo.Upsert(ctx, value.Projection{
			ContractAddress: s.contractAddress,
			CurrentValue:    tx.Value,
			LastBlockNumber: *tx.BlockNumber,
			LastTxHash:      tx.TxHash,
			UpdatedAt:       shared.NowUTC(),
		}); err != nil {
			return false, fmt.Errorf("upsert value projection: %w", err)
		}
	}

	return true, nil
}

func reconcileStatus(receipt transaction.Receipt, currentBlock uint64, confirmations uint64) (transaction.Status, *uint64, string) {
	block := receipt.BlockNumber
	if !receipt.Success {
		return transaction.StatusFailed, &block, "transaction reverted"
	}
	if currentBlock >= block && currentBlock-block >= confirmations {
		return transaction.StatusConfirmed, &block, ""
	}
	return transaction.StatusMined, &block, ""
}

func normalizeSyncStatus(current transaction.Status, next transaction.Status) transaction.Status {
	if current == transaction.StatusSubmitted && next == transaction.StatusConfirmed {
		return transaction.StatusMined
	}
	return next
}
