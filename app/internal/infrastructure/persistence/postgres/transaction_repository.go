package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"backend/internal/domain/shared"
	"backend/internal/domain/transaction"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TransactionRepository struct {
	pool *pgxpool.Pool
}

func NewTransactionRepository(pool *pgxpool.Pool) *TransactionRepository {
	return &TransactionRepository{pool: pool}
}

func (r *TransactionRepository) Save(ctx context.Context, tx *transaction.Transaction) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO contract_writes
		(request_id, tx_hash, value, from_address, contract_address, status, block_number, error_message, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
	`,
		tx.RequestID,
		tx.TxHash,
		fmt.Sprintf("%d", tx.Value),
		tx.FromAddress,
		tx.ContractAddress,
		tx.Status,
		tx.BlockNumber,
		tx.ErrorMessage,
		tx.CreatedAt,
		tx.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert transaction: %w", err)
	}
	return nil
}

func (r *TransactionRepository) FindByHash(ctx context.Context, txHash string) (*transaction.Transaction, error) {
	var tx transaction.Transaction
	var status string
	var valueText string
	var block *int64
	err := r.pool.QueryRow(ctx, `
		SELECT request_id, tx_hash, value, from_address, contract_address, status, block_number, error_message, created_at, updated_at
		FROM contract_writes
		WHERE tx_hash = $1
	`, shared.NormalizeHash(txHash)).Scan(
		&tx.RequestID,
		&tx.TxHash,
		&valueText,
		&tx.FromAddress,
		&tx.ContractAddress,
		&status,
		&block,
		&tx.ErrorMessage,
		&tx.CreatedAt,
		&tx.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, transaction.ErrTransactionNotFound
		}
		return nil, fmt.Errorf("query transaction by hash: %w", err)
	}

	var valueU64 uint64
	if _, err := fmt.Sscanf(valueText, "%d", &valueU64); err != nil {
		return nil, fmt.Errorf("parse transaction value: %w", err)
	}
	tx.Value = valueU64
	tx.Status = transaction.Status(status)
	if block != nil {
		v := uint64(*block)
		tx.BlockNumber = &v
	}
	return &tx, nil
}

func (r *TransactionRepository) FindPending(ctx context.Context, limit int) ([]*transaction.Transaction, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT request_id, tx_hash, value, from_address, contract_address, status, block_number, error_message, created_at, updated_at
		FROM contract_writes
		WHERE status IN ('submitted','mined')
		ORDER BY created_at ASC, id ASC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("query pending transactions: %w", err)
	}
	defer rows.Close()

	items := make([]*transaction.Transaction, 0, limit)
	for rows.Next() {
		var tx transaction.Transaction
		var status string
		var valueText string
		var block *int64
		if err := rows.Scan(
			&tx.RequestID,
			&tx.TxHash,
			&valueText,
			&tx.FromAddress,
			&tx.ContractAddress,
			&status,
			&block,
			&tx.ErrorMessage,
			&tx.CreatedAt,
			&tx.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan pending transaction: %w", err)
		}
		var valueU64 uint64
		if _, err := fmt.Sscanf(valueText, "%d", &valueU64); err != nil {
			return nil, fmt.Errorf("parse transaction value: %w", err)
		}
		tx.Value = valueU64
		tx.Status = transaction.Status(status)
		if block != nil {
			v := uint64(*block)
			tx.BlockNumber = &v
		}
		items = append(items, &tx)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate pending transactions: %w", err)
	}
	return items, nil
}

func (r *TransactionRepository) UpdateStatus(ctx context.Context, in transaction.UpdateStatusInput) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE contract_writes
		SET status = $2, block_number = $3, error_message = $4, updated_at = $5
		WHERE tx_hash = $1
	`, shared.NormalizeHash(in.TxHash), in.Status, in.BlockNumber, in.ErrorMessage, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("update transaction status: %w", err)
	}
	return nil
}
