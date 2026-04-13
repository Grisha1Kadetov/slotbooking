package slot_test

import (
	"context"
	"errors"
	"testing"
	"time"

	schedulemodel "github.com/Grisha1Kadetov/slotbooking/internal/model/schedule"
	slotmodel "github.com/Grisha1Kadetov/slotbooking/internal/model/slot"
	slotservice "github.com/Grisha1Kadetov/slotbooking/internal/service/slot"
	"github.com/google/uuid"
)

type slotRepositoryMock struct {
	createBulkFunc                  func(ctx context.Context, slots []slotmodel.Slot) ([]slotmodel.Slot, error)
	getByIdFunc                     func(ctx context.Context, slotId uuid.UUID) (slotmodel.Slot, error)
	getAvailableByRoomIdAndDateFunc func(ctx context.Context, roomID uuid.UUID, date time.Time) ([]slotmodel.Slot, error)
	hasAnyByRoomIdAndDateFunc       func(ctx context.Context, roomID uuid.UUID, date time.Time) (bool, error)
}

func (m *slotRepositoryMock) CreateBulk(ctx context.Context, slots []slotmodel.Slot) ([]slotmodel.Slot, error) {
	if m.createBulkFunc == nil {
		return nil, errors.New("unexpected CreateBulk call")
	}
	return m.createBulkFunc(ctx, slots)
}

func (m *slotRepositoryMock) GetById(ctx context.Context, slotId uuid.UUID) (slotmodel.Slot, error) {
	if m.getByIdFunc == nil {
		return slotmodel.Slot{}, errors.New("unexpected GetById call")
	}
	return m.getByIdFunc(ctx, slotId)
}

func (m *slotRepositoryMock) GetAvailableByRoomIdAndDate(ctx context.Context, roomID uuid.UUID, date time.Time) ([]slotmodel.Slot, error) {
	if m.getAvailableByRoomIdAndDateFunc == nil {
		return nil, errors.New("unexpected GetAvailableByRoomIdAndDate call")
	}
	return m.getAvailableByRoomIdAndDateFunc(ctx, roomID, date)
}

func (m *slotRepositoryMock) HasAnyByRoomIdAndDate(ctx context.Context, roomID uuid.UUID, date time.Time) (bool, error) {
	if m.hasAnyByRoomIdAndDateFunc == nil {
		return false, errors.New("unexpected HasAnyByRoomIdAndDate call")
	}
	return m.hasAnyByRoomIdAndDateFunc(ctx, roomID, date)
}

type scheduleProviderMock struct {
	getByRoomIdFunc func(ctx context.Context, roomID uuid.UUID) (schedulemodel.Schedule, error)
}

func (m *scheduleProviderMock) GetByRoomId(ctx context.Context, roomID uuid.UUID) (schedulemodel.Schedule, error) {
	if m.getByRoomIdFunc == nil {
		return schedulemodel.Schedule{}, errors.New("unexpected GetByRoomId call")
	}
	return m.getByRoomIdFunc(ctx, roomID)
}

type roomProviderMock struct {
	existsByIdFunc func(ctx context.Context, roomID uuid.UUID) (bool, error)
}

func (m *roomProviderMock) ExistsById(ctx context.Context, roomID uuid.UUID) (bool, error) {
	if m.existsByIdFunc == nil {
		return false, errors.New("unexpected ExistsById call")
	}
	return m.existsByIdFunc(ctx, roomID)
}

func TestSlotService_GetById(t *testing.T) {
	ctx := context.Background()
	slotID := uuid.New()
	expected := slotmodel.Slot{
		ID:      slotID,
		RoomID:  uuid.New(),
		StartAt: time.Now().UTC(),
		EndAt:   time.Now().UTC().Add(30 * time.Minute),
	}

	repo := &slotRepositoryMock{
		getByIdFunc: func(ctx context.Context, gotSlotID uuid.UUID) (slotmodel.Slot, error) {
			if gotSlotID != slotID {
				t.Fatalf("GetById() slotId = %v, want %v", gotSlotID, slotID)
			}
			return expected, nil
		},
	}

	service := slotservice.New(repo, &scheduleProviderMock{}, &roomProviderMock{}, 30*time.Minute, 7)

	got, err := service.GetById(ctx, slotID)
	if err != nil {
		t.Fatalf("GetById() error = %v", err)
	}
	if got.ID != expected.ID {
		t.Fatalf("GetById() ID = %v, want %v", got.ID, expected.ID)
	}
}

