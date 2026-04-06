package query

import (
	"context"
	"errors"
	"testing"

	"backend/internal/application/dto"
	"backend/internal/domain/shared"
	"backend/internal/domain/value"

	"github.com/stretchr/testify/require"
)

type fakeContractReader struct {
	value uint64
	err   error
}

func (f fakeContractReader) GetValue(context.Context) (uint64, error) { return f.value, f.err }

type fakeProjectionRepo struct {
	projection *value.Projection
	err        error
}

func (f fakeProjectionRepo) Upsert(context.Context, value.Projection) error { return nil }
func (f fakeProjectionRepo) GetByContract(context.Context, string) (*value.Projection, error) {
	return f.projection, f.err
}

func TestGetValue_SourceChain(t *testing.T) {
	svc := NewGetValueService(fakeContractReader{value: 55}, fakeProjectionRepo{}, "0xcontract")
	res, err := svc.Execute(context.Background(), dto.GetValueQuery{Source: "chain"})
	require.NoError(t, err)
	require.EqualValues(t, 55, res.Value)
	require.Equal(t, "chain", res.Source)
}

func TestGetValue_ProjectionFallbackToChain(t *testing.T) {
	svc := NewGetValueService(
		fakeContractReader{value: 99},
		fakeProjectionRepo{err: shared.ErrNotFound},
		"0xcontract",
	)
	res, err := svc.Execute(context.Background(), dto.GetValueQuery{})
	require.NoError(t, err)
	require.EqualValues(t, 99, res.Value)
	require.Equal(t, "chain", res.Source)
}

func TestGetValue_ProjectionErrorWithoutFallbackWhenForced(t *testing.T) {
	svc := NewGetValueService(
		fakeContractReader{value: 99},
		fakeProjectionRepo{err: errors.New("db down")},
		"0xcontract",
	)
	_, err := svc.Execute(context.Background(), dto.GetValueQuery{Source: "projection"})
	require.Error(t, err)
}
