package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/bsko/casino-transaction-system/internal/entity"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

const (
	defaultLimit = 1000
)

type TransactionEventRepository struct {
	masterDB *DB
	slaveDB  *DB
}

type transactionEventRow struct {
	ID              int64     `db:"id"`
	UserID          string    `db:"user_id"`
	TransactionType string    `db:"transaction_type"`
	Amount          int64     `db:"amount"`
	CreatedAt       time.Time `db:"created_at"`
}

func NewTransactionEventRepository(master *DB, slave *DB) *TransactionEventRepository {
	return &TransactionEventRepository{
		masterDB: master,
		slaveDB:  slave,
	}
}

func (t *TransactionEventRepository) GetListByFilter(filter entity.TransactionEventFilter) ([]entity.TransactionEvent, error) {
	if t.slaveDB == nil {
		log.Printf("failed to connect to slave database")
		return nil, sql.ErrConnDone
	}

	qb := sq.Select("id", "user_id", "transaction_type", "amount", "created_at").
		From("transaction_events").
		PlaceholderFormat(sq.Dollar).
		OrderBy("created_at DESC")

	if filter.UserID != nil {
		qb = qb.Where(sq.Eq{"user_id": filter.UserID.UUID.String()})
	}

	if filter.TransactionType != nil {
		qb = qb.Where(sq.Eq{"transaction_type": string(*filter.TransactionType)})
	}

	if filter.AmountFrom != nil {
		qb = qb.Where(sq.GtOrEq{"amount": int64(*filter.AmountFrom)})
	}

	if filter.AmountTo != nil {
		qb = qb.Where(sq.LtOrEq{"amount": int64(*filter.AmountTo)})
	}

	if filter.CreatedFrom != nil {
		qb = qb.Where(sq.GtOrEq{"created_at": *filter.CreatedFrom})
	}

	if filter.CreatedTo != nil {
		qb = qb.Where(sq.LtOrEq{"created_at": *filter.CreatedTo})
	}

	limit := filter.Limit
	if limit <= 0 || limit > defaultLimit {
		limit = defaultLimit
	}
	qb = qb.Limit(uint64(limit))

	if filter.Offset > 0 {
		qb = qb.Offset(uint64(filter.Offset))
	}

	query, args, err := qb.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	var rows []transactionEventRow
	err = t.slaveDB.Select(&rows, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch transactions: %w", err)
	}

	events := make([]entity.TransactionEvent, 0, len(rows))
	for _, row := range rows {
		parsedUUID, err := uuid.Parse(row.UserID)
		if err != nil {
			return nil, fmt.Errorf("failed to parse user_id: %w", err)
		}

		events = append(events, entity.TransactionEvent{
			UserID: entity.UserID{
				UUID: parsedUUID,
			},
			TransactionType: entity.TransactionType(row.TransactionType),
			Amount:          entity.Money(row.Amount),
			CreatedAt:       row.CreatedAt,
		})
	}

	return events, nil
}

func (t *TransactionEventRepository) BatchStore(ctx context.Context, batch []entity.TransactionEvent) error {
	if t.masterDB == nil {
		return fmt.Errorf("master database connection is not initialized, call Connect() first")
	}

	if len(batch) == 0 {
		return nil
	}

	qb := sq.Insert("transaction_events").
		Columns("user_id", "transaction_type", "amount", "created_at").
		PlaceholderFormat(sq.Dollar)

	for _, event := range batch {
		qb = qb.Values(
			event.UserID.UUID.String(),
			string(event.TransactionType),
			int64(event.Amount),
			event.CreatedAt,
		)
	}

	query, args, err := qb.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build insert query: %w", err)
	}

	_, err = t.masterDB.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to insert transactions: %w", err)
	}
	return nil
}
