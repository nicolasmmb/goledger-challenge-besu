package blockchain

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/ethclient"
)

func Dial(ctx context.Context, rpcURL string) (*ethclient.Client, error) {
	client, err := ethclient.DialContext(ctx, rpcURL)
	if err != nil {
		return nil, fmt.Errorf("dial besu rpc: %w", err)
	}
	return client, nil
}
