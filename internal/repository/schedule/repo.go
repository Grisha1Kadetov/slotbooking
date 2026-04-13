package schedule

import (
	"context"
	"errors"

	"github.com/Grisha1Kadetov/slotbooking/internal/model/schedule"
	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var psq = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

const (
	UniqueViolationCode     = "23505"
	ForeignKeyViolationCode = "23503"
)

type ScheduleRepository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *ScheduleRepository {
	return &ScheduleRepository{pool: pool}
}

func (repo *ScheduleRepository) Create(ctx context.Context, s schedule.Schedule) (schedule.Schedule, error) {
	query, args, _ := psq.Insert("schedules").
		Columns("room_id", "days_of_week", "start_time", "end_time").
		Values(s.RoomID, s.DaysOfWeek, s.StartTime, s.EndTime).
		Suffix("RETURNING id").ToSql()

	var id uuid.UUID
	err := repo.pool.QueryRow(ctx, query, args...).Scan(&id)
	if err != nil {
		return schedule.Schedule{}, checkError(err)
	}

	s.ID = id
	return s, nil
}

func (repo *ScheduleRepository) GetByRoomId(ctx context.Context, roomID uuid.UUID) (schedule.Schedule, error) {
	query, args, _ := psq.Select("id", "room_id", "days_of_week", "start_time", "end_time").
		From("schedules").
		Where(sq.Eq{"room_id": roomID}).ToSql()

	var s schedule.Schedule
	err := repo.pool.QueryRow(ctx, query, args...).Scan(
		&s.ID,
		&s.RoomID,
		&s.DaysOfWeek,
		&s.StartTime,
		&s.EndTime,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return schedule.Schedule{}, schedule.ErrNotFound
		}
		return schedule.Schedule{}, err
	}

	return s, nil
}

func checkError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case UniqueViolationCode:
			return schedule.ErrScheduleExists
		case ForeignKeyViolationCode:
			return schedule.ErrRoomNotFound
		}
	}
	return err
}
