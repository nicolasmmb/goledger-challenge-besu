package value

import "time"

type Projection struct {
	ContractAddress string
	CurrentValue    uint64
	LastBlockNumber uint64
	LastTxHash      string
	UpdatedAt       time.Time
}

func (p Projection) CanAdvance(block uint64) bool {
	return block >= p.LastBlockNumber
}
