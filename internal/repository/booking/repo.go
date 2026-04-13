package booking

import (
	"context"
	"errors"
	"time"

	"github.com/Grisha1Kadetov/slotbooking/internal/model/booking"
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

type BookingRepository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *BookingRepository {
	return &BookingRepository{pool: pool}
}

func (repo *BookingRepository) Create(ctx context.Context, b booking.Booking) (booking.Booking, error) {
	query, args, _ := psq.Insert("bookings").
		Columns("slot_id", "user_id", "status", "conference_link").
		Values(b.SlotId, b.UserId, b.Status, b.ConferenceLink).
		Suffix("RETURNING id, created_at").
		ToSql()

	var id uuid.UUID
	var createdAt time.Time

	err := repo.pool.QueryRow(ctx, query, args...).Scan(&id, &createdAt)
	if err != nil {
		return booking.Booking{}, checkError(err)
	}

	b.Id = id
	b.CreatedAt = &createdAt
	return b, nil
}

func (repo *BookingRepository) GetListAll(ctx context.Context, page, pageSize int) ([]booking.Booking, int, error) {
	var total int
	err := repo.pool.QueryRow(ctx, "SELECT COUNT(*) FROM bookings").Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize

	query, args, _ := psq.Select("id", "slot_id", "user_id", "status", "conference_link", "created_at").
		From("bookings").
		OrderBy("created_at DESC").
		Limit(uint64(pageSize)).
		Offset(uint64(offset)).
		ToSql()

	rows, err := repo.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, total, err
	}
	defer rows.Close()

	bookings := []booking.Booking{}
	for rows.Next() {
		var b booking.Booking
		var createdAt time.Time

		err := rows.Scan(
			&b.Id,
			&b.SlotId,
			&b.UserId,
			&b.Status,
			&b.ConferenceLink,
			&createdAt,
		)
		if err != nil {
			return nil, total, err
		}

		b.CreatedAt = &createdAt
		bookings = append(bookings, b)
	}

	return bookings, total, rows.Err()
}

func (repo *BookingRepository) GetByUserId(ctx context.Context, userID uuid.UUID) ([]booking.Booking, error) {
	query, args, _ := psq.Select("b.id", "b.slot_id", "b.user_id", "b.status", "b.conference_link", "b.created_at").
		From("bookings b").
		Join("slots s ON s.id = b.slot_id").
		Where(sq.Eq{"b.user_id": userID}).
		Where("s.start_at > NOW()").
		OrderBy("s.start_at ASC").
		ToSql()

	rows, err := repo.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	bookings := []booking.Booking{}
	for rows.Next() {
		var b booking.Booking
		var createdAt time.Time

		err := rows.Scan(
			&b.Id,
			&b.SlotId,
			&b.UserId,
			&b.Status,
			&b.ConferenceLink,
			&createdAt,
		)
		if err != nil {
			return nil, err
		}

		b.CreatedAt = &createdAt
		bookings = append(bookings, b)
	}

	return bookings, rows.Err()
}

func (repo *BookingRepository) CancelById(ctx context.Context, bookingID uuid.UUID) (booking.Booking, error) {
	query, args, _ := psq.Update("bookings").
		Set("status", booking.StatusCancelled).
		Where(sq.Eq{"id": bookingID}).
		Suffix("RETURNING id, slot_id, user_id, status, conference_link, created_at").
		ToSql()

	var b booking.Booking
	var createdAt time.Time

	err := repo.pool.QueryRow(ctx, query, args...).Scan(
		&b.Id,
		&b.SlotId,
		&b.UserId,
		&b.Status,
		&b.ConferenceLink,
		&createdAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return booking.Booking{}, booking.ErrNotFound
		}
		return booking.Booking{}, err
	}

	b.CreatedAt = &createdAt
	return b, nil
}

func (repo *BookingRepository) GetById(ctx context.Context, bookingID uuid.UUID) (booking.Booking, error) {
	query, args, _ := psq.Select("id", "slot_id", "user_id", "status", "conference_link", "created_at").
		From("bookings").
		Where(sq.Eq{"id": bookingID}).
		ToSql()

	var b booking.Booking
	var createdAt time.Time

	err := repo.pool.QueryRow(ctx, query, args...).Scan(
		&b.Id,
		&b.SlotId,
		&b.UserId,
		&b.Status,
		&b.ConferenceLink,
		&createdAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return booking.Booking{}, booking.ErrNotFound
		}
		return booking.Booking{}, err
	}

	b.CreatedAt = &createdAt
	return b, nil
}

func checkError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case UniqueViolationCode:
			return booking.ErrSlotAlreadyBooked
		case ForeignKeyViolationCode:
			if pgErr.ConstraintName == "bookings_slot_id_fkey" {
				return booking.ErrSlotNotFound
			}
			return booking.ErrUserNotFound
		}
	}

	return err
}
