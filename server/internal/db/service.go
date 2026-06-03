//go:generate sqlc generate -f ../../sqlc.yaml
package db

import (
	"context"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	Queries *Queries
	Pool    *pgxpool.Pool
}

func InitStore(ctx context.Context) (*Store, error) {
	dbURL := os.Getenv("DATABASE_URL")

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		return nil, err
	}

	return &Store{
		Queries: New(pool),
		Pool:    pool,
	}, nil
}

func (s *Store) Close() {
	if s.Pool != nil {
		s.Pool.Close()
	}
}
