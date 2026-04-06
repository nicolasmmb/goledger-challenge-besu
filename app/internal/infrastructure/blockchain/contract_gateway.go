package blockchain

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"time"

	"backend/internal/domain/shared"
	"backend/internal/domain/transaction"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

const simpleStorageABI = `[{"inputs":[{"internalType":"uint256","name":"x","type":"uint256"}],"name":"set","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[],"name":"get","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"storedData","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"}]`

type Gateway struct {
	client          *ethclient.Client
	contractABI     abi.ABI
	contractAddress common.Address
	privateKey      string
	chainID         *big.Int
}

func NewGateway(
	client *ethclient.Client,
	contractAddress string,
	privateKey string,
	chainID int64,
) (*Gateway, error) {
	parsedABI, err := abi.JSON(strings.NewReader(simpleStorageABI))
	if err != nil {
		return nil, fmt.Errorf("parse simple storage abi: %w", err)
	}

	return &Gateway{
		client:          client,
		contractABI:     parsedABI,
		contractAddress: common.HexToAddress(contractAddress),
		privateKey:      strings.TrimPrefix(privateKey, "0x"),
		chainID:         big.NewInt(chainID),
	}, nil
}

func (g *Gateway) SetValue(ctx context.Context, value uint64) (string, error) {
	var tx *types.Transaction
	err := Retry(ctx, 3, 300*time.Millisecond, func() error {
		priv, err := crypto.HexToECDSA(g.privateKey)
		if err != nil {
			return fmt.Errorf("decode private key: %w", err)
		}

		auth, err := bind.NewKeyedTransactorWithChainID(priv, g.chainID)
		if err != nil {
			return fmt.Errorf("new transactor: %w", err)
		}
		auth.Context = ctx

		contract := bind.NewBoundContract(g.contractAddress, g.contractABI, g.client, g.client, g.client)
		tx, err = contract.Transact(auth, "set", big.NewInt(0).SetUint64(value))
		if err != nil {
			return fmt.Errorf("transact set: %w", err)
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	return shared.NormalizeHash(tx.Hash().Hex()), nil
}

func (g *Gateway) GetValue(ctx context.Context) (uint64, error) {
	var output []interface{}
	caller := bind.CallOpts{Pending: false, Context: ctx}
	contract := bind.NewBoundContract(g.contractAddress, g.contractABI, g.client, g.client, g.client)
	if err := contract.Call(&caller, &output, "get"); err != nil {
		return 0, fmt.Errorf("call get: %w", err)
	}
	if len(output) == 0 {
		return 0, fmt.Errorf("contract get returned empty output")
	}
	v, ok := output[0].(*big.Int)
	if !ok {
		return 0, fmt.Errorf("unexpected contract get output type: %T", output[0])
	}
	return v.Uint64(), nil
}

func (g *Gateway) GetReceipt(ctx context.Context, txHash string) (transaction.Receipt, error) {
	receipt, err := g.client.TransactionReceipt(ctx, common.HexToHash(txHash))
	if err != nil {
		// Not mined yet
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			return transaction.Receipt{Exists: false}, nil
		}
		return transaction.Receipt{}, fmt.Errorf("transaction receipt: %w", err)
	}
	return transaction.Receipt{
		Exists:      true,
		Success:     receipt.Status == types.ReceiptStatusSuccessful,
		BlockNumber: receipt.BlockNumber.Uint64(),
	}, nil
}

func (g *Gateway) CurrentBlock(ctx context.Context) (uint64, error) {
	n, err := g.client.BlockNumber(ctx)
	if err != nil {
		return 0, fmt.Errorf("get current block: %w", err)
	}
	return n, nil
}
