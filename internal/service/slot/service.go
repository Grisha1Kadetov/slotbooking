package slot

import (
	"context"
	"errors"
	"time"

	"github.com/Grisha1Kadetov/slotbooking/internal/model/schedule"
	"github.com/Grisha1Kadetov/slotbooking/internal/model/slot"
	"github.com/google/uuid"
)

var errWeekdayNotInclude = errors.New("this weekday not include")

type SlotService struct {
	slotRepo         slotRepository
	scheduleProvider scheduleProvider
	roomProvider     roomProvider
	slotDur          time.Duration
	pregenDays       int
}

func New(slotRepo slotRepository, scheduleProvider scheduleProvider, roomProvider roomProvider, slotDur time.Duration, pregenDays int) *SlotService {
	return &SlotService{
		slotRepo:         slotRepo,
		scheduleProvider: scheduleProvider,
		roomProvider:     roomProvider,
		slotDur:          slotDur,
		pregenDays:       pregenDays,
	}
}

func (s *SlotService) GetById(ctx context.Context, slotId uuid.UUID) (slot.Slot, error) {
	return s.slotRepo.GetById(ctx, slotId)
}

func (s *SlotService) GenerateDaySlotsByRoomId(ctx context.Context, date time.Time, roomId uuid.UUID) ([]slot.Slot, error) {
	sch, err := s.scheduleProvider.GetByRoomId(ctx, roomId)
	if err != nil {
		if errors.Is(err, schedule.ErrNotFound) {
			return []slot.Slot{}, nil
		}
		return nil, err
	}
	return s.GenerateDaySlotsBySchedule(ctx, date, sch)
}

func (s *SlotService) GenerateDaySlotsBySchedule(ctx context.Context, date time.Time, schedule schedule.Schedule) ([]slot.Slot, error) {
	if !contains(date.Weekday(), schedule.DaysOfWeek) {
		return nil, errWeekdayNotInclude
	}
	newSlots := []slot.Slot{}

	endT := dateWithTime(date.UTC(), schedule.EndTime.Hour(), schedule.EndTime.Minute())
	date = dateWithTime(date.UTC(), schedule.StartTime.Hour(), schedule.StartTime.Minute())

	for date.Before(endT) {
		endSlotT := date.Add(s.slotDur)
		if endSlotT.Before(endT) || endSlotT.Equal(endT) {
			s := slot.Slot{
				RoomID:  schedule.RoomID,
				StartAt: date,
				EndAt:   endSlotT,
			}
			newSlots = append(newSlots, s)
		}
		date = endSlotT
	}

	return s.slotRepo.CreateBulk(ctx, newSlots)
}

func (s *SlotService) PreGenerateSlotsByRoomId(ctx context.Context, dateFromInclude time.Time, roomId uuid.UUID) error {
	return s.PreGenerateSlotsByRoomIdWithDuration(ctx, dateFromInclude, s.pregenDays, roomId)
}

func (s *SlotService) PreGenerateSlotsByRoomIdWithDuration(ctx context.Context, dateFromInclude time.Time, nextDays int, roomId uuid.UUID) error {
	_, err := s.GenerateDaySlotsByRoomId(ctx, dateFromInclude, roomId)
	if err != nil {
		if !errors.Is(err, errWeekdayNotInclude) && !errors.Is(err, slot.ErrTimeOverlap) {
			return err
		}
	}

	for nextDays != 0 {
		dateFromInclude = dateFromInclude.AddDate(0, 0, 1)
		_, err := s.GenerateDaySlotsByRoomId(ctx, dateFromInclude, roomId)
		if err != nil {
			if !errors.Is(err, errWeekdayNotInclude) && !errors.Is(err, slot.ErrTimeOverlap) {
				return err
			}
		}
		nextDays--
	}
	return nil
}

func (s *SlotService) GetAvailableByRoomIdAndDate(ctx context.Context, roomID uuid.UUID, date time.Time) ([]slot.Slot, error) {
	hasRoom, err := s.roomProvider.ExistsById(ctx, roomID)
	if err != nil {
		return nil, err
	}
	if !hasRoom {
		return nil, slot.ErrRoomNotFound
	}

	date = dateWithTime(date, 0, 0).UTC()

	hasSlots, err := s.slotRepo.HasAnyByRoomIdAndDate(ctx, roomID, date)
	if err != nil {
		return nil, err
	}

	if hasSlots {
		slots, err := s.slotRepo.GetAvailableByRoomIdAndDate(ctx, roomID, date)
		if err != nil {
			return nil, err
		}
		return slots, nil
	}

	slots, err := s.GenerateDaySlotsByRoomId(ctx, date, roomID)
	if err != nil {
		if errors.Is(err, errWeekdayNotInclude) {
			return []slot.Slot{}, nil
		}
		if errors.Is(err, slot.ErrTimeOverlap) {
			return s.slotRepo.GetAvailableByRoomIdAndDate(ctx, roomID, date)
		}
		return nil, err
	}
	return slots, nil
}

func equalWeekday(w time.Weekday, i int) bool {
	return (int(w)+6)%7+1 == i
}

func contains(w time.Weekday, ws []int) bool {
	for _, i := range ws {
		if equalWeekday(w, i) {
			return true
		}
	}
	return false
}

func dateWithTime(date time.Time, h, m int) time.Time {
	return time.Date(date.Year(), date.Month(), date.Day(), h, m, 0, 0, date.Location())
}
