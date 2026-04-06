package value

import "context"

type Repository interface {
	Upsert(ctx context.Context, p Projection) error
	GetByContract(ctx context.Context, contractAddress string) (*Projection, error)
}

type ContractReader interface {
	GetValue(ctx context.Context) (uint64, error)
}

type ContractWriter interface {
	SetValue(ctx context.Context, value uint64) (string, error)
}
