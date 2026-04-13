package errorrenderer

import (
	"net/http"

	"github.com/go-chi/render"
)

type err struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type errorResponse struct {
	Error err `json:"error"`
}

type Status string

const (
	StatusInvalidRequest    Status = "INVALID_REQUEST"
	StatusUnauthorized      Status = "UNAUTHORIZED"
	StatusNotFound          Status = "NOT_FOUND"
	StatusRoomNotFound      Status = "ROOM_NOT_FOUND"
	StatusSlotNotFound      Status = "SLOT_NOT_FOUND"
	StatusSlotAlreadyBooked Status = "SLOT_ALREADY_BOOKED"
	StatusBookingNotFound   Status = "BOOKING_NOT_FOUND"
	StatusForbidden         Status = "FORBIDDEN"
	StatusScheduleExists    Status = "SCHEDULE_EXISTS"
	StatusInternalError     Status = "INTERNAL_ERROR"
)

func RenderError(w http.ResponseWriter, r *http.Request, status int, code, message string) {
	render.Status(r, status)
	render.JSON(w, r, errorResponse{
		Error: err{
			Code:    code,
			Message: message,
		},
	})
}

func RenderInternalError(w http.ResponseWriter, r *http.Request) {
	RenderError(w, r, http.StatusInternalServerError, string(StatusInternalError), "internal server error")
}

func RenderBadRequestError(w http.ResponseWriter, r *http.Request, message string) {
	RenderError(w, r, http.StatusBadRequest, string(StatusInvalidRequest), message)
}
