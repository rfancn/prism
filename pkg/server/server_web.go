package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hdget/sdk"
	panicUtils "github.com/hdget/utils/panic"
	"github.com/rfancn/prism/g"
	"github.com/rfancn/prism/pkg/middleware"
	"github.com/rfancn/prism/pkg/monitor"
	"github.com/rfancn/prism/pkg/proxy"
	"github.com/rfancn/prism/pkg/types"
	"github.com/rfancn/prism/repository"
)

const gracefulShutdownTime = 15 * time.Second

// WebServer implements the Server interface using Gin.
type WebServer struct {
	engine     *gin.Engine
	httpServer *http.Server
	ctx        context.Context
	cancel     context.CancelFunc
	wg         *sync.WaitGroup
}

// New creates a new web server.
func New(address string) (Server, error) {
	// Initialize middlewares first
	middleware.Initialize(&middleware.Config{
		DefaultRPS:   100,
		DefaultBurst: 200,
	})

	// Set Gin mode
	if !g.Debug {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()

	// Create server
	wg := &sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())

	srv := &WebServer{
		engine: engine,
		httpServer: &http.Server{
			Addr:    address,
			Handler: engine,
		},
		ctx:    ctx,
		cancel: cancel,
		wg:     wg,
	}

	// Setup routes
	if err := srv.setupRoutes(); err != nil {
		return nil, err
	}

	return srv, nil
}

// setupRoutes configures all routes and middlewares.
func (s *WebServer) setupRoutes() error {
	// Get middlewares
	loggerMdw, err := middleware.Get(middleware.NameLogger)
	if err != nil {
		return fmt.Errorf("failed to get logger middleware: %w", err)
	}

	authMdw, _ := middleware.Get(middleware.NameAuth)
	ratelimitMdw, _ := middleware.Get(middleware.NameRateLimit)

	// Apply global middlewares
	s.engine.Use(gin.Recovery())
	s.engine.Use(loggerMdw)

	// Health and metrics endpoints (no auth required)
	s.engine.GET("/health", monitor.HealthHandler(cmdVersion))
	s.engine.GET("/ready", monitor.ReadyHandler())
	s.engine.GET("/metrics", monitor.MetricsHandler())

	// Create proxy handler
	proxyHandler := proxy.NewProxyHandler(&types.TargetTLSConfig{})

	// Create route group with auth and rate limit middleware
	apiGroup := s.engine.Group("")
	if authMdw != nil {
		apiGroup.Use(authMdw)
	}
	if ratelimitMdw != nil {
		apiGroup.Use(ratelimitMdw)
	}

	// Load routes from database and register with Gin
	queries := repository.New()
	routes, err := queries.ListEnabledRoutes(context.Background())
	if err != nil {
		return fmt.Errorf("failed to load routes: %w", err)
	}

	for _, route := range routes {
		// Convert pattern: {tenant} -> :tenant
		ginPattern := convertPattern(route.Pattern)

		sdk.Logger().Debug("registering route",
			"id", route.ID,
			"pattern", route.Pattern,
			"gin_pattern", ginPattern,
			"target", route.TargetUrl,
		)

		// Register with Gin - capture route for closure
		routeCopy := route
		apiGroup.Any(ginPattern, proxyHandler.Handler(routeCopy))
	}

	return nil
}

// convertPattern converts {param} syntax to Gin's :param syntax.
// Example: /api/{tenant}/users -> /api/:tenant/users
func convertPattern(pattern string) string {
	re := regexp.MustCompile(`\{(\w+)\}`)
	return re.ReplaceAllString(pattern, ":$1")
}

// Run starts the server and blocks until interrupted.
func (s *WebServer) Run() {
	// Listen for interrupt signals
	chanSignal := make(chan os.Signal, 1)
	signal.Notify(
		chanSignal,
		syscall.SIGINT,  // Ctrl+C
		syscall.SIGTERM, // System termination
	)

	go s.Start()

	for receivedSignal := range chanSignal {
		sdk.Logger().Debug("received signal", "signal", receivedSignal.String())
		switch receivedSignal {
		case syscall.SIGINT, syscall.SIGTERM:
			sdk.Logger().Info("stopping server")
			s.Stop()
			return
		}
	}
}

// Start starts the server in a goroutine.
func (s *WebServer) Start() {
	defer func() {
		if r := recover(); r != nil {
			panicUtils.RecordErrorStack(g.App)
		}
	}()

	s.wg.Add(1)
	defer s.wg.Done()

	sdk.Logger().Info("starting server", "address", s.httpServer.Addr)

	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		sdk.Logger().Fatal("server error", "err", err)
	}
}

// Stop gracefully stops the server.
func (s *WebServer) Stop() {
	// Cancel context
	s.cancel()

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), gracefulShutdownTime)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		sdk.Logger().Error("server shutdown error", "err", err)
	}

	// Wait for goroutines
	s.wg.Wait()

	sdk.Logger().Info("server stopped")
}

// cmdVersion is set during build
var cmdVersion = "dev"
