package room_test

import (
	"context"
	"errors"
	"testing"
	"time"

	roommodel "github.com/Grisha1Kadetov/slotbooking/internal/model/room"
	roomservice "github.com/Grisha1Kadetov/slotbooking/internal/service/room"
	"github.com/google/uuid"
)

type roomRepositoryMock struct {
	createFunc func(ctx context.Context, r roommodel.Room) (roommodel.Room, error)
	getAllFunc func(ctx context.Context) ([]roommodel.Room, error)
}

func (m *roomRepositoryMock) Create(ctx context.Context, r roommodel.Room) (roommodel.Room, error) {
	if m.createFunc == nil {
		return roommodel.Room{}, errors.New("unexpected Create call")
	}
	return m.createFunc(ctx, r)
}

func (m *roomRepositoryMock) GetAll(ctx context.Context) ([]roommodel.Room, error) {
	if m.getAllFunc == nil {
		return nil, errors.New("unexpected GetAll call")
	}
	return m.getAllFunc(ctx)
}

func TestRoomService_CreateRoom(t *testing.T) {
	ctx := context.Background()
	expectedID := uuid.New()
	description := "Small meeting room"
	capacity := 6
	input := roommodel.Room{
		Name:        "Blue room",
		Description: &description,
		Capacity:    &capacity,
	}

	repo := &roomRepositoryMock{
		createFunc: func(ctx context.Context, r roommodel.Room) (roommodel.Room, error) {
			if r.Name != input.Name {
				t.Fatalf("Create() Name = %q, want %q", r.Name, input.Name)
			}
			if r.Description == nil || input.Description == nil || *r.Description != *input.Description {
				t.Fatalf("Create() Description = %v, want %v", valueOrNilString(r.Description), valueOrNilString(input.Description))
			}
			if r.Capacity == nil || input.Capacity == nil || *r.Capacity != *input.Capacity {
				t.Fatalf("Create() Capacity = %v, want %v", valueOrNilInt(r.Capacity), valueOrNilInt(input.Capacity))
			}

			r.ID = expectedID
			return r, nil
		},
	}

	service := roomservice.New(repo)

	got, err := service.CreateRoom(ctx, input)
	if err != nil {
		t.Fatalf("CreateRoom() error = %v", err)
	}

	if got.ID != expectedID {
		t.Fatalf("CreateRoom() ID = %v, want %v", got.ID, expectedID)
	}
	if got.Name != input.Name {
		t.Fatalf("CreateRoom() Name = %q, want %q", got.Name, input.Name)
	}
	if got.Description == nil || input.Description == nil || *got.Description != *input.Description {
		t.Fatalf("CreateRoom() Description = %v, want %v", valueOrNilString(got.Description), valueOrNilString(input.Description))
	}
	if got.Capacity == nil || input.Capacity == nil || *got.Capacity != *input.Capacity {
		t.Fatalf("CreateRoom() Capacity = %v, want %v", valueOrNilInt(got.Capacity), valueOrNilInt(input.Capacity))
	}
}

func TestRoomService_GetAllRooms(t *testing.T) {
	ctx := context.Background()
	description1 := "Small meeting room"
	capacity1 := 6
	createdAt1 := time.Now()
	description2 := "Large meeting room"
	capacity2 := 12
	createdAt2 := time.Now().Add(time.Minute)

	expected := []roommodel.Room{
		{
			ID:          uuid.New(),
			Name:        "Blue room",
			Description: &description1,
			Capacity:    &capacity1,
			CreatedAt:   &createdAt1,
		},
		{
			ID:          uuid.New(),
			Name:        "Green room",
			Description: &description2,
			Capacity:    &capacity2,
			CreatedAt:   &createdAt2,
		},
	}

	repo := &roomRepositoryMock{
		getAllFunc: func(ctx context.Context) ([]roommodel.Room, error) {
			return expected, nil
		},
	}

	service := roomservice.New(repo)

	got, err := service.GetAllRooms(ctx)
	if err != nil {
		t.Fatalf("GetAllRooms() error = %v", err)
	}

	if len(got) != len(expected) {
		t.Fatalf("GetAllRooms() len = %d, want %d", len(got), len(expected))
	}

	for i := range expected {
		if got[i].ID != expected[i].ID {
			t.Fatalf("GetAllRooms()[%d].ID = %v, want %v", i, got[i].ID, expected[i].ID)
		}
		if got[i].Name != expected[i].Name {
			t.Fatalf("GetAllRooms()[%d].Name = %q, want %q", i, got[i].Name, expected[i].Name)
		}
		if valueOrNilString(got[i].Description) != valueOrNilString(expected[i].Description) {
			t.Fatalf("GetAllRooms()[%d].Description = %v, want %v", i, valueOrNilString(got[i].Description), valueOrNilString(expected[i].Description))
		}
		if valueOrNilInt(got[i].Capacity) != valueOrNilInt(expected[i].Capacity) {
			t.Fatalf("GetAllRooms()[%d].Capacity = %v, want %v", i, valueOrNilInt(got[i].Capacity), valueOrNilInt(expected[i].Capacity))
		}
		if valueOrNilTime(got[i].CreatedAt) != valueOrNilTime(expected[i].CreatedAt) {
			t.Fatalf("GetAllRooms()[%d].CreatedAt = %v, want %v", i, valueOrNilTime(got[i].CreatedAt), valueOrNilTime(expected[i].CreatedAt))
		}
	}
}

func valueOrNilString(s *string) interface{} {
	if s == nil {
		return nil
	}
	return *s
}

func valueOrNilInt(i *int) interface{} {
	if i == nil {
		return nil
	}
	return *i
}

func valueOrNilTime(t *time.Time) interface{} {
	if t == nil {
		return nil
	}
	return *t
}
