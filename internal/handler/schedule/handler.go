package schedule

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/Grisha1Kadetov/slotbooking/internal/model/schedule"
	errrend "github.com/Grisha1Kadetov/slotbooking/internal/pkg/errorrenderer"
	"github.com/Grisha1Kadetov/slotbooking/internal/pkg/log"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/google/uuid"
)

const timeLayout = "15:04"

type ScheduleHandler struct {
	service service
	l       log.Logger
}

func New(s service, l log.Logger) *ScheduleHandler {
	return &ScheduleHandler{
		service: s,
		l:       l,
	}
}

func (h *ScheduleHandler) Create(w http.ResponseWriter, r *http.Request) {
	var request CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		errrend.RenderBadRequestError(w, r, "invalid JSON")
		return
	}
	roomIDStr := chi.URLParam(r, "roomId")
	roomID, err := uuid.Parse(roomIDStr)
	if err != nil {
		errrend.RenderBadRequestError(w, r, "invalid roomId")
		return
	}
	s, err := validateCreateRequest(w, r, request)
	if err != nil {
		return
	}
	s.RoomID = roomID

	createdSchedule, err := h.service.Create(r.Context(), s)
	if err != nil {
		if errors.Is(err, schedule.ErrRoomNotFound) {
			errrend.RenderError(w, r, http.StatusNotFound, string(errrend.StatusRoomNotFound), "room not found")
			return
		}
		if errors.Is(err, schedule.ErrScheduleExists) {
			errrend.RenderError(w, r, http.StatusConflict, string(errrend.StatusScheduleExists), "schedule for this room already exists and cannot be changed")
			return
		}

		errrend.RenderInternalError(w, r)
		h.l.Error("failed to create schedule", log.Err(err))
		return
	}

	response := ToCreateResponse(createdSchedule)
	render.Status(r, http.StatusCreated)
	render.JSON(w, r, response)
}

func validateCreateRequest(w http.ResponseWriter, r *http.Request, request CreateRequest) (schedule.Schedule, error) {
	err := errors.New("validate error")
	if request.DaysOfWeek == nil {
		errrend.RenderBadRequestError(w, r, "daysOfWeek is required")
		return schedule.Schedule{}, err
	}
	if request.StartTime == "" {
		errrend.RenderBadRequestError(w, r, "startTime is required")
		return schedule.Schedule{}, err
	}
	if request.EndTime == "" {
		errrend.RenderBadRequestError(w, r, "endTime is required")
		return schedule.Schedule{}, err
	}

	startTime, err := time.Parse(timeLayout, request.StartTime)
	if err != nil {
		errrend.RenderBadRequestError(w, r, "startTime must be in HH:MM format")
		return schedule.Schedule{}, err
	}

	endTime, err := time.Parse(timeLayout, request.EndTime)
	if err != nil {
		errrend.RenderBadRequestError(w, r, "endTime must be in HH:MM format")
		return schedule.Schedule{}, err
	}

	if !startTime.Before(endTime) {
		errrend.RenderBadRequestError(w, r, "startTime must be before endTime")
		return schedule.Schedule{}, err
	}

	seen := make(map[int]any, len(request.DaysOfWeek))
	for _, day := range request.DaysOfWeek {
		if day < 1 || day > 7 {
			errrend.RenderBadRequestError(w, r, "daysOfWeek must contain values from 1 to 7")
			return schedule.Schedule{}, err
		}
		if _, ok := seen[day]; ok {
			errrend.RenderBadRequestError(w, r, "daysOfWeek must not contain duplicates")
			return schedule.Schedule{}, err
		}
		seen[day] = true
	}

	return schedule.Schedule{
		DaysOfWeek: request.DaysOfWeek,
		StartTime:  startTime,
		EndTime:    endTime,
	}, nil
}
