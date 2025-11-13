package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"connectrpc.com/connect"
	"github.com/gorilla/mux"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	authv1connect "github.com/FACorreiaa/skillsphere-proto/gen/go/auth/v1/authv1connect"

	services2 "github.com/FACorreiaa/skillsphere-api/internal/domain/auth/handler"
	"github.com/FACorreiaa/skillsphere-api/internal/domain/auth/repository"
	handler2 "github.com/FACorreiaa/skillsphere-api/internal/domain/auth/service"
	"github.com/FACorreiaa/skillsphere-api/pkg/interceptors"

	_ "github.com/lib/pq"
)

func main() {
	// Load configuration from environment
	jwtSecret := []byte(os.Getenv("JWT_SECRET"))
	if len(jwtSecret) == 0 {
		log.Fatal("JWT_SECRET environment variable is required")
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	// Initialize database
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test database connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// Initialize OAuth providers (Google and Apple)
	oauthConfig := handler2.LoadOAuthConfigFromEnv()
	if err := handler2.InitOAuth(oauthConfig); err != nil {
		log.Fatalf("Failed to initialize OAuth: %v", err)
	}

	// Initialize repositories
	authRepo := repository.NewAuthRepository(db)

	// Initialize handler
	tokenManager := handler2.NewTokenManager(
		jwtSecret,      // access token secret
		jwtSecret,      // refresh token secret (use different in production!)
		15*time.Minute, // access token TTL
		7*24*time.Hour, // refresh token TTL
	)
	emailService := handler2.NewEmailService()
	authService := services2.NewAuthService(authRepo, tokenManager, emailService)

	// Initialize interceptors
	authInterceptor := interceptors.NewAuthInterceptor(jwtSecret)

	// Create router
	router := mux.NewRouter()

	// Setup OAuth HTTP routes (must be HTTP, not Connect RPC)
	router.HandleFunc("/service/{provider}", authService.HandleOAuthStart).Methods("GET")
	router.HandleFunc("/service/{provider}/callback", authService.HandleOAuthCallback).Methods("GET")

	// Setup Connect RPC handler
	// Public endpoints (no service required)
	publicAuthPath, publicAuthHandler := authv1connect.NewAuthServiceHandler(
		authService,
		connect.WithInterceptors(
		// Add logging, recovery, etc. here
		),
	)

	// Protected endpoints (service required)
	protectedAuthPath, protectedAuthHandler := authv1connect.NewAuthServiceHandler(
		authService,
		connect.WithInterceptors(
			authInterceptor.UnaryInterceptor(),
			// Add other interceptors here
		),
	)

	// Register service
	// For demo, all service endpoints are public
	// In production, you'd split them:
	// - Public: Register, Login, OAuthLogin, RequestPasswordReset, ResetPassword, VerifyEmail
	// - Protected: Logout, RefreshToken, ChangePassword, etc.
	router.Handle(publicAuthPath, publicAuthHandler)

	// Health check endpoint
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	// Wrap with h2c for HTTP/2 support (required for Connect)
	h2cHandler := h2c.NewHandler(router, &http2.Server{})

	// Start server
	addr := ":" + getEnv("PORT", "8080")
	log.Printf("Server listening on %s", addr)
	log.Printf("OAuth endpoints:")
	log.Printf("  - Google: http://localhost%s/auth/google", addr)
	log.Printf("  - Apple: http://localhost%s/auth/apple", addr)
	log.Printf("Connect RPC endpoint: http://localhost%s%s", addr, publicAuthPath)

	if err := http.ListenAndServe(addr, h2cHandler); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
