package schedule

import "errors"

var (
	ErrRoomNotFound   = errors.New("room not found")
	ErrNotFound       = errors.New("schedule not found")
	ErrScheduleExists = errors.New("schedule already exists")
)
