package example

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// MyServiceRepository handles database operations for myservice
type MyServiceRepository interface {
	Create(ctx context.Context, input string, output string) (*MyServiceRecord, error)
	GetByID(ctx context.Context, id int64) (*MyServiceRecord, error)
	List(ctx context.Context, limit, offset int) ([]*MyServiceRecord, error)
}

type myServiceRepository struct {
	pool *pgxpool.Pool
}

// NewMyServiceRepository creates a new instance of MyServiceRepository
func NewMyServiceRepository(pool *pgxpool.Pool) MyServiceRepository {
	return &myServiceRepository{
		pool: pool,
	}
}

// Create inserts a new record into the database
func (r *myServiceRepository) Create(ctx context.Context, input string, output string) (*MyServiceRecord, error) {
	query := `
		INSERT INTO myservice_records (input, output, created_at, updated_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id, input, output, created_at, updated_at
	`

	now := time.Now()
	record := &MyServiceRecord{}

	err := r.pool.QueryRow(ctx, query, input, output, now, now).Scan(
		&record.ID,
		&record.Input,
		&record.Output,
		&record.CreatedAt,
		&record.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create record: %w", err)
	}

	return record, nil
}

// GetByID retrieves a record by ID
func (r *myServiceRepository) GetByID(ctx context.Context, id int64) (*MyServiceRecord, error) {
	query := `
		SELECT id, input, output, created_at, updated_at
		FROM myservice_records
		WHERE id = $1
	`

	record := &MyServiceRecord{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&record.ID,
		&record.Input,
		&record.Output,
		&record.CreatedAt,
		&record.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("record not found")
		}
		return nil, fmt.Errorf("failed to get record: %w", err)
	}

	return record, nil
}

// List retrieves a list of records with pagination
func (r *myServiceRepository) List(ctx context.Context, limit, offset int) ([]*MyServiceRecord, error) {
	query := `
		SELECT id, input, output, created_at, updated_at
		FROM myservice_records
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list records: %w", err)
	}
	defer rows.Close()

	var records []*MyServiceRecord
	for rows.Next() {
		record := &MyServiceRecord{}
		err := rows.Scan(
			&record.ID,
			&record.Input,
			&record.Output,
			&record.CreatedAt,
			&record.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan record: %w", err)
		}
		records = append(records, record)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return records, nil
}
