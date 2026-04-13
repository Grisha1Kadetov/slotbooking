package schedule

import (
	"github.com/Grisha1Kadetov/slotbooking/internal/model/schedule"
	"github.com/google/uuid"
)

type CreateRequest struct {
	DaysOfWeek []int  `json:"daysOfWeek"`
	StartTime  string `json:"startTime"`
	EndTime    string `json:"endTime"`
}

type ScheduleResponse struct {
	ID         uuid.UUID `json:"id"`
	RoomID     uuid.UUID `json:"roomId"`
	DaysOfWeek []int     `json:"daysOfWeek"`
	StartTime  string    `json:"startTime"`
	EndTime    string    `json:"endTime"`
}

type CreateResponse struct {
	Schedule ScheduleResponse `json:"schedule"`
}

func ToScheduleResponse(s schedule.Schedule) ScheduleResponse {
	return ScheduleResponse{
		ID:         s.ID,
		RoomID:     s.RoomID,
		DaysOfWeek: s.DaysOfWeek,
		StartTime:  s.StartTime.Format("15:04"),
		EndTime:    s.EndTime.Format("15:04"),
	}
}

func ToCreateResponse(s schedule.Schedule) CreateResponse {
	return CreateResponse{
		Schedule: ToScheduleResponse(s),
	}
}
