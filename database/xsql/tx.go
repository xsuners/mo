package xsql

import (
	"context"
	"database/sql"
)

type Tx struct {
	*sql.Tx
}

var _ SQL = (*Tx)(nil)

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

func (tx *Tx) Query(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return tx.QueryContext(ctx, query, args...)
}

func (tx *Tx) One(ctx context.Context, query string, args ...any) *sql.Row {
	return tx.QueryRowContext(ctx, query, args...)
}
