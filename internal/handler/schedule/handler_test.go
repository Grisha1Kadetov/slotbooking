package schedule_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	handlerschedule "github.com/Grisha1Kadetov/slotbooking/internal/handler/schedule"
	schedulemodel "github.com/Grisha1Kadetov/slotbooking/internal/model/schedule"
	"github.com/Grisha1Kadetov/slotbooking/internal/pkg/log"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type loggerMock struct{}

func (l *loggerMock) Debug(msg string, fields ...log.Field) {}
func (l *loggerMock) Info(msg string, fields ...log.Field)  {}
func (l *loggerMock) Warn(msg string, fields ...log.Field)  {}
func (l *loggerMock) Error(msg string, fields ...log.Field) {}
func (l *loggerMock) Panic(msg string, fields ...log.Field) {}
func (l *loggerMock) Close()                                {}

type serviceMock struct {
	createFunc      func(ctx context.Context, s schedulemodel.Schedule) (schedulemodel.Schedule, error)
	getByRoomIdFunc func(ctx context.Context, roomID uuid.UUID) (schedulemodel.Schedule, error)
}

func (m *serviceMock) Create(ctx context.Context, s schedulemodel.Schedule) (schedulemodel.Schedule, error) {
	if m.createFunc == nil {
		return schedulemodel.Schedule{}, errors.New("unexpected Create call")
	}
	return m.createFunc(ctx, s)
}

func (m *serviceMock) GetByRoomId(ctx context.Context, roomID uuid.UUID) (schedulemodel.Schedule, error) {
	if m.getByRoomIdFunc == nil {
		return schedulemodel.Schedule{}, errors.New("unexpected GetByRoomId call")
	}
	return m.getByRoomIdFunc(ctx, roomID)
}

func requestWithRoomID(body []byte, roomID string) *http.Request {
	req := httptest.NewRequest(http.MethodPost, "/rooms/"+roomID+"/schedule/create", bytes.NewReader(body))
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("roomId", roomID)
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
	return req.WithContext(ctx)
}

