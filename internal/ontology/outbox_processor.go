package ontology

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// KafkaProducer publishes envelopes to Kafka or any stream bus.
type KafkaProducer interface {
	Publish(ctx context.Context, topic string, payload []byte) error
}

// TripleStoreClient ingests JSON-LD payloads into a graph database.
type TripleStoreClient interface {
	Insert(ctx context.Context, payload []byte) error
}

// OutboxProcessor drains ontology events into downstream sinks.
type OutboxProcessor struct {
	db        *sql.DB
	topic     string
	kafka     KafkaProducer
	triple    TripleStoreClient
	batchSize int
}

// NewOutboxProcessor creates a processor with the provided dependencies.
func NewOutboxProcessor(db *sql.DB, topic string, kafka KafkaProducer, triple TripleStoreClient) *OutboxProcessor {
	return &OutboxProcessor{
		db:        db,
		topic:     topic,
		kafka:     kafka,
		triple:    triple,
		batchSize: 100,
	}
}

// Run continuously drains the outbox with the supplied polling interval.
func (p *OutboxProcessor) Run(ctx context.Context, interval time.Duration) error {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		if err := p.ProcessBatch(ctx); err != nil {
			return err
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

// ProcessBatch reads the next batch and forwards it to every sink.
func (p *OutboxProcessor) ProcessBatch(ctx context.Context) error {
	if p == nil || p.db == nil {
		return errors.New("outbox processor uninitialized")
	}

	tx, err := p.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	rows, err := tx.QueryContext(ctx, `
        SELECT id, event_type, payload
        FROM ontology_outbox
        WHERE delivered_at IS NULL
        ORDER BY created_at
        FOR UPDATE SKIP LOCKED
        LIMIT $1`, p.batchSize)
	if err != nil {
		return fmt.Errorf("select outbox rows: %w", err)
	}
	defer rows.Close()

	type record struct {
		id    string
		event []byte
	}
	var batch []record

	for rows.Next() {
		var id, eventType string
		var payload []byte
		if err := rows.Scan(&id, &eventType, &payload); err != nil {
			return fmt.Errorf("scan outbox row: %w", err)
		}
		batch = append(batch, record{id: id, event: payload})
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate outbox rows: %w", err)
	}

	if len(batch) == 0 {
		return tx.Commit()
	}

	for _, rec := range batch {
		if err := p.dispatch(ctx, rec.event); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `UPDATE ontology_outbox SET delivered_at = NOW() WHERE id = $1`, rec.id); err != nil {
			return fmt.Errorf("mark delivered: %w", err)
		}
	}

	return tx.Commit()
}

func (p *OutboxProcessor) dispatch(ctx context.Context, payload []byte) error {
	if len(payload) == 0 {
		return nil
	}

	if p.kafka != nil && p.topic != "" {
		if err := p.kafka.Publish(ctx, p.topic, payload); err != nil {
			return fmt.Errorf("publish kafka: %w", err)
		}
	}

	if p.triple != nil {
		if err := p.triple.Insert(ctx, payload); err != nil {
			return fmt.Errorf("insert triple store: %w", err)
		}
	}

	return nil
}