func TestSlotService_GenerateDaySlotsByRoomId(t *testing.T) {
	ctx := context.Background()
	roomID := uuid.New()
	date := time.Date(2026, 3, 30, 0, 0, 0, 0, time.UTC) // monday
	scheduleData := schedulemodel.Schedule{
		RoomID:     roomID,
		DaysOfWeek: []int{1},
		StartTime:  time.Date(0, 1, 1, 9, 0, 0, 0, time.UTC),
		EndTime:    time.Date(0, 1, 1, 10, 0, 0, 0, time.UTC),
	}

	t.Run("schedule not found returns empty slice", func(t *testing.T) {
		repo := &slotRepositoryMock{}
		scheduleProvider := &scheduleProviderMock{
			getByRoomIdFunc: func(ctx context.Context, roomID uuid.UUID) (schedulemodel.Schedule, error) {
				return schedulemodel.Schedule{}, schedulemodel.ErrNotFound
			},
		}

		service := slotservice.New(repo, scheduleProvider, &roomProviderMock{}, 30*time.Minute, 7)

		got, err := service.GenerateDaySlotsByRoomId(ctx, date, roomID)
		if err != nil {
			t.Fatalf("GenerateDaySlotsByRoomId() error = %v", err)
		}
		if len(got) != 0 {
			t.Fatalf("GenerateDaySlotsByRoomId() len = %d, want 0", len(got))
		}
	})

	t.Run("success", func(t *testing.T) {
		repo := &slotRepositoryMock{
			createBulkFunc: func(ctx context.Context, slots []slotmodel.Slot) ([]slotmodel.Slot, error) {
				if len(slots) != 2 {
					t.Fatalf("CreateBulk() len = %d, want 2", len(slots))
				}
				if slots[0].RoomID != roomID || slots[1].RoomID != roomID {
					t.Fatal("CreateBulk() received slots with wrong room ID")
				}
				return slots, nil
			},
		}
		scheduleProvider := &scheduleProviderMock{
			getByRoomIdFunc: func(ctx context.Context, gotRoomID uuid.UUID) (schedulemodel.Schedule, error) {
				if gotRoomID != roomID {
					t.Fatalf("GetByRoomId() roomID = %v, want %v", gotRoomID, roomID)
				}
				return scheduleData, nil
			},
		}

		service := slotservice.New(repo, scheduleProvider, &roomProviderMock{}, 30*time.Minute, 7)

		got, err := service.GenerateDaySlotsByRoomId(ctx, date, roomID)
		if err != nil {
			t.Fatalf("GenerateDaySlotsByRoomId() error = %v", err)
		}
		if len(got) != 2 {
			t.Fatalf("GenerateDaySlotsByRoomId() len = %d, want 2", len(got))
		}
		if got[0].StartAt.Hour() != 9 || got[0].StartAt.Minute() != 0 {
			t.Fatalf("first slot starts at %v, want 09:00", got[0].StartAt)
		}
		if got[1].StartAt.Hour() != 9 || got[1].StartAt.Minute() != 30 {
			t.Fatalf("second slot starts at %v, want 09:30", got[1].StartAt)
		}
	})
}

func TestSlotService_GenerateDaySlotsBySchedule(t *testing.T) {
	ctx := context.Background()
	roomID := uuid.New()
	scheduleData := schedulemodel.Schedule{
		RoomID:     roomID,
		DaysOfWeek: []int{1},
		StartTime:  time.Date(0, 1, 1, 9, 0, 0, 0, time.UTC),
		EndTime:    time.Date(0, 1, 1, 10, 0, 0, 0, time.UTC),
	}

	t.Run("weekday not included", func(t *testing.T) {
		date := time.Date(2026, 3, 31, 0, 0, 0, 0, time.UTC) // tuesday
		repo := &slotRepositoryMock{}
		service := slotservice.New(repo, &scheduleProviderMock{}, &roomProviderMock{}, 30*time.Minute, 7)

		_, err := service.GenerateDaySlotsBySchedule(ctx, date, scheduleData)
		if err == nil {
			t.Fatal("GenerateDaySlotsBySchedule() expected error, got nil")
		}
	})

	t.Run("success", func(t *testing.T) {
		date := time.Date(2026, 3, 30, 0, 0, 0, 0, time.UTC) // monday
		repo := &slotRepositoryMock{
			createBulkFunc: func(ctx context.Context, slots []slotmodel.Slot) ([]slotmodel.Slot, error) {
				return slots, nil
			},
		}
		service := slotservice.New(repo, &scheduleProviderMock{}, &roomProviderMock{}, 30*time.Minute, 7)

		got, err := service.GenerateDaySlotsBySchedule(ctx, date, scheduleData)
		if err != nil {
			t.Fatalf("GenerateDaySlotsBySchedule() error = %v", err)
		}
		if len(got) != 2 {
			t.Fatalf("GenerateDaySlotsBySchedule() len = %d, want 2", len(got))
		}
		if got[0].EndAt.Sub(got[0].StartAt) != 30*time.Minute {
			t.Fatalf("slot duration = %v, want %v", got[0].EndAt.Sub(got[0].StartAt), 30*time.Minute)
		}
	})
}

