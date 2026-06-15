package database

import (
	"context"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type Querier interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type LockedConn struct {
	conn *pgx.Conn
	mu   sync.Mutex
}

func NewLockedConn(conn *pgx.Conn) *LockedConn {
	return &LockedConn{conn: conn}
}

func (db *LockedConn) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	db.mu.Lock()
	defer db.mu.Unlock()
	return db.conn.Exec(ctx, sql, args...)
}

func (db *LockedConn) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	db.mu.Lock()
	rows, err := db.conn.Query(ctx, sql, args...)
	if err != nil {
		db.mu.Unlock()
		return nil, err
	}
	return &lockedRows{Rows: rows, unlock: db.mu.Unlock}, nil
}

func (db *LockedConn) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	db.mu.Lock()
	return &lockedRow{row: db.conn.QueryRow(ctx, sql, args...), unlock: db.mu.Unlock}
}

type lockedRow struct {
	row    pgx.Row
	unlock func()
}

func (row *lockedRow) Scan(dest ...any) error {
	defer row.unlock()
	return row.row.Scan(dest...)
}

type lockedRows struct {
	pgx.Rows
	unlock func()
}

func (rows *lockedRows) Close() {
	rows.Rows.Close()
	rows.unlock()
}
