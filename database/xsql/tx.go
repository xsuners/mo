package xsql

import (
	"context"
	"database/sql"
)

func (db *Database) Tx(ctx context.Context, fns ...func(tx *Tx) error) error {
	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}
	for _, fn := range fns {
		if err = fn(&Tx{tx}); err != nil {
			if err = tx.Rollback(); err != nil {
				return err
			}
			return err
		}
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}

type Tx struct {
	*sql.Tx
}

func (tx *Tx) Exec(ctx context.Context, query string, args ...interface{}) (af, id int64, err error) {
	ret, err := tx.Tx.ExecContext(ctx, query, args...)
	if err != nil {
		return
	}
	af, err = ret.RowsAffected()
	if err != nil {
		return
	}
	id, err = ret.LastInsertId()
	if err != nil {
		return
	}
	return
}
