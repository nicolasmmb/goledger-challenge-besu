package query

import (
	"context"
	"fmt"

	"backend/internal/application/dto"
	"backend/internal/domain/value"
)

type CheckConsistencyService struct {
	chainReader     value.ContractReader
	projectionRepo  value.Repository
	contractAddress string
}

func NewCheckConsistencyService(
	chainReader value.ContractReader,
	projectionRepo value.Repository,
	contractAddress string,
) *CheckConsistencyService {
	return &CheckConsistencyService{
		chainReader:     chainReader,
		projectionRepo:  projectionRepo,
		contractAddress: contractAddress,
	}
}

func (s *CheckConsistencyService) Execute(ctx context.Context) (dto.CheckConsistencyResult, error) {
	onChain, err := s.chainReader.GetValue(ctx)
	if err != nil {
		return dto.CheckConsistencyResult{}, fmt.Errorf("get chain value: %w", err)
	}
	proj, err := s.projectionRepo.GetByContract(ctx, s.contractAddress)
	if err != nil {
		return dto.CheckConsistencyResult{}, fmt.Errorf("get projection value: %w", err)
	}
	return dto.CheckConsistencyResult{
		OnChain:  onChain,
		InDB:     proj.CurrentValue,
		Contract: s.contractAddress,
		Match:    onChain == proj.CurrentValue,
	}, nil
}
