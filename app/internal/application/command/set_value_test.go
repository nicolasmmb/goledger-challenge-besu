package command

import (
	"context"
	"testing"

	"backend/internal/application/dto"
	"backend/internal/domain/transaction"

	"github.com/stretchr/testify/require"
)

type fakeWriter struct {
	txHash string
	err    error
}

func (f fakeWriter) SetValue(context.Context, uint64) (string, error) { return f.txHash, f.err }

type fakeTxRepo struct {
	last *transaction.Transaction
	err  error
}

func (f *fakeTxRepo) Save(_ context.Context, tx *transaction.Transaction) error {
	f.last = tx
	return f.err
}
func (f *fakeTxRepo) FindByHash(context.Context, string) (*transaction.Transaction, error) {
	return nil, nil
}
func (f *fakeTxRepo) FindPending(context.Context, int) ([]*transaction.Transaction, error) {
	return nil, nil
}
func (f *fakeTxRepo) UpdateStatus(context.Context, transaction.UpdateStatusInput) error { return nil }

func TestSetValueExecute(t *testing.T) {
	repo := &fakeTxRepo{}
	svc := NewSetValueService(fakeWriter{txHash: "0xabc"}, repo, "0xfrom", "0xcontract")

	out, err := svc.Execute(context.Background(), dto.SetValueCommand{Value: 42})
	require.NoError(t, err)
	require.Equal(t, "0xabc", out.TxHash)
	require.NotNil(t, repo.last)
	require.EqualValues(t, 42, repo.last.Value)
}
