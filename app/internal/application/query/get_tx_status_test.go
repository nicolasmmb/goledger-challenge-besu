package query

import (
	"context"
	"testing"

	"backend/internal/domain/transaction"

	"github.com/stretchr/testify/require"
)

type fakeTxRepo struct {
	tx          *transaction.Transaction
	findErr     error
	updatedWith *transaction.UpdateStatusInput
	updateErr   error
}

func (f *fakeTxRepo) Save(context.Context, *transaction.Transaction) error { return nil }
func (f *fakeTxRepo) FindByHash(context.Context, string) (*transaction.Transaction, error) {
	return f.tx, f.findErr
}
func (f *fakeTxRepo) FindPending(context.Context, int) ([]*transaction.Transaction, error) {
	return nil, nil
}
func (f *fakeTxRepo) UpdateStatus(_ context.Context, in transaction.UpdateStatusInput) error {
	f.updatedWith = &in
	return f.updateErr
}

type fakeTxReader struct {
	receipt      transaction.Receipt
	receiptErr   error
	currentBlock uint64
	currentErr   error
}

func (f *fakeTxReader) GetReceipt(context.Context, string) (transaction.Receipt, error) {
	return f.receipt, f.receiptErr
}
func (f *fakeTxReader) CurrentBlock(context.Context) (uint64, error) {
	return f.currentBlock, f.currentErr
}

func TestGetTransactionStatus_NoRefreshReturnsStored(t *testing.T) {
	repo := &fakeTxRepo{
		tx: &transaction.Transaction{
			TxHash:    "0xabc",
			Status:    transaction.StatusSubmitted,
			RequestID: "r1",
		},
	}
	svc := NewGetTransactionStatusService(repo, &fakeTxReader{}, 2)

	res, err := svc.Execute(context.Background(), "0xabc", false)
	require.NoError(t, err)
	require.Equal(t, transaction.StatusSubmitted, res.Status)
	require.Nil(t, repo.updatedWith)
}

func TestGetTransactionStatus_RefreshToMinedFirst(t *testing.T) {
	repo := &fakeTxRepo{
		tx: &transaction.Transaction{
			TxHash:    "0xabc",
			Status:    transaction.StatusSubmitted,
			RequestID: "r1",
		},
	}
	reader := &fakeTxReader{
		receipt:      transaction.Receipt{Exists: true, Success: true, BlockNumber: 10},
		currentBlock: 13,
	}
	svc := NewGetTransactionStatusService(repo, reader, 2)

	res, err := svc.Execute(context.Background(), "0xabc", true)
	require.NoError(t, err)
	require.Equal(t, transaction.StatusMined, res.Status)
	require.NotNil(t, repo.updatedWith)
	require.Equal(t, transaction.StatusMined, repo.updatedWith.Status)
}

func TestGetTransactionStatus_RefreshReceiptNotFoundKeepsSubmitted(t *testing.T) {
	repo := &fakeTxRepo{
		tx: &transaction.Transaction{
			TxHash:    "0xabc",
			Status:    transaction.StatusSubmitted,
			RequestID: "r1",
		},
	}
	reader := &fakeTxReader{
		receipt: transaction.Receipt{Exists: false},
	}
	svc := NewGetTransactionStatusService(repo, reader, 2)

	res, err := svc.Execute(context.Background(), "0xabc", true)
	require.NoError(t, err)
	require.Equal(t, transaction.StatusSubmitted, res.Status)
	require.Nil(t, repo.updatedWith)
}
