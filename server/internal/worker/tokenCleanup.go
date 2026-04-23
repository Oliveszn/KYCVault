package worker

import (
	"context"
	"kycvault/internal/repository"
	"time"

	"go.uber.org/zap"
)

type TokenCleanupWorker struct {
	repo     repository.AuthRepository
	logger   *zap.Logger
	interval time.Duration
}

func NewTokenCleanupWorker(
	repo repository.AuthRepository,
	logger *zap.Logger,
	interval time.Duration,
) *TokenCleanupWorker {
	return &TokenCleanupWorker{
		repo:     repo,
		logger:   logger,
		interval: interval,
	}
}

func (w *TokenCleanupWorker) Start(ctx context.Context) {
	w.logger.Info("token cleanup worker started", zap.Duration("interval", w.interval))
	ticker := time.NewTicker(w.interval)

	go func() {
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				w.logger.Info("token cleanup worker stopped")
				return

			case <-ticker.C:
				w.run(ctx)
			}
		}
	}()
}

func (w *TokenCleanupWorker) run(ctx context.Context) {
	w.logger.Info("running token cleanup job")

	deleted, err := w.repo.DeleteExpiredTokens(ctx, 100)
	w.logger.Info("deleted expired tokens", zap.Int64("count", deleted))
	if err != nil {
		w.logger.Error("failed to delete expired tokens", zap.Error(err))
		return
	}

	w.logger.Info("expired tokens cleaned", zap.Int64("deleted", deleted))
}
