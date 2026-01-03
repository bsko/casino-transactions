package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/bsko/casino-transaction-system/internal/config"
	"github.com/jmoiron/sqlx"
)

type DB struct {
	conn *sqlx.DB
}

func NewDB(conn *sqlx.DB) *DB {
	return &DB{
		conn: conn,
	}
}

func (db *DB) Connect(conf *config.Postgres) error {
	connStr := fmt.Sprintf(
		"%s?user=%s&password=%s",
		conf.ConnectionString,
		conf.User,
		conf.Password,
	)

	conn, err := sqlx.Connect("postgres", connStr)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	conn.SetMaxOpenConns(conf.MaxOpenConns)
	conn.SetMaxIdleConns(conf.MaxIdleConns)
	conn.SetConnMaxLifetime(time.Duration(conf.MaxConnLifetime) * time.Minute)
	db.conn = conn
	return nil
}

func (db *DB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	tx, err := db.conn.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	res, err := tx.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}
	return res, nil
}

func (db *DB) Select(dest interface{}, query string, args ...interface{}) error {
	return db.conn.Select(dest, query, args...)
}

func (db *DB) Close() error {
	if err := db.conn.Close(); err != nil {
		return fmt.Errorf("failed to close connection: %w", err)
	}
	return nil
}
