package command

import (
	"context"
	"fmt"

	"backend/internal/application/dto"
	"backend/internal/domain/shared"
	"backend/internal/domain/transaction"
	"backend/internal/domain/value"

	"github.com/google/uuid"
)

type SetValueService struct {
	writer          value.ContractWriter
	transactionRepo transaction.Repository
	fromAddress     string
	contractAddress string
}

func NewSetValueService(
	writer value.ContractWriter,
	transactionRepo transaction.Repository,
	fromAddress string,
	contractAddress string,
) *SetValueService {
	return &SetValueService{
		writer:          writer,
		transactionRepo: transactionRepo,
		fromAddress:     fromAddress,
		contractAddress: contractAddress,
	}
}

func (s *SetValueService) Execute(ctx context.Context, cmd dto.SetValueCommand) (dto.SetValueResult, error) {
	requestID := uuid.NewString()

	txHash, err := s.writer.SetValue(ctx, cmd.Value)
	if err != nil {
		return dto.SetValueResult{}, fmt.Errorf("set contract value: %w", err)
	}

	tx := transaction.NewSubmitted(
		requestID,
		shared.NormalizeHash(txHash),
		cmd.Value,
		s.fromAddress,
		s.contractAddress,
		shared.NowUTC(),
	)
	if err := s.transactionRepo.Save(ctx, tx); err != nil {
		return dto.SetValueResult{}, fmt.Errorf("save transaction: %w", err)
	}

	return dto.SetValueResult{
		RequestID: requestID,
		TxHash:    tx.TxHash,
		Status:    tx.Status,
	}, nil
}
