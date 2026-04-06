package bootstrap

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"backend/internal/application/command"
	"backend/internal/application/query"
	appsync "backend/internal/application/sync"
	"backend/internal/infrastructure/blockchain"
	"backend/internal/infrastructure/config"
	"backend/internal/infrastructure/metrics"
	"backend/internal/infrastructure/persistence/postgres"
	httpapi "backend/internal/interfaces/http"
	"backend/internal/interfaces/http/handlers"
	"backend/internal/interfaces/http/middleware"
	"backend/internal/jobs"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
)

type App struct {
	Handler http.Handler
	Worker  *jobs.SyncWorker
	Close   func(context.Context) error
}

func Wire(ctx context.Context, log *slog.Logger, cfg config.Config) (*App, error) {
	pool, err := postgres.NewPool(ctx, cfg.DatabaseURL, cfg.DBMaxOpenConns, cfg.DBMaxIdleConns)
	if err != nil {
		return nil, err
	}

	chainClient, err := blockchain.Dial(ctx, cfg.BesuRPCURL)
	if err != nil {
		pool.Close()
		return nil, err
	}

	gateway, err := blockchain.NewGateway(chainClient, cfg.ContractAddress, cfg.PrivateKey, cfg.ChainID)
	if err != nil {
		pool.Close()
		chainClient.Close()
		return nil, err
	}

	txRepo := postgres.NewTransactionRepository(pool)
	projRepo := postgres.NewValueProjectionRepository(pool)

	setSvc := command.NewSetValueService(gateway, txRepo, cfg.FromAddress, cfg.ContractAddress)
	getValueSvc := query.NewGetValueService(gateway, projRepo, cfg.ContractAddress)
	getTxSvc := query.NewGetTransactionStatusService(txRepo, gateway, uint64(cfg.ConfirmationsRequired))
	checkSvc := query.NewCheckConsistencyService(gateway, projRepo, cfg.ContractAddress)
	syncSvc := appsync.NewReconcilePendingService(
		txRepo,
		gateway,
		projRepo,
		uint64(cfg.ConfirmationsRequired),
		cfg.ContractAddress,
	)

	healthHandler := handlers.NewHealthHandler(func(r *http.Request) error {
		return readyCheck(r.Context(), pool, chainClient)
	})
	valueHandler := handlers.NewValueHandler(setSvc, getValueSvc, checkSvc, syncSvc)
	txHandler := handlers.NewTransactionHandler(getTxSvc)

	registry := prometheus.NewRegistry()
	appMetrics := metrics.New(registry)
	router := httpapi.NewRouter(healthHandler, valueHandler, txHandler, metrics.Handler(registry))
	chain := httpapi.Chain(
		router,
		middleware.RequestID,
		middleware.Recover(log),
		middleware.Timeout(cfg.RequestTimeout),
		middleware.Logging(log, appMetrics),
	)

	worker := jobs.NewSyncWorker(log, syncSvc, cfg.SyncInterval, cfg.SyncBatchSize)

	return &App{
		Handler: chain,
		Worker:  worker,
		Close: func(ctx context.Context) error {
			var errs []error
			done := make(chan struct{})
			go func() {
				pool.Close()
				chainClient.Close()
				close(done)
			}()

			select {
			case <-ctx.Done():
				errs = append(errs, ctx.Err())
			case <-done:
			}
			if len(errs) > 0 {
				return fmt.Errorf("close resources: %w", errs[0])
			}
			return nil
		},
	}, nil
}

func readyCheck(ctx context.Context, pool *pgxpool.Pool, chainClient *ethclient.Client) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("database not ready: %w", err)
	}
	if _, err := chainClient.BlockNumber(ctx); err != nil {
		return fmt.Errorf("besu not ready: %w", err)
	}
	return nil
}
