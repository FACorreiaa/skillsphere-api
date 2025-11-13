package api

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/FACorreiaa/skillsphere-api/pkg/config"
	"github.com/FACorreiaa/skillsphere-api/pkg/db"
)

// Dependencies holds all application dependencies
type Dependencies struct {
	Config *config.Config
	DB     *db.DB
	Logger *slog.Logger

	// Repositories
	// Add your repositories here as you implement them
	// AuthRepo *repository.AuthRepository

	// Services
	// Add your handler here as you implement them
	// AuthService *handler.AuthService

	// Handlers
	// Add your service here as you implement them
	// AuthHandler *service.AuthHandler
}

// InitDependencies initializes all application dependencies
func InitDependencies(cfg *config.Config, logger *slog.Logger) (*Dependencies, error) {
	deps := &Dependencies{
		Config: cfg,
		Logger: logger,
	}

	// Initialize database
	if err := deps.initDatabase(); err != nil {
		return nil, fmt.Errorf("failed to init database: %w", err)
	}

	// Initialize repositories
	deps.initRepositories()

	// Initialize handler
	deps.initServices()

	// Initialize service
	deps.initHandlers()

	logger.Info("all dependencies initialized successfully")

	return deps, nil
}

// initDatabase initializes the database connection and runs migrations
func (d *Dependencies) initDatabase() error {
	database, err := db.New(db.Config{
		DSN:             d.Config.Database.DSN(),
		MaxConns:        25,
		MinConns:        5,
		MaxConnLifetime: 5 * time.Minute,
		MaxConnIdleTime: 10 * time.Minute,
	}, d.Logger)
	if err != nil {
		return err
	}

	d.DB = database

	// Run migrations
	if err := d.DB.RunMigrations(); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	d.Logger.Info("database connected and migrations completed successfully")
	return nil
}

// initRepositories initializes all repository layer dependencies
func (d *Dependencies) initRepositories() {
	// Initialize repositories here as you implement them
	// d.AuthRepo = repository.NewAuthRepository(d.DB.Pool)

	d.Logger.Info("repositories initialized")
}

// initServices initializes all service layer dependencies
func (d *Dependencies) initServices() {
	// Initialize handler here as you implement them
	// d.AuthService = handler.NewAuthService(d.AuthRepo, tokenManager, emailService)

	d.Logger.Info("handler initialized")
}

// initHandlers initializes all service dependencies
func (d *Dependencies) initHandlers() {
	// Initialize service here as you implement them
	// d.AuthHandler = service.NewAuthHandler(d.AuthService, d.Logger)

	d.Logger.Info("service initialized")
}

// Cleanup closes all resources
func (d *Dependencies) Cleanup() {
	if d.DB != nil {
		d.DB.Close()
	}
	d.Logger.Info("cleanup completed")
}
