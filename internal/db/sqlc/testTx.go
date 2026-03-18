package db

import "context"

// TODO Test setup for running a transaction. This is just an example of how to use the execTx function, and can be removed when we have actual transactions to run.
func (s *SQLStore) TestTx(ctx context.Context, arg int) (int, error) {

	err := s.ExecTx(ctx, func(q *Queries) error {
		return nil
	})

	return 0, err
}
