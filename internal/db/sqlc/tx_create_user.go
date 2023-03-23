package db

import "context"

type CreateUserTxParams struct {
	CreateUserParams
	AfterCreate func(user User) error
}

type CreateUserTxResult struct {
	User User
}

func (store *SQLStore) CreateUserTx(
	ctx context.Context,
	arg CreateUserTxParams,
) (result CreateUserTxResult, txError error) {
	txError = store.execTx(ctx, func(q *Queries) (err error) {
		if result.User, err = q.CreateUser(ctx, arg.CreateUserParams); err != nil {
			return err
		}

		return arg.AfterCreate(result.User)
	})

	return result, txError
}
