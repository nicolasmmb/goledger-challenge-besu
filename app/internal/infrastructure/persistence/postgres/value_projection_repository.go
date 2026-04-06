package postgres

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"backend/internal/domain/shared"
	"backend/internal/domain/value"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ValueProjectionRepository struct {
	pool *pgxpool.Pool
}

func NewValueProjectionRepository(pool *pgxpool.Pool) *ValueProjectionRepository {
	return &ValueProjectionRepository{pool: pool}
}

func (r *ValueProjectionRepository) Upsert(ctx context.Context, p value.Projection) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO contract_state_projection (contract_address, current_value, last_block_number, last_tx_hash, updated_at)
		VALUES ($1,$2,$3,$4,$5)
		ON CONFLICT (contract_address)
		DO UPDATE SET
			current_value = EXCLUDED.current_value,
			last_block_number = EXCLUDED.last_block_number,
			last_tx_hash = EXCLUDED.last_tx_hash,
			updated_at = EXCLUDED.updated_at
		WHERE EXCLUDED.last_block_number >= contract_state_projection.last_block_number
	`,
		p.ContractAddress,
		strconv.FormatUint(p.CurrentValue, 10),
		int64(p.LastBlockNumber),
		p.LastTxHash,
		p.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("upsert value projection: %w", err)
	}
	return nil
}

func (r *ValueProjectionRepository) GetByContract(ctx context.Context, contractAddress string) (*value.Projection, error) {
	var p value.Projection
	var rawValue string
	var block int64
	err := r.pool.QueryRow(ctx, `
		SELECT contract_address, current_value, last_block_number, last_tx_hash, updated_at
		FROM contract_state_projection
		WHERE contract_address = $1
	`, contractAddress).Scan(
		&p.ContractAddress,
		&rawValue,
		&block,
		&p.LastTxHash,
		&p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, shared.ErrNotFound
		}
		return nil, fmt.Errorf("query value projection: %w", err)
	}
	currentValue, err := strconv.ParseUint(rawValue, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("parse current value: %w", err)
	}
	p.CurrentValue = currentValue
	p.LastBlockNumber = uint64(block)
	return &p, nil
}

func SeedProjection(contractAddress string, currentValue uint64) value.Projection {
	return value.Projection{
		ContractAddress: contractAddress,
		CurrentValue:    currentValue,
		LastBlockNumber: 0,
		LastTxHash:      "",
		UpdatedAt:       time.Now().UTC(),
	}
}
