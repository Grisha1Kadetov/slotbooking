package schedule

import (
	"context"
	"time"

	"github.com/Grisha1Kadetov/slotbooking/internal/model/schedule"
	"github.com/google/uuid"
)

type ScheduleService struct {
	scheduleRepo  repository
	pregenService pregenerateSlotService
}

func New(scheduleRepo repository, pregenService pregenerateSlotService) *ScheduleService {
	return &ScheduleService{
		scheduleRepo:  scheduleRepo,
		pregenService: pregenService,
	}
}

func (ss *ScheduleService) Create(ctx context.Context, s schedule.Schedule) (schedule.Schedule, error) {
	createdSchedule, err := ss.scheduleRepo.Create(ctx, s)
	if err != nil {
		return schedule.Schedule{}, err
	}
	err = ss.pregenService.PreGenerateSlotsByRoomId(ctx, time.Now(), createdSchedule.RoomID)
	if err != nil {
		return createdSchedule, err
	}
	return createdSchedule, nil
}

func (ss *ScheduleService) GetByRoomId(ctx context.Context, roomID uuid.UUID) (schedule.Schedule, error) {
	return ss.scheduleRepo.GetByRoomId(ctx, roomID)
}
