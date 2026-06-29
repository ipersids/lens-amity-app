//go:generate sqlc generate -f ../../sqlc.yaml
package db

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	Queries *Queries
	Pool    *pgxpool.Pool
}

func NewStore(ctx context.Context, dbURL string) (*Store, error) {
	if ctx == nil {
		return nil, errors.New("db: nil context")
	}
	if strings.TrimSpace(dbURL) == "" {
		return nil, errors.New("db: database url is required")
	}

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create pool: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("database ping failed: %w", err)
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
