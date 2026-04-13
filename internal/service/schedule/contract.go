package schedule

import (
	"context"
	"time"

	"github.com/Grisha1Kadetov/slotbooking/internal/model/schedule"
	"github.com/google/uuid"
)

type repository interface {
	Create(ctx context.Context, s schedule.Schedule) (schedule.Schedule, error)
	GetByRoomId(ctx context.Context, roomID uuid.UUID) (schedule.Schedule, error)
}

type pregenerateSlotService interface {
	PreGenerateSlotsByRoomId(ctx context.Context, dateFromInclude time.Time, roomId uuid.UUID) error
}
