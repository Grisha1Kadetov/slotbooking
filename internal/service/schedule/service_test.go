package schedule_test

import (
	"context"
	"errors"
	"testing"
	"time"

	schedulemodel "github.com/Grisha1Kadetov/slotbooking/internal/model/schedule"
	scheduleservice "github.com/Grisha1Kadetov/slotbooking/internal/service/schedule"
	"github.com/google/uuid"
)

type repositoryMock struct {
	createFunc      func(ctx context.Context, s schedulemodel.Schedule) (schedulemodel.Schedule, error)
	getByRoomIdFunc func(ctx context.Context, roomID uuid.UUID) (schedulemodel.Schedule, error)
}

func (m *repositoryMock) Create(ctx context.Context, s schedulemodel.Schedule) (schedulemodel.Schedule, error) {
	if m.createFunc == nil {
		return schedulemodel.Schedule{}, errors.New("unexpected Create call")
	}
	return m.createFunc(ctx, s)
}

func (m *repositoryMock) GetByRoomId(ctx context.Context, roomID uuid.UUID) (schedulemodel.Schedule, error) {
	if m.getByRoomIdFunc == nil {
		return schedulemodel.Schedule{}, errors.New("unexpected GetByRoomId call")
	}
	return m.getByRoomIdFunc(ctx, roomID)
}

type pregenerateSlotServiceMock struct {
	preGenerateSlotsByRoomIdFunc func(ctx context.Context, dateFromInclude time.Time, roomId uuid.UUID) error
}

func (m *pregenerateSlotServiceMock) PreGenerateSlotsByRoomId(ctx context.Context, dateFromInclude time.Time, roomId uuid.UUID) error {
	if m.preGenerateSlotsByRoomIdFunc == nil {
		return errors.New("unexpected PreGenerateSlotsByRoomId call")
	}
	return m.preGenerateSlotsByRoomIdFunc(ctx, dateFromInclude, roomId)
}

func TestScheduleService_Create(t *testing.T) {
	ctx := context.Background()
	scheduleID := uuid.New()
	roomID := uuid.New()
	input := schedulemodel.Schedule{
		RoomID:     roomID,
		DaysOfWeek: []int{1, 3, 5},
		StartTime:  time.Date(0, 1, 1, 9, 0, 0, 0, time.UTC),
		EndTime:    time.Date(0, 1, 1, 18, 0, 0, 0, time.UTC),
	}

	t.Run("success", func(t *testing.T) {
		var pregeneratedRoomID uuid.UUID

		repo := &repositoryMock{
			createFunc: func(ctx context.Context, s schedulemodel.Schedule) (schedulemodel.Schedule, error) {
				if s.RoomID != input.RoomID {
					t.Fatalf("Create() RoomID = %v, want %v", s.RoomID, input.RoomID)
				}
				if len(s.DaysOfWeek) != len(input.DaysOfWeek) {
					t.Fatalf("Create() DaysOfWeek len = %d, want %d", len(s.DaysOfWeek), len(input.DaysOfWeek))
				}
				s.ID = scheduleID
				return s, nil
			},
		}
		pregen := &pregenerateSlotServiceMock{
			preGenerateSlotsByRoomIdFunc: func(ctx context.Context, dateFromInclude time.Time, roomId uuid.UUID) error {
				pregeneratedRoomID = roomId
				if roomId != roomID {
					t.Fatalf("PreGenerateSlotsByRoomId() roomId = %v, want %v", roomId, roomID)
				}
				return nil
			},
		}

		service := scheduleservice.New(repo, pregen)

		got, err := service.Create(ctx, input)
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		if got.ID != scheduleID {
			t.Fatalf("Create() ID = %v, want %v", got.ID, scheduleID)
		}
		if pregeneratedRoomID != roomID {
			t.Fatalf("PreGenerateSlotsByRoomId() was called with roomId = %v, want %v", pregeneratedRoomID, roomID)
		}
	})

	t.Run("repo error", func(t *testing.T) {
		expectedErr := errors.New("repo create error")
		repo := &repositoryMock{
			createFunc: func(ctx context.Context, s schedulemodel.Schedule) (schedulemodel.Schedule, error) {
				return schedulemodel.Schedule{}, expectedErr
			},
		}
		pregen := &pregenerateSlotServiceMock{}

		service := scheduleservice.New(repo, pregen)

		_, err := service.Create(ctx, input)
		if !errors.Is(err, expectedErr) {
			t.Fatalf("Create() error = %v, want %v", err, expectedErr)
		}
	})

	t.Run("pregenerate error returns created schedule", func(t *testing.T) {
		expectedErr := errors.New("pregenerate error")
		repo := &repositoryMock{
			createFunc: func(ctx context.Context, s schedulemodel.Schedule) (schedulemodel.Schedule, error) {
				s.ID = scheduleID
				return s, nil
			},
		}
		pregen := &pregenerateSlotServiceMock{
			preGenerateSlotsByRoomIdFunc: func(ctx context.Context, dateFromInclude time.Time, roomId uuid.UUID) error {
				return expectedErr
			},
		}

		service := scheduleservice.New(repo, pregen)

		got, err := service.Create(ctx, input)
		if !errors.Is(err, expectedErr) {
			t.Fatalf("Create() error = %v, want %v", err, expectedErr)
		}
		if got.ID != scheduleID {
			t.Fatalf("Create() returned schedule ID = %v, want %v", got.ID, scheduleID)
		}
		if got.RoomID != roomID {
			t.Fatalf("Create() returned RoomID = %v, want %v", got.RoomID, roomID)
		}
	})
}

func TestScheduleService_GetByRoomId(t *testing.T) {
	ctx := context.Background()
	roomID := uuid.New()
	expected := schedulemodel.Schedule{
		ID:         uuid.New(),
		RoomID:     roomID,
		DaysOfWeek: []int{1, 2, 3},
		StartTime:  time.Date(0, 1, 1, 10, 0, 0, 0, time.UTC),
		EndTime:    time.Date(0, 1, 1, 19, 0, 0, 0, time.UTC),
	}

	repo := &repositoryMock{
		getByRoomIdFunc: func(ctx context.Context, gotRoomID uuid.UUID) (schedulemodel.Schedule, error) {
			if gotRoomID != roomID {
				t.Fatalf("GetByRoomId() roomID = %v, want %v", gotRoomID, roomID)
			}
			return expected, nil
		},
	}
	pregen := &pregenerateSlotServiceMock{}

	service := scheduleservice.New(repo, pregen)

	got, err := service.GetByRoomId(ctx, roomID)
	if err != nil {
		t.Fatalf("GetByRoomId() error = %v", err)
	}

	if got.ID != expected.ID {
		t.Fatalf("GetByRoomId() ID = %v, want %v", got.ID, expected.ID)
	}
	if got.RoomID != expected.RoomID {
		t.Fatalf("GetByRoomId() RoomID = %v, want %v", got.RoomID, expected.RoomID)
	}
	if len(got.DaysOfWeek) != len(expected.DaysOfWeek) {
		t.Fatalf("GetByRoomId() DaysOfWeek len = %d, want %d", len(got.DaysOfWeek), len(expected.DaysOfWeek))
	}
}