func TestSlotService_PreGenerateSlotsByRoomId(t *testing.T) {
	ctx := context.Background()
	roomID := uuid.New()
	date := time.Date(2026, 3, 30, 0, 0, 0, 0, time.UTC) // monday
	scheduleData := schedulemodel.Schedule{
		RoomID:     roomID,
		DaysOfWeek: []int{1},
		StartTime:  time.Date(0, 1, 1, 9, 0, 0, 0, time.UTC),
		EndTime:    time.Date(0, 1, 1, 10, 0, 0, 0, time.UTC),
	}

	t.Run("ignores weekday not included and overlap", func(t *testing.T) {
		calls := 0
		repo := &slotRepositoryMock{
			createBulkFunc: func(ctx context.Context, slots []slotmodel.Slot) ([]slotmodel.Slot, error) {
				calls++
				if calls == 1 {
					return nil, slotmodel.ErrTimeOverlap
				}
				return slots, nil
			},
		}
		scheduleProvider := &scheduleProviderMock{
			getByRoomIdFunc: func(ctx context.Context, roomID uuid.UUID) (schedulemodel.Schedule, error) {
				return scheduleData, nil
			},
		}
		service := slotservice.New(repo, scheduleProvider, &roomProviderMock{}, 30*time.Minute, 1)

		err := service.PreGenerateSlotsByRoomId(ctx, date, roomID)
		if err != nil {
			t.Fatalf("PreGenerateSlotsByRoomId() error = %v", err)
		}
	})

	t.Run("returns unexpected error", func(t *testing.T) {
		expectedErr := errors.New("db error")
		repo := &slotRepositoryMock{
			createBulkFunc: func(ctx context.Context, slots []slotmodel.Slot) ([]slotmodel.Slot, error) {
				return nil, expectedErr
			},
		}
		scheduleProvider := &scheduleProviderMock{
			getByRoomIdFunc: func(ctx context.Context, roomID uuid.UUID) (schedulemodel.Schedule, error) {
				return scheduleData, nil
			},
		}
		service := slotservice.New(repo, scheduleProvider, &roomProviderMock{}, 30*time.Minute, 0)

		err := service.PreGenerateSlotsByRoomId(ctx, date, roomID)
		if !errors.Is(err, expectedErr) {
			t.Fatalf("PreGenerateSlotsByRoomId() error = %v, want %v", err, expectedErr)
		}
	})
}

