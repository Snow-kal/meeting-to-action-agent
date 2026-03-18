package syncer

import (
	"context"

	"github.com/Snow-kal/meeting-to-action-agent/internal/domain"
)

type TaskSyncer interface {
	SyncTasks(ctx context.Context, tasks []domain.Task) ([]domain.SyncResult, error)
}
