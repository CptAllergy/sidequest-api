package db

import (
	"context"
	"fmt"
)

// TODO check back to this file when thinking about implementing some transactions again (like delete all user's quests)
func (s *SQLStore) ExecTx(ctx context.Context, fn func(*Queries) error) error {
	tx, err := s.connPool.Begin(ctx)
	if err != nil {
		return err
	}

	qtx := s.Queries.WithTx(tx)
	err = fn(qtx)
	if err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("tx err: %v, rb err: %v", err, rbErr)
		}
		return err
	}

	return tx.Commit(ctx)
}
