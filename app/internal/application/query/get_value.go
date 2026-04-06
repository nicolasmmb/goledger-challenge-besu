package query

import (
	"context"
	"fmt"
	"strings"

	"backend/internal/application/dto"
	"backend/internal/domain/value"
)

type GetValueService struct {
	chainReader       value.ContractReader
	projectionRepo    value.Repository
	contractAddress   string
	defaultReadSource string
}

func NewGetValueService(
	chainReader value.ContractReader,
	projectionRepo value.Repository,
	contractAddress string,
) *GetValueService {
	return &GetValueService{
		chainReader:       chainReader,
		projectionRepo:    projectionRepo,
		contractAddress:   contractAddress,
		defaultReadSource: "projection",
	}
}

func (s *GetValueService) Execute(ctx context.Context, q dto.GetValueQuery) (dto.ValueResult, error) {
	sourceRaw := strings.ToLower(strings.TrimSpace(q.Source))
	source := sourceRaw
	if source == "" {
		source = s.defaultReadSource
	}

	if source == "chain" {
		v, err := s.chainReader.GetValue(ctx)
		if err != nil {
			return dto.ValueResult{}, fmt.Errorf("get chain value: %w", err)
		}
		return dto.ValueResult{Value: v, Source: "chain"}, nil
	}

	proj, err := s.projectionRepo.GetByContract(ctx, s.contractAddress)
	if err != nil {
		if sourceRaw == "projection" {
			return dto.ValueResult{}, fmt.Errorf("get projection value: %w", err)
		}
		// projection-first with chain fallback.
		v, chainErr := s.chainReader.GetValue(ctx)
		if chainErr != nil {
			return dto.ValueResult{}, fmt.Errorf("get projection value: %w", err)
		}
		return dto.ValueResult{Value: v, Source: "chain"}, nil
	}

	return dto.ValueResult{Value: proj.CurrentValue, Source: "projection"}, nil
}
