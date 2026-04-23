package slot

import (
	"context"
	"errors"
	"time"

	"github.com/Grisha1Kadetov/slotbooking/internal/model/slot"
	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var psq = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

const (
	ForeignKeyViolationCode = "23503"
	ExclusionViolationCode  = "23P01"
)

type SlotRepository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *SlotRepository {
	return &SlotRepository{pool: pool}
}

func (repo *SlotRepository) GetById(ctx context.Context, slotId uuid.UUID) (slot.Slot, error) {
	query, args, _ := psq.Select("id", "room_id", "start_at", "end_at").
		From("slots").Where(sq.Eq{"id": slotId}).ToSql()

	row := repo.pool.QueryRow(ctx, query, args...)

	var s slot.Slot
	if err := row.Scan(&s.ID, &s.RoomID, &s.StartAt, &s.EndAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return slot.Slot{}, slot.ErrNotFound
		}
		return slot.Slot{}, err
	}

	return s, nil
}

func (repo *SlotRepository) CreateBulk(ctx context.Context, slots []slot.Slot) ([]slot.Slot, error) {
	if len(slots) == 0 {
		return []slot.Slot{}, nil
	}

	builder := psq.Insert("slots").
		Columns("room_id", "start_at", "end_at")

	for _, s := range slots {
		builder = builder.Values(s.RoomID, s.StartAt, s.EndAt)
	}

	query, args, err := builder.Suffix("RETURNING id, room_id, start_at, end_at").ToSql()
	if err != nil {
		return nil, nil
	}

	rows, err := repo.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, checkError(err)
	}
	defer rows.Close()

	var result []slot.Slot
	for rows.Next() {
		var s slot.Slot
		if err := rows.Scan(&s.ID, &s.RoomID, &s.StartAt, &s.EndAt); err != nil {
			return nil, err
		}
		result = append(result, s)
	}

	return result, rows.Err()
}

func (repo *SlotRepository) GetAvailableByRoomIdAndDate(ctx context.Context, roomID uuid.UUID, date time.Time) ([]slot.Slot, error) {
	nextDate := date.AddDate(0, 0, 1)
	lowerBound := date
	now := time.Now().UTC()
	if now.After(lowerBound) {
		lowerBound = now
	}

	query, args, _ := psq.Select(
		"s.id",
		"s.room_id",
		"s.start_at",
		"s.end_at",
	).
		From("slots s").
		LeftJoin("bookings b ON b.slot_id = s.id AND b.status = 'active'").
		Where(sq.Eq{"s.room_id": roomID}).
		Where("s.start_at >= ? AND s.start_at < ?", lowerBound, nextDate).
		Where("b.slot_id IS NULL").
		OrderBy("s.start_at ASC").
		ToSql()

	rows, err := repo.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	slots := []slot.Slot{}
	for rows.Next() {
		var s slot.Slot
		err := rows.Scan(
			&s.ID,
			&s.RoomID,
			&s.StartAt,
			&s.EndAt,
		)
		if err != nil {
			return nil, err
		}

		slots = append(slots, s)
	}

	return slots, rows.Err()
}

func (repo *SlotRepository) HasAnyByRoomIdAndDate(ctx context.Context, roomID uuid.UUID, date time.Time) (bool, error) {
	nextDate := date.AddDate(0, 0, 1)

	query, args, _ := psq.Select("1").
		From("slots").
		Where(sq.Eq{"room_id": roomID}).
		Where("start_at >= ? AND start_at < ?", date, nextDate).
		Limit(1).
		ToSql()

	var one int
	err := repo.pool.QueryRow(ctx, query, args...).Scan(&one)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func checkError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case ForeignKeyViolationCode:
			return slot.ErrRoomNotFound
		case ExclusionViolationCode:
			return slot.ErrTimeOverlap
		}
	}
	return err
}
