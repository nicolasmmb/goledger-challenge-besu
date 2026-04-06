package sync

import (
	"context"
	"sync"
	"testing"
	"time"

	"backend/internal/domain/transaction"
	"backend/internal/domain/value"

	"github.com/stretchr/testify/require"
)

type fakeSyncTxRepo struct {
	pending []*transaction.Transaction
	updates []transaction.UpdateStatusInput
	mu      sync.Mutex
}

func (f *fakeSyncTxRepo) Save(context.Context, *transaction.Transaction) error { return nil }
func (f *fakeSyncTxRepo) FindByHash(context.Context, string) (*transaction.Transaction, error) {
	return nil, nil
}
func (f *fakeSyncTxRepo) FindPending(context.Context, int) ([]*transaction.Transaction, error) {
	return f.pending, nil
}
func (f *fakeSyncTxRepo) UpdateStatus(_ context.Context, in transaction.UpdateStatusInput) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.updates = append(f.updates, in)
	return nil
}

type fakeSyncReader struct {
	receipts map[string]transaction.Receipt
	block    uint64
}

func (f *fakeSyncReader) GetReceipt(_ context.Context, txHash string) (transaction.Receipt, error) {
	r, ok := f.receipts[txHash]
	if !ok {
		return transaction.Receipt{Exists: false}, nil
	}
	return r, nil
}
func (f *fakeSyncReader) CurrentBlock(context.Context) (uint64, error) {
	return f.block, nil
}

type fakeProjectionRepo struct {
	items []value.Projection
	mu    sync.Mutex
}

func (f *fakeProjectionRepo) Upsert(_ context.Context, p value.Projection) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.items = append(f.items, p)
	return nil
}
func (f *fakeProjectionRepo) GetByContract(context.Context, string) (*value.Projection, error) {
	return nil, nil
}

func TestReconcilePending_SubmittedGoesToMinedFirst(t *testing.T) {
	now := time.Now().UTC()
	tx := transaction.NewSubmitted("r1", "0x1", 42, "0xfrom", "0xcontract", now)
	repo := &fakeSyncTxRepo{pending: []*transaction.Transaction{tx}}
	reader := &fakeSyncReader{
		receipts: map[string]transaction.Receipt{
			"0x1": {Exists: true, Success: true, BlockNumber: 10},
		},
		block: 13,
	}
	proj := &fakeProjectionRepo{}
	svc := NewReconcilePendingService(repo, reader, proj, 2, "0xcontract")

	res, err := svc.Execute(context.Background(), 10)
	require.NoError(t, err)
	require.Equal(t, 1, res.Processed)
	require.Equal(t, 1, res.Updated)
	require.Len(t, repo.updates, 1)
	require.Equal(t, transaction.StatusMined, repo.updates[0].Status)
	require.Len(t, proj.items, 0)
}

func TestReconcilePending_ReceiptMissingDoesNotUpdate(t *testing.T) {
	now := time.Now().UTC()
	tx := transaction.NewSubmitted("r1", "0x1", 42, "0xfrom", "0xcontract", now)
	repo := &fakeSyncTxRepo{pending: []*transaction.Transaction{tx}}
	reader := &fakeSyncReader{receipts: map[string]transaction.Receipt{}, block: 20}
	proj := &fakeProjectionRepo{}
	svc := NewReconcilePendingService(repo, reader, proj, 2, "0xcontract")

	res, err := svc.Execute(context.Background(), 10)
	require.NoError(t, err)
	require.Equal(t, 1, res.Processed)
	require.Equal(t, 0, res.Updated)
	require.Len(t, repo.updates, 0)
	require.Len(t, proj.items, 0)
}

func TestReconcilePending_MinedCanConfirmAndProject(t *testing.T) {
	now := time.Now().UTC()
	tx := transaction.NewSubmitted("r1", "0x1", 77, "0xfrom", "0xcontract", now)
	block := uint64(10)
	require.NoError(t, tx.TransitionTo(transaction.StatusMined, &block, "", now.Add(time.Second)))

	repo := &fakeSyncTxRepo{pending: []*transaction.Transaction{tx}}
	reader := &fakeSyncReader{
		receipts: map[string]transaction.Receipt{
			"0x1": {Exists: true, Success: true, BlockNumber: 10},
		},
		block: 13,
	}
	proj := &fakeProjectionRepo{}
	svc := NewReconcilePendingService(repo, reader, proj, 2, "0xcontract")

	res, err := svc.Execute(context.Background(), 10)
	require.NoError(t, err)
	require.Equal(t, 1, res.Processed)
	require.Equal(t, 1, res.Updated)
	require.Len(t, repo.updates, 1)
	require.Equal(t, transaction.StatusConfirmed, repo.updates[0].Status)
	require.Len(t, proj.items, 1)
	require.EqualValues(t, 77, proj.items[0].CurrentValue)
}
