package schedule

import (
	"context"

	"github.com/Grisha1Kadetov/slotbooking/internal/model/schedule"
	"github.com/google/uuid"
)

type service interface {
	Create(ctx context.Context, schedule schedule.Schedule) (schedule.Schedule, error)
	GetByRoomId(ctx context.Context, roomID uuid.UUID) (schedule.Schedule, error)
}
