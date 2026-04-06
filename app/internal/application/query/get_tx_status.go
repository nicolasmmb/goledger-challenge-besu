package query

import (
	"context"
	"errors"
	"fmt"

	"backend/internal/application/dto"
	"backend/internal/domain/shared"
	"backend/internal/domain/transaction"
)

type GetTransactionStatusService struct {
	transactionRepo transaction.Repository
	chainReader     transaction.Reader
	confirmations   uint64
}

func NewGetTransactionStatusService(
	transactionRepo transaction.Repository,
	chainReader transaction.Reader,
	confirmations uint64,
) *GetTransactionStatusService {
	return &GetTransactionStatusService{
		transactionRepo: transactionRepo,
		chainReader:     chainReader,
		confirmations:   confirmations,
	}
}

func (s *GetTransactionStatusService) Execute(ctx context.Context, txHash string, refresh bool) (dto.GetTransactionStatusResult, error) {
	tx, err := s.transactionRepo.FindByHash(ctx, shared.NormalizeHash(txHash))
	if err != nil {
		return dto.GetTransactionStatusResult{}, fmt.Errorf("find transaction: %w", err)
	}

	if refresh && !transaction.IsTerminalStatus(tx.Status) {
		nextStatus, blockNumber, errMsg, err := s.resolveChainStatus(ctx, tx.TxHash)
		if err != nil {
			return dto.GetTransactionStatusResult{}, fmt.Errorf("refresh transaction status: %w", err)
		}
		nextStatus = normalizeNextStatus(tx.Status, nextStatus)
		if nextStatus != tx.Status {
			if err := tx.TransitionTo(nextStatus, blockNumber, errMsg, shared.NowUTC()); err != nil {
				return dto.GetTransactionStatusResult{}, err
			}
			if err := s.transactionRepo.UpdateStatus(ctx, transaction.UpdateStatusInput{
				TxHash:       tx.TxHash,
				Status:       tx.Status,
				BlockNumber:  tx.BlockNumber,
				ErrorMessage: tx.ErrorMessage,
			}); err != nil {
				return dto.GetTransactionStatusResult{}, fmt.Errorf("persist refreshed status: %w", err)
			}
		}
	}

	return dto.GetTransactionStatusResult{
		TxHash:       tx.TxHash,
		Status:       tx.Status,
		BlockNumber:  tx.BlockNumber,
		ErrorMessage: tx.ErrorMessage,
	}, nil
}

func (s *GetTransactionStatusService) resolveChainStatus(ctx context.Context, txHash string) (transaction.Status, *uint64, string, error) {
	receipt, err := s.chainReader.GetReceipt(ctx, txHash)
	if err != nil {
		return "", nil, "", err
	}
	if !receipt.Exists {
		return transaction.StatusSubmitted, nil, "", nil
	}
	if !receipt.Success {
		block := receipt.BlockNumber
		return transaction.StatusFailed, &block, "transaction reverted", nil
	}

	current, err := s.chainReader.CurrentBlock(ctx)
	if err != nil {
		return "", nil, "", err
	}

	block := receipt.BlockNumber
	if confirmationsReached(current, block, s.confirmations) {
		return transaction.StatusConfirmed, &block, "", nil
	}
	return transaction.StatusMined, &block, "", nil
}

func confirmationsReached(current uint64, txBlock uint64, required uint64) bool {
	if current < txBlock {
		return false
	}
	return current-txBlock >= required
}

func normalizeNextStatus(current transaction.Status, next transaction.Status) transaction.Status {
	if current == transaction.StatusSubmitted && next == transaction.StatusConfirmed {
		return transaction.StatusMined
	}
	return next
}

func IsNotFound(err error) bool {
	return errors.Is(err, transaction.ErrTransactionNotFound) || errors.Is(err, shared.ErrNotFound)
}