func TestSlotService_GetAvailableByRoomIdAndDate(t *testing.T) {
	ctx := context.Background()
	roomID := uuid.New()
	date := time.Date(2026, 3, 30, 13, 45, 0, 0, time.UTC)
	scheduleData := schedulemodel.Schedule{
		RoomID:     roomID,
		DaysOfWeek: []int{1},
		StartTime:  time.Date(0, 1, 1, 9, 0, 0, 0, time.UTC),
		EndTime:    time.Date(0, 1, 1, 10, 0, 0, 0, time.UTC),
	}

	t.Run("room not found", func(t *testing.T) {
		repo := &slotRepositoryMock{}
		scheduleProvider := &scheduleProviderMock{}
		roomProvider := &roomProviderMock{
			existsByIdFunc: func(ctx context.Context, roomID uuid.UUID) (bool, error) {
				return false, nil
			},
		}
		service := slotservice.New(repo, scheduleProvider, roomProvider, 30*time.Minute, 7)

		_, err := service.GetAvailableByRoomIdAndDate(ctx, roomID, date)
		if !errors.Is(err, slotmodel.ErrRoomNotFound) {
			t.Fatalf("GetAvailableByRoomIdAndDate() error = %v, want %v", err, slotmodel.ErrRoomNotFound)
		}
	})

	t.Run("returns existing available slots", func(t *testing.T) {
		available := []slotmodel.Slot{{ID: uuid.New(), RoomID: roomID}}
		repo := &slotRepositoryMock{
			hasAnyByRoomIdAndDateFunc: func(ctx context.Context, gotRoomID uuid.UUID, gotDate time.Time) (bool, error) {
				if gotRoomID != roomID {
					t.Fatalf("HasAnyByRoomIdAndDate() roomID = %v, want %v", gotRoomID, roomID)
				}
				if gotDate.Hour() != 0 || gotDate.Minute() != 0 {
					t.Fatalf("HasAnyByRoomIdAndDate() date = %v, want date at 00:00", gotDate)
				}
				return true, nil
			},
			getAvailableByRoomIdAndDateFunc: func(ctx context.Context, roomID uuid.UUID, date time.Time) ([]slotmodel.Slot, error) {
				return available, nil
			},
		}
		roomProvider := &roomProviderMock{
			existsByIdFunc: func(ctx context.Context, roomID uuid.UUID) (bool, error) {
				return true, nil
			},
		}
		service := slotservice.New(repo, &scheduleProviderMock{}, roomProvider, 30*time.Minute, 7)

		got, err := service.GetAvailableByRoomIdAndDate(ctx, roomID, date)
		if err != nil {
			t.Fatalf("GetAvailableByRoomIdAndDate() error = %v", err)
		}
		if len(got) != 1 {
			t.Fatalf("GetAvailableByRoomIdAndDate() len = %d, want 1", len(got))
		}
	})

	t.Run("generates slots when none exist", func(t *testing.T) {
		repo := &slotRepositoryMock{
			hasAnyByRoomIdAndDateFunc: func(ctx context.Context, roomID uuid.UUID, date time.Time) (bool, error) {
				return false, nil
			},
			createBulkFunc: func(ctx context.Context, slots []slotmodel.Slot) ([]slotmodel.Slot, error) {
				return slots, nil
			},
		}
		scheduleProvider := &scheduleProviderMock{
			getByRoomIdFunc: func(ctx context.Context, roomID uuid.UUID) (schedulemodel.Schedule, error) {
				return scheduleData, nil
			},
		}
		roomProvider := &roomProviderMock{
			existsByIdFunc: func(ctx context.Context, roomID uuid.UUID) (bool, error) {
				return true, nil
			},
		}
		service := slotservice.New(repo, scheduleProvider, roomProvider, 30*time.Minute, 7)

		got, err := service.GetAvailableByRoomIdAndDate(ctx, roomID, date)
		if err != nil {
			t.Fatalf("GetAvailableByRoomIdAndDate() error = %v", err)
		}
		if len(got) != 2 {
			t.Fatalf("GetAvailableByRoomIdAndDate() len = %d, want 2", len(got))
		}
	})

	t.Run("returns empty slice when generated day is not in schedule", func(t *testing.T) {
		tuesdayDate := time.Date(2026, 3, 31, 12, 0, 0, 0, time.UTC)
		repo := &slotRepositoryMock{
			hasAnyByRoomIdAndDateFunc: func(ctx context.Context, roomID uuid.UUID, date time.Time) (bool, error) {
				return false, nil
			},
		}
		scheduleProvider := &scheduleProviderMock{
			getByRoomIdFunc: func(ctx context.Context, roomID uuid.UUID) (schedulemodel.Schedule, error) {
				return scheduleData, nil
			},
		}
		roomProvider := &roomProviderMock{
			existsByIdFunc: func(ctx context.Context, roomID uuid.UUID) (bool, error) {
				return true, nil
			},
		}
		service := slotservice.New(repo, scheduleProvider, roomProvider, 30*time.Minute, 7)

		got, err := service.GetAvailableByRoomIdAndDate(ctx, roomID, tuesdayDate)
		if err != nil {
			t.Fatalf("GetAvailableByRoomIdAndDate() error = %v", err)
		}
		if len(got) != 0 {
			t.Fatalf("GetAvailableByRoomIdAndDate() len = %d, want 0", len(got))
		}
	})
}
