package booking

import (
	"time"

	"github.com/google/uuid"
)

type Booking struct {
	Id             uuid.UUID
	SlotId         uuid.UUID
	UserId         uuid.UUID
	Status         Status
	ConferenceLink *string
	CreatedAt      *time.Time
}

type Status string

const (
	StatusActive    Status = "active"
	StatusCancelled Status = "cancelled"
)
