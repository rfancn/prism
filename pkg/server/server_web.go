package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hdget/sdk"
	panicUtils "github.com/hdget/utils/panic"
	"github.com/rfancn/prism/g"
	"github.com/rfancn/prism/pkg/middleware"
	"github.com/rfancn/prism/pkg/monitor"
	"github.com/rfancn/prism/pkg/router"
	"github.com/rfancn/prism/pkg/types"
)

const gracefulShutdownTime = 15 * time.Second

// WebServer implements the Server interface using Gin.
type WebServer struct {
	engine      *gin.Engine
	httpServer  *http.Server
	httpsServer *http.Server
	tlsConfig   *types.ServerTLSConfig
	router      *router.Router
	ctx         context.Context
	cancel      context.CancelFunc
	wg          *sync.WaitGroup
}

// New creates a new web server.
func New(address string, tlsConfig *types.ServerTLSConfig) (Server, error) {
	// Initialize middlewares first
	middleware.Initialize()

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

	// Setup HTTPS server if TLS is configured
	if tlsConfig != nil && tlsConfig.Enabled {
		// Verify certificate files exist
		if _, err := os.Stat(tlsConfig.CertFile); os.IsNotExist(err) {
			return nil, fmt.Errorf("TLS certificate file not found: %s", tlsConfig.CertFile)
		}
		if _, err := os.Stat(tlsConfig.KeyFile); os.IsNotExist(err) {
			return nil, fmt.Errorf("TLS key file not found: %s", tlsConfig.KeyFile)
		}

		srv.tlsConfig = tlsConfig
		srv.httpsServer = &http.Server{
			Addr:    address,
			Handler: engine,
		}
		sdk.Logger().Info("TLS server configured", "cert", tlsConfig.CertFile, "key", tlsConfig.KeyFile)
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

	// Apply global middlewares
	s.engine.Use(gin.Recovery())
	s.engine.Use(loggerMdw)

	// Health and metrics endpoints
	s.engine.GET("/health", monitor.HealthHandler(cmdVersion))
	s.engine.GET("/ready", monitor.ReadyHandler())
	s.engine.GET("/metrics", monitor.MetricsHandler())

	// 创建路由管理器
	pluginPaths := g.Config.App.Plugin.Paths
	if len(pluginPaths) == 0 {
		// 默认插件路径
		pluginPaths = []string{"./plugins"}
	}

	s.router, err = router.NewRouter(pluginPaths)
	if err != nil {
		return fmt.Errorf("failed to create router: %w", err)
	}

	// 加载插件
	ctx := context.Background()
	if err := s.router.LoadPlugins(ctx); err != nil {
		sdk.Logger().Warn("failed to load plugins", "err", err)
	}

	// 加载路由配置
	if err := s.router.LoadConfig(ctx); err != nil {
		return fmt.Errorf("failed to load router config: %w", err)
	}

	// 注册路由处理器
	// 使用通配符路由捕获所有请求，由 Router.Handler 处理匹配逻辑
	// NoRoute 只能用于 gin.Engine，不能用于 RouterGroup
	s.engine.NoRoute(s.router.Handler())

	sdk.Logger().Info("routes configured with new router",
		"plugin_paths", pluginPaths,
	)

	return nil
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

	// Start HTTPS server if configured
	if s.httpsServer != nil {
		go s.StartTLS()
	}

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

	sdk.Logger().Info("starting HTTP server", "address", s.httpServer.Addr)

	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		sdk.Logger().Fatal("HTTP server error", "err", err)
	}
}

// StartTLS starts the HTTPS server in a goroutine.
func (s *WebServer) StartTLS() {
	defer func() {
		if r := recover(); r != nil {
			panicUtils.RecordErrorStack(g.App)
		}
	}()

	s.wg.Add(1)
	defer s.wg.Done()

	sdk.Logger().Info("starting HTTPS server", "address", s.httpsServer.Addr)

	if err := s.httpsServer.ListenAndServeTLS(s.tlsConfig.CertFile, s.tlsConfig.KeyFile); err != nil && err != http.ErrServerClosed {
		sdk.Logger().Fatal("HTTPS server error", "err", err)
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
		sdk.Logger().Error("HTTP server shutdown error", "err", err)
	}

	// Shutdown HTTPS server if configured
	if s.httpsServer != nil {
		if err := s.httpsServer.Shutdown(ctx); err != nil {
			sdk.Logger().Error("HTTPS server shutdown error", "err", err)
		}
	}

	// Wait for goroutines
	s.wg.Wait()

	sdk.Logger().Info("server stopped")
}

// GetRouter 获取路由管理器
func (s *WebServer) GetRouter() *router.Router {
	return s.router
}

// cmdVersion is set during build
var cmdVersion = "dev"