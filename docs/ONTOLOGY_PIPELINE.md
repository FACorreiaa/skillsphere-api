# Ontology Event Pipeline

SkillSphere now emits JSON-LD envelopes whenever business events occur (user registration, session scheduling, match recommendations). Those events flow through the `ontology_outbox` table, a Go worker pushes them to Kafka and the triple store, and downstream services can reason over the live graph.

## Components

1. **Domain services** – `AuthService`, `SessionService`, and `MatchingService` call the builders in `internal/ontology` to generate JSON-LD payloads immediately after a state change succeeds.
2. **Outbox table** – `internal/ontology/OutboxEmitter` persists the JSON-LD blobs (`payload JSONB`) into `ontology_outbox` with a timestamp.
3. **Worker** – `cmd/ontologyworker` polls the outbox, publishes each event to Kafka (`ONTOLOGY_KAFKA_TOPIC`) via `LogProducer` (swap with a real producer later) and POSTs the same JSON-LD to the configured triple-store endpoint (`ONTOLOGY_TRIPLESTORE_ENDPOINT`).
4. **Triple store** – Any RDF store that accepts JSON-LD (Neptune, Blazegraph, Oxigraph). The worker's HTTP client POSTs the JSON-LD document using `application/ld+json` so you can attach it to SPARQL Update handlers or ingestion APIs.

## Running the Worker

```bash
export DB_HOST=localhost
export DB_PORT=5438
export DB_USER=postgres
export DB_PASSWORD=postgres
export DB_NAME=skillsphere
export ONTOLOGY_KAFKA_TOPIC=skillsphere.ontology
export ONTOLOGY_TRIPLESTORE_ENDPOINT=http://localhost:7878/graph

GOFLAGS=-mod=mod go run ./cmd/ontologyworker
```

The worker uses the same `config.Load()` routine, so `.env` files work as well. Use systemd, Docker, or a managed queue consumer to keep it running in production.

## Extending Emission to More Services

- **Sessions** – `internal/domain/session/service.Service` converts the persisted session into an `ontology.SessionEvent` and emits it, so scheduling, rescheduling, and cancellations all have graph representations.
- **Matching** – `internal/domain/matching/service.Service` emits `ontology.MatchEvent` envelopes when recommendations are saved. The payload lists algorithm provenance, skill overlaps, and scores for downstream analytics.
- **Future services** – Reuse the pattern: build a domain-specific `XYZEvent` struct under `internal/ontology`, convert the persisted record, and call `Emitter.Emit`. The outbox + worker guarantees delivery to Kafka and the triple store.

## Deployment Notes

- Keep the worker stateless; run N copies to scale throughput. `SELECT ... FOR UPDATE SKIP LOCKED` prevents duplicate deliveries.
- Replace `LogProducer` with your Kafka/Redpanda client (e.g., Segmentio Kafka, Sarama) and point `HTTPTripleStoreClient` at a real ingestion endpoint.
- The SHACL shapes in `ontology/shapes.ttl` can be applied inside the triple store or as part of the worker right before inserting, ensuring the knowledge graph stays valid.
