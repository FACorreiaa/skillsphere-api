package ontology

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
)

const insertOutboxStmt = `
	INSERT INTO ontology_outbox (id, event_type, payload, created_at)
	VALUES ($1, $2, $3, $4)
`

// OutboxEmitter persists ontology events into a relational outbox table.
type OutboxEmitter struct {
	db     *sql.DB
	logger *slog.Logger
}

// NewOutboxEmitter builds an emitter backed by the provided sql.DB.
func NewOutboxEmitter(db *sql.DB, logger *slog.Logger) *OutboxEmitter {
	return &OutboxEmitter{
		db:     db,
		logger: logger,
	}
}

// Emit marshals the event and inserts it into the outbox.
func (e *OutboxEmitter) Emit(ctx context.Context, event Event) error {
	if e == nil || e.db == nil {
		return nil
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal ontology event: %w", err)
	}

	if _, err := e.db.ExecContext(
		ctx,
		insertOutboxStmt,
		uuid.New(),
		event.Type,
		payload,
		time.Now().UTC(),
	); err != nil {
		return fmt.Errorf("insert ontology outbox event: %w", err)
	}

	if e.logger != nil {
		e.logger.DebugContext(ctx, "ontology event enqueued", "event_type", event.Type)
	}
	return nil
}
