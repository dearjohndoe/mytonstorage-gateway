package workers

import (
	"context"
	"log/slog"
	"time"

	"mytonstorage-gateway/pkg/workers/cleaner"
)

type workerFunc = func(ctx context.Context) (interval time.Duration, err error)

type worker struct {
	cleaner cleaner.Worker
	logger  *slog.Logger
}

type Workers interface {
	Start(ctx context.Context) (err error)
}

func (w *worker) Start(ctx context.Context) (err error) {
	go w.run(ctx, "CleanupOldData", w.cleaner.CleanupOldData)

	return nil
}

func (w *worker) run(ctx context.Context, name string, f workerFunc) {
	logger := w.logger.With(slog.String("run_worker", name))

	for {
		select {
		case <-ctx.Done():
			return
		default:
			interval, err := f(ctx)
			if err != nil {
				logger.Error(err.Error())
			}
			if interval <= 0 {
				interval = time.Second
			}
			t := time.NewTimer(interval)
			select {
			case <-ctx.Done():
				t.Stop()
				return
			case <-t.C:
			}
		}
	}
}

func NewWorkers(
	cleaner cleaner.Worker,
	logger *slog.Logger,
) Workers {
	return &worker{
		cleaner: cleaner,
		logger:  logger,
	}
}
