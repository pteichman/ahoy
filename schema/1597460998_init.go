package schema

import (
	"database/sql"

	"golang.org/x/net/context"
)

func initUp(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, "CREATE TABLE users ()")
	if err != nil {
		return err
	}

	return nil
}

func initDown(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, "DROP TABLE users")
	if err != nil {
		return err
	}

	return nil
}
