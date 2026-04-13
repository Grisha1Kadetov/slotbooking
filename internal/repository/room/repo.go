package room

import (
	"context"
	"errors"
	"time"

	"github.com/Grisha1Kadetov/slotbooking/internal/model/room"
	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var psq = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

type RoomRepository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *RoomRepository {
	return &RoomRepository{pool: pool}
}

func (repo *RoomRepository) Create(ctx context.Context, r room.Room) (room.Room, error) {
	query, args, _ := psq.Insert("rooms").
		Columns("name", "description", "capacity").
		Values(r.Name, r.Description, r.Capacity).
		Suffix("RETURNING id, created_at").ToSql()

	var id uuid.UUID
	var createdAt time.Time
	err := repo.pool.QueryRow(ctx, query, args...).Scan(&id, &createdAt)
	if err != nil {
		return room.Room{}, err
	}
	r.ID = id
	r.CreatedAt = &createdAt

	return r, err
}

func (repo *RoomRepository) GetAll(ctx context.Context) ([]room.Room, error) {
	query, args, _ := psq.Select("id", "name", "description", "capacity", "created_at").From("rooms").ToSql()

	rows, err := repo.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rooms := []room.Room{}
	for rows.Next() {
		var room room.Room
		err := rows.Scan(&room.ID, &room.Name, &room.Description, &room.Capacity, &room.CreatedAt)
		if err != nil {
			return nil, err
		}
		rooms = append(rooms, room)
	}

	return rooms, rows.Err()
}

func (repo *RoomRepository) ExistsById(ctx context.Context, roomId uuid.UUID) (bool, error) {
	query, args, _ := psq.Select("1").
		From("rooms").
		Where(sq.Eq{"id": roomId}).
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
