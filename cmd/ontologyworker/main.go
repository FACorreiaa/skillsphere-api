package main

import (
	"context"
	"database/sql"
	"log/slog"
	"os"
	"os/signal"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/FACorreiaa/skillsphere-api/internal/ontology"
	"github.com/FACorreiaa/skillsphere-api/pkg/config"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{}))
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		logger.Error("load config", "error", err)
		os.Exit(1)
	}

	db, err := sql.Open("pgx", cfg.Database.DSN())
	if err != nil {
		logger.Error("open database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	kafkaTopic := getenv("ONTOLOGY_KAFKA_TOPIC", "skillsphere.ontology")
	kafkaProducer := ontology.NewLogProducer(logger)
	tripleStoreEndpoint := os.Getenv("ONTOLOGY_TRIPLESTORE_ENDPOINT")
	tripleClient := ontology.NewHTTPTripleStoreClient(tripleStoreEndpoint)

	processor := ontology.NewOutboxProcessor(db, kafkaTopic, kafkaProducer, tripleClient)
	if err := processor.Run(ctx, 2*time.Second); err != nil {
		logger.Error("ontology worker exited", "error", err)
		os.Exit(1)
	}
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
