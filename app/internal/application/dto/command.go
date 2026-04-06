package dto

import "backend/internal/domain/transaction"

type SetValueCommand struct {
	Value uint64 `json:"value"`
}

type SetValueResult struct {
	RequestID string             `json:"request_id"`
	TxHash    string             `json:"tx_hash"`
	Status    transaction.Status `json:"status"`
}

type SyncResult struct {
	Processed int `json:"processed"`
	Updated   int `json:"updated"`
}
