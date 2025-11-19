package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"my-token-points/internal/service/balance"
	"my-token-points/internal/service/points"
	"my-token-points/internal/service/scheduler"
)

// ServerConfig API服务器配置
type ServerConfig struct {
	Host string
	Port int
	Mode string // debug, release, test
}

// Server API服务器
type Server struct {
	config         *ServerConfig
	router         *gin.Engine
	server         *http.Server
	logger         *logrus.Logger
	balanceService *balance.BalanceService
	pointsService  *points.PointsService
	scheduler      *scheduler.Scheduler
}

// NewServer 创建API服务器
func NewServer(
	config *ServerConfig,
	balanceService *balance.BalanceService,
	pointsService *points.PointsService,
	scheduler *scheduler.Scheduler,
	logger *logrus.Logger,
) *Server {
	// 设置Gin模式
	if config.Mode != "" {
		gin.SetMode(config.Mode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建路由
	router := gin.New()

	// 创建处理器
	handlers := NewHandlers(balanceService, pointsService, scheduler)

	// 设置路由
	SetupRoutes(router, handlers)

	return &Server{
		config:         config,
		router:         router,
		logger:         logger,
		balanceService: balanceService,
		pointsService:  pointsService,
		scheduler:      scheduler,
	}
}

// Start 启动API服务器
func (s *Server) Start() error {
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)

	s.server = &http.Server{
		Addr:           addr,
		Handler:        s.router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	s.logger.Infof("Starting API server on %s", addr)

	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}

// Stop 停止API服务器
func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("Stopping API server...")

	if s.server == nil {
		return nil
	}

	if err := s.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown server: %w", err)
	}

	s.logger.Info("API server stopped")
	return nil
}

// GetRouter 获取路由器（用于测试）
func (s *Server) GetRouter() *gin.Engine {
	return s.router
}

