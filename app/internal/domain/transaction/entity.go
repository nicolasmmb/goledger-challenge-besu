package transaction

import (
	"fmt"
	"time"
)

type Transaction struct {
	RequestID       string
	TxHash          string
	Value           uint64
	FromAddress     string
	ContractAddress string
	Status          Status
	BlockNumber     *uint64
	ErrorMessage    string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func NewSubmitted(
	requestID string,
	txHash string,
	value uint64,
	fromAddress string,
	contractAddress string,
	now time.Time,
) *Transaction {
	return &Transaction{
		RequestID:       requestID,
		TxHash:          txHash,
		Value:           value,
		FromAddress:     fromAddress,
		ContractAddress: contractAddress,
		Status:          StatusSubmitted,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}

func (t *Transaction) TransitionTo(next Status, blockNumber *uint64, errMsg string, now time.Time) error {
	if !validTransition(t.Status, next) {
		return fmt.Errorf("%w: %s -> %s", ErrInvalidStatusTransition, t.Status, next)
	}
	t.Status = next
	t.BlockNumber = blockNumber
	t.ErrorMessage = errMsg
	t.UpdatedAt = now
	return nil
}

func validTransition(from, to Status) bool {
	switch from {
	case StatusSubmitted:
		return to == StatusMined || to == StatusFailed || to == StatusSubmitted
	case StatusMined:
		return to == StatusConfirmed || to == StatusFailed || to == StatusMined
	case StatusConfirmed, StatusFailed:
		return to == from
	default:
		return false
	}
}