func TestScheduleHandler_Create(t *testing.T) {
	t.Run("invalid json", func(t *testing.T) {
		h := handlerschedule.New(&serviceMock{}, &loggerMock{})
		req := requestWithRoomID([]byte("{"), uuid.New().String())
		rr := httptest.NewRecorder()

		h.Create(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusBadRequest)
		}
	})

	t.Run("invalid room id", func(t *testing.T) {
		h := handlerschedule.New(&serviceMock{}, &loggerMock{})
		body := []byte(`{"daysOfWeek":[1],"startTime":"09:00","endTime":"18:00"}`)
		req := requestWithRoomID(body, "bad-uuid")
		rr := httptest.NewRecorder()

		h.Create(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusBadRequest)
		}
	})

	t.Run("daysOfWeek is required", func(t *testing.T) {
		h := handlerschedule.New(&serviceMock{}, &loggerMock{})
		body := []byte(`{"startTime":"09:00","endTime":"18:00"}`)
		req := requestWithRoomID(body, uuid.New().String())
		rr := httptest.NewRecorder()

		h.Create(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusBadRequest)
		}
	})

	t.Run("startTime is required", func(t *testing.T) {
		h := handlerschedule.New(&serviceMock{}, &loggerMock{})
		body := []byte(`{"daysOfWeek":[1],"endTime":"18:00"}`)
		req := requestWithRoomID(body, uuid.New().String())
		rr := httptest.NewRecorder()

		h.Create(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusBadRequest)
		}
	})

	t.Run("endTime is required", func(t *testing.T) {
		h := handlerschedule.New(&serviceMock{}, &loggerMock{})
		body := []byte(`{"daysOfWeek":[1],"startTime":"09:00"}`)
		req := requestWithRoomID(body, uuid.New().String())
		rr := httptest.NewRecorder()

		h.Create(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusBadRequest)
		}
	})

	t.Run("invalid startTime format", func(t *testing.T) {
		h := handlerschedule.New(&serviceMock{}, &loggerMock{})
		body := []byte(`{"daysOfWeek":[1],"startTime":"9","endTime":"18:00"}`)
		req := requestWithRoomID(body, uuid.New().String())
		rr := httptest.NewRecorder()

		h.Create(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusBadRequest)
		}
	})

	t.Run("invalid endTime format", func(t *testing.T) {
		h := handlerschedule.New(&serviceMock{}, &loggerMock{})
		body := []byte(`{"daysOfWeek":[1],"startTime":"09:00","endTime":"18"}`)
		req := requestWithRoomID(body, uuid.New().String())
		rr := httptest.NewRecorder()

		h.Create(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusBadRequest)
		}
	})

	t.Run("startTime must be before endTime", func(t *testing.T) {
		h := handlerschedule.New(&serviceMock{}, &loggerMock{})
		body := []byte(`{"daysOfWeek":[1],"startTime":"18:00","endTime":"09:00"}`)
		req := requestWithRoomID(body, uuid.New().String())
		rr := httptest.NewRecorder()

		h.Create(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusBadRequest)
		}
	})

	t.Run("daysOfWeek out of range", func(t *testing.T) {
		h := handlerschedule.New(&serviceMock{}, &loggerMock{})
		body := []byte(`{"daysOfWeek":[0],"startTime":"09:00","endTime":"18:00"}`)
		req := requestWithRoomID(body, uuid.New().String())
		rr := httptest.NewRecorder()

		h.Create(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusBadRequest)
		}
	})

	t.Run("daysOfWeek duplicates", func(t *testing.T) {
		h := handlerschedule.New(&serviceMock{}, &loggerMock{})
		body := []byte(`{"daysOfWeek":[1,1],"startTime":"09:00","endTime":"18:00"}`)
		req := requestWithRoomID(body, uuid.New().String())
		rr := httptest.NewRecorder()

		h.Create(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusBadRequest)
		}
	})

	t.Run("room not found", func(t *testing.T) {
		h := handlerschedule.New(&serviceMock{
			createFunc: func(ctx context.Context, s schedulemodel.Schedule) (schedulemodel.Schedule, error) {
				return schedulemodel.Schedule{}, schedulemodel.ErrRoomNotFound
			},
		}, &loggerMock{})
		body := []byte(`{"daysOfWeek":[1],"startTime":"09:00","endTime":"18:00"}`)
		req := requestWithRoomID(body, uuid.New().String())
		rr := httptest.NewRecorder()

		h.Create(rr, req)

		if rr.Code != http.StatusNotFound {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusNotFound)
		}
	})

	t.Run("schedule already exists", func(t *testing.T) {
		h := handlerschedule.New(&serviceMock{
			createFunc: func(ctx context.Context, s schedulemodel.Schedule) (schedulemodel.Schedule, error) {
				return schedulemodel.Schedule{}, schedulemodel.ErrScheduleExists
			},
		}, &loggerMock{})
		body := []byte(`{"daysOfWeek":[1],"startTime":"09:00","endTime":"18:00"}`)
		req := requestWithRoomID(body, uuid.New().String())
		rr := httptest.NewRecorder()

		h.Create(rr, req)

		if rr.Code != http.StatusConflict {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusConflict)
		}
	})

	t.Run("success", func(t *testing.T) {
		roomID := uuid.New()
		scheduleID := uuid.New()
		h := handlerschedule.New(&serviceMock{
			createFunc: func(ctx context.Context, s schedulemodel.Schedule) (schedulemodel.Schedule, error) {
				if s.RoomID != roomID {
					t.Fatalf("roomID = %v, want %v", s.RoomID, roomID)
				}
				if len(s.DaysOfWeek) != 2 || s.DaysOfWeek[0] != 1 || s.DaysOfWeek[1] != 3 {
					t.Fatalf("daysOfWeek = %v, want [1 3]", s.DaysOfWeek)
				}
				if s.StartTime.Format("15:04") != "09:00" {
					t.Fatalf("startTime = %s, want 09:00", s.StartTime.Format("15:04"))
				}
				if s.EndTime.Format("15:04") != "18:00" {
					t.Fatalf("endTime = %s, want 18:00", s.EndTime.Format("15:04"))
				}
				s.ID = scheduleID
				return s, nil
			},
		}, &loggerMock{})
		body := []byte(`{"daysOfWeek":[1,3],"startTime":"09:00","endTime":"18:00"}`)
		req := requestWithRoomID(body, roomID.String())
		rr := httptest.NewRecorder()

		h.Create(rr, req)

		if rr.Code != http.StatusCreated {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusCreated)
		}

		var resp handlerschedule.CreateResponse
		if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
			t.Fatalf("unmarshal response: %v", err)
		}
		if resp.Schedule.ID != scheduleID {
			t.Fatalf("schedule id = %v, want %v", resp.Schedule.ID, scheduleID)
		}
		if resp.Schedule.RoomID != roomID {
			t.Fatalf("room id = %v, want %v", resp.Schedule.RoomID, roomID)
		}
		if resp.Schedule.StartTime != "09:00" {
			t.Fatalf("startTime = %q, want %q", resp.Schedule.StartTime, "09:00")
		}
		if resp.Schedule.EndTime != "18:00" {
			t.Fatalf("endTime = %q, want %q", resp.Schedule.EndTime, "18:00")
		}
	})
}
