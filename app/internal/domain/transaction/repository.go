package transaction

import "context"

type UpdateStatusInput struct {
	TxHash       string
	Status       Status
	BlockNumber  *uint64
	ErrorMessage string
}

type Receipt struct {
	Exists      bool
	Success     bool
	BlockNumber uint64
}

type Repository interface {
	Save(ctx context.Context, tx *Transaction) error
	FindByHash(ctx context.Context, txHash string) (*Transaction, error)
	FindPending(ctx context.Context, limit int) ([]*Transaction, error)
	UpdateStatus(ctx context.Context, in UpdateStatusInput) error
}

type Reader interface {
	GetReceipt(ctx context.Context, txHash string) (Receipt, error)
	CurrentBlock(ctx context.Context) (uint64, error)
}
