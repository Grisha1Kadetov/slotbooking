package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPool(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}

	const retries = 3
	for i := 0; i <= retries; i++ {
		if err = pool.Ping(ctx); err == nil {
			return pool, nil
		} else {
			time.Sleep(time.Second * 1)
		}
	}
	return nil, err
}
