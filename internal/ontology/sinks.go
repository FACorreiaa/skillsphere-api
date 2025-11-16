package ontology

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
)

// LogProducer logs Kafka payloads.
type LogProducer struct {
	logger *slog.Logger
}

func NewLogProducer(logger *slog.Logger) *LogProducer {
	return &LogProducer{logger: logger}
}

// Publish implements KafkaProducer.
func (p *LogProducer) Publish(ctx context.Context, topic string, payload []byte) error {
	if p == nil || p.logger == nil {
		return nil
	}
	p.logger.InfoContext(ctx, "kafka publish", "topic", topic, "payload", string(payload))
	return nil
}

// HTTPTripleStoreClient posts JSON-LD payloads to an HTTP endpoint.
type HTTPTripleStoreClient struct {
	endpoint string
	client   *http.Client
}

func NewHTTPTripleStoreClient(endpoint string) *HTTPTripleStoreClient {
	return &HTTPTripleStoreClient{
		endpoint: endpoint,
		client:   &http.Client{},
	}
}

// Insert implements TripleStoreClient.
func (c *HTTPTripleStoreClient) Insert(ctx context.Context, payload []byte) error {
	if c == nil || c.endpoint == "" {
		return nil
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/ld+json")
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}
