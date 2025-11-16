package ontology

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
)

// Emitter publishes ontology events to downstream sinks.
type Emitter interface {
	Emit(ctx context.Context, event Event) error
}

// Sender delivers marshalled JSON-LD payloads.
type Sender interface {
	Send(ctx context.Context, payload []byte) error
}

// SenderFunc adapts a function to the Sender interface.
type SenderFunc func(ctx context.Context, payload []byte) error

// Send implements Sender.
func (f SenderFunc) Send(ctx context.Context, payload []byte) error {
	return f(ctx, payload)
}

// JSONEmitter serializes events as JSON-LD documents and delivers them via a sender.
type JSONEmitter struct {
	sender Sender
}

// NewJSONEmitter constructs an emitter that uses the provided sender.
func NewJSONEmitter(sender Sender) *JSONEmitter {
	return &JSONEmitter{sender: sender}
}

// Emit marshals the event and delegates delivery to the sender.
func (e *JSONEmitter) Emit(ctx context.Context, event Event) error {
	if e == nil || e.sender == nil {
		return nil
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal ontology event: %w", err)
	}
	return e.sender.Send(ctx, payload)
}

// NopEmitter is a drop-in emitter that silently discards events.
type NopEmitter struct{}

// Emit implements Emitter.
func (NopEmitter) Emit(context.Context, Event) error { return nil }

// FanoutEmitter replicates events to multiple emitter implementations.
type FanoutEmitter []Emitter

// Emit forwards the event to each configured emitter, returning the first error encountered.
func (f FanoutEmitter) Emit(ctx context.Context, event Event) error {
	for _, emitter := range f {
		if emitter == nil {
			continue
		}
		if err := emitter.Emit(ctx, event); err != nil {
			return err
		}
	}
	return nil
}

// LogSender writes JSON-LD payloads to a slog logger.
type LogSender struct {
	logger *slog.Logger
}

// NewLogSender creates a sender that logs each event payload.
func NewLogSender(logger *slog.Logger) *LogSender {
	return &LogSender{logger: logger}
}

// Send implements Sender.
func (s *LogSender) Send(ctx context.Context, payload []byte) error {
	if s == nil || s.logger == nil {
		return nil
	}
	s.logger.InfoContext(ctx, "ontology event emitted", "payload", string(payload))
	return nil
}
