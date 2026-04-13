package booking

import "errors"

var (
	ErrNotOwner          = errors.New("forbiden")
	ErrUserNotFound      = errors.New("user not found")
	ErrNotFound          = errors.New("booking not found")
	ErrSlotNotFound      = errors.New("slot not found")
	ErrSlotAlreadyBooked = errors.New("slot already booked")
	ErrOldBooking        = errors.New("only futur slots can be booked")
)
