package jobs

import (
	"context"
	"log/slog"
	"time"

	"backend/internal/application/sync"
)

type SyncWorker struct {
	log      *slog.Logger
	service  *sync.ReconcilePendingService
	interval time.Duration
	batch    int
}

func NewSyncWorker(log *slog.Logger, service *sync.ReconcilePendingService, interval time.Duration, batch int) *SyncWorker {
	return &SyncWorker{
		log:      log,
		service:  service,
		interval: interval,
		batch:    batch,
	}
}

func (w *SyncWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			runCtx, cancel := context.WithTimeout(ctx, w.interval)
			res, err := w.service.Execute(runCtx, w.batch)
			cancel()
			if err != nil {
				w.log.Error("sync reconcile failed", "err", err)
				continue
			}
			if res.Processed > 0 {
				w.log.Info("sync reconcile tick", "processed", res.Processed, "updated", res.Updated)
			}
		}
	}
}
