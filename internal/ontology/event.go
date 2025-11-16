package ontology

import (
	"encoding/json"
	"time"
)

const (
	// DefaultContext is the JSON-LD context emitted with every event.
	DefaultContext = "https://ontology.skillsphere.dev/generated.context.jsonld"
)

// Event represents a JSON-LD resource with arbitrary properties.
type Event struct {
	Context   string
	ID        string
	Type      string
	Timestamp time.Time

	props map[string]any
}

// NewEvent creates a new event with the default JSON-LD context.
func NewEvent(id, typ string) Event {
	return Event{
		Context: DefaultContext,
		ID:      id,
		Type:    typ,
		props:   make(map[string]any),
	}
}

// SetContext overrides the default JSON-LD context.
func (e *Event) SetContext(ctx string) {
	if ctx == "" {
		return
	}
	e.Context = ctx
}

// SetTimestamp attaches an RFC3339 timestamp to the event.
func (e *Event) SetTimestamp(ts time.Time) {
	if ts.IsZero() {
		return
	}
	e.Timestamp = ts
}

// Set assigns a JSON-LD property.
func (e *Event) Set(key string, value any) {
	if e.props == nil {
		e.props = make(map[string]any)
	}
	e.props[key] = value
}

// Properties returns the internal property map for inspection/testing.
func (e *Event) Properties() map[string]any {
	return e.props
}

// MarshalJSON renders the event as a JSON-LD document.
func (e Event) MarshalJSON() ([]byte, error) {
	document := make(map[string]any, len(e.props)+3)

	ctx := e.Context
	if ctx == "" {
		ctx = DefaultContext
	}

	document["@context"] = ctx
	if e.ID != "" {
		document["@id"] = e.ID
	}
	if e.Type != "" {
		document["@type"] = e.Type
	}
	if !e.Timestamp.IsZero() {
		document["schema:dateModified"] = e.Timestamp.UTC().Format(time.RFC3339Nano)
	}

	for key, value := range e.props {
		document[key] = value
	}

	return json.Marshal(document)
}
