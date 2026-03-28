package db

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Store defines all functions to execute db queries and transactions
type Store interface {
	Querier
	Transactor
}

type Transactor interface {
	ExecTx(ctx context.Context, fn func(Querier) error) error
}

// SQLStore provides all functions to execute SQL queries and transactions
type SQLStore struct {
	connPool *pgxpool.Pool
	*Queries
}

// NewStore creates a new store
func NewStore(connPool *pgxpool.Pool) Store {
	return &SQLStore{
		connPool: connPool,
		Queries:  New(connPool),
	}
}

func (s *SQLStore) ExecTx(ctx context.Context, fn func(Querier) error) error {
	return pgx.BeginFunc(ctx, s.connPool, func(tx pgx.Tx) error {
		qtx := s.Queries.WithTx(tx)
		return fn(qtx)
	})
}
