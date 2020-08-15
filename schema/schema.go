package schema

import (
	"database/sql"
	"fmt"

	"golang.org/x/net/context"
)

type Change struct {
	Version int
	Name    string
	Fwd     func(ctx context.Context, tx *sql.Tx) error
	Rev     func(ctx context.Context, tx *sql.Tx) error
}

func Schema() []Change {
	schema := []Change{
		{1597460998, "1597460998_init", initUp, initDown},
	}
	return schema
}

func Apply(ctx context.Context, db *sql.DB, changes []Change, version int) error {
	curVersion, err := userVersion(ctx, db)
	if err != nil {
		return err
	}

	for _, c := range changes {
		// Skip anything already applied.
		if c.Version <= curVersion {
			continue
		}

		if c.Version <= version {
			applyOne(ctx, db, c)
		}
	}

	return nil
}

func userVersion(ctx context.Context, db *sql.DB) (int, error) {
	var version int

	err := db.QueryRowContext(ctx, "PRAGMA user_version;").Scan(&version)
	if err == sql.ErrNoRows {
		return 0, nil
	}

	if err != nil {
		return 0, err
	}

	return version, nil
}

func applyOne(ctx context.Context, db *sql.DB, change Change) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("beginTx[%s]: %s", change.Name, err)
	}

	if err = change.Fwd(ctx, tx); err != nil {
		tx.Rollback()
		return fmt.Errorf("fwd[%s]: %s", change.Name, err)
	}

	q := fmt.Sprintf("PRAGMA user_version = %d", change.Version)
	if _, err = tx.ExecContext(ctx, q); err != nil {
		tx.Rollback()
		return fmt.Errorf("userVersion[%s]: %s", change.Name, err)
	}

	return tx.Commit()
}
