package user

import (
	"context"
	"errors"

	"github.com/Grisha1Kadetov/slotbooking/internal/model/user"
	sq "github.com/Masterminds/squirrel"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var psq = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

const uniqueViolationCode = "23505"

type UserRepository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (repo *UserRepository) Create(ctx context.Context, u user.User) (user.User, error) {
	query, args, _ := psq.Insert("users").
		Columns("email", "password_hash", "role").
		Values(u.Email, u.PasswordHash, u.Role).
		Suffix("RETURNING id, created_at").ToSql()

	err := repo.pool.QueryRow(ctx, query, args...).Scan(&u.ID, &u.CreatedAt)
	if err != nil {
		return user.User{}, checkUnicError(err)
	}

	return u, nil
}

func (repo *UserRepository) GetByEmail(ctx context.Context, email string) (user.User, error) {
	query, args, _ := psq.Select("id", "email", "password_hash", "role", "created_at").From("users").Where(sq.Eq{"email": email}).Limit(1).ToSql()

	var u user.User
	err := repo.pool.QueryRow(ctx, query, args...).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return user.User{}, user.ErrNotFound
		}
		return user.User{}, err
	}

	return u, nil
}

func checkUnicError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == uniqueViolationCode {
		return user.ErrUnicEmail
	}
	return err
}
