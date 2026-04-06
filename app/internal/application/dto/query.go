package dto

import "backend/internal/domain/transaction"

type GetValueQuery struct {
	Source string
}

type ValueResult struct {
	Value  uint64 `json:"value"`
	Source string `json:"source"`
}

type GetTransactionStatusResult struct {
	TxHash       string             `json:"tx_hash"`
	Status       transaction.Status `json:"status"`
	BlockNumber  *uint64            `json:"block_number,omitempty"`
	ErrorMessage string             `json:"error_message,omitempty"`
}

type CheckConsistencyResult struct {
	OnChain  uint64 `json:"on_chain"`
	InDB     uint64 `json:"in_db"`
	Contract string `json:"contract_address"`
	Match    bool   `json:"match"`
}
