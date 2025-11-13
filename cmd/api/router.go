package api

import (
	"net/http"

	"connectrpc.com/connect"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/FACorreiaa/skillsphere-proto/gen/myservice/myserviceconnect"

	"github.com/FACorreiaa/skillsphere-api/pkg/interceptors"
	"github.com/FACorreiaa/skillsphere-api/pkg/observability"
)

// SetupRouter configures all routes and returns the HTTP handler
func SetupRouter(deps *Dependencies) http.Handler {
	mux := http.NewServeMux()

	// Setup interceptor chain
	interceptorChain := connect.WithInterceptors(
		interceptors.NewRecoveryInterceptor(deps.Logger),
		interceptors.NewLoggingInterceptor(deps.Logger),
		interceptors.NewAuthInterceptor(deps.Logger),
		observability.NewMetricsInterceptor(),
	)

	// Register Connect RPC routes
	registerConnectRoutes(mux, deps, interceptorChain)

	// Register health and metrics routes
	registerUtilityRoutes(mux, deps)

	return mux
}

// registerConnectRoutes registers all Connect RPC service handlers
func registerConnectRoutes(mux *http.ServeMux, deps *Dependencies, opts connect.HandlerOption) {
	// MyService routes
	myServicePath, myServiceHandler := myserviceconnect.NewMyServiceHandler(
		deps.MyServiceHandler,
		opts,
	)
	mux.Handle(myServicePath, myServiceHandler)
	deps.Logger.Info("registered Connect RPC service", "path", myServicePath)

	// Add more services here as you implement them:
	// userServicePath, userServiceHandler := userserviceconnect.NewUserServiceHandler(
	//     deps.UserServiceHandler,
	//     opts,
	// )
	// mux.Handle(userServicePath, userServiceHandler)
}

// registerUtilityRoutes registers health check, metrics, and other utility routes
func registerUtilityRoutes(mux *http.ServeMux, deps *Dependencies) {
	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if err := deps.DB.Health(); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("database unhealthy"))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})
	deps.Logger.Info("registered health check", "path", "/health")

	// Readiness check endpoint
	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ready"))
	})
	deps.Logger.Info("registered readiness check", "path", "/ready")

	// Metrics endpoint (Prometheus)
	if deps.Config.Observability.MetricsEnabled {
		mux.Handle("/metrics", promhttp.Handler())
		deps.Logger.Info("registered metrics endpoint", "path", "/metrics")
	}
}
