package api

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"my-token-points/internal/service/balance"
	"my-token-points/internal/service/points"
	"my-token-points/internal/service/scheduler"
)

// Handlers API处理器
type Handlers struct {
	balanceService *balance.BalanceService
	pointsService  *points.PointsService
	scheduler      *scheduler.Scheduler
}

// NewHandlers 创建API处理器
func NewHandlers(
	balanceService *balance.BalanceService,
	pointsService *points.PointsService,
	scheduler *scheduler.Scheduler,
) *Handlers {
	return &Handlers{
		balanceService: balanceService,
		pointsService:  pointsService,
		scheduler:      scheduler,
	}
}

// Response 通用响应结构
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// HealthCheckHandler 健康检查
func (h *Handlers) HealthCheckHandler(c *gin.Context) {
	c.JSON(http.StatusOK, Response{
		Success: true,
		Data: gin.H{
			"status":    "healthy",
			"timestamp": time.Now().Unix(),
			"scheduler": h.scheduler.IsRunning(),
		},
	})
}

// GetBalanceHandler 查询用户余额
// GET /api/v1/balance/:chain/:address
func (h *Handlers) GetBalanceHandler(c *gin.Context) {
	chainName := c.Param("chain")
	userAddress := c.Param("address")

	if chainName == "" || userAddress == "" {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Error:   "chain and address are required",
		})
		return
	}

	balance, err := h.balanceService.GetUserBalance(c.Request.Context(), chainName, userAddress)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	if balance == nil {
		c.JSON(http.StatusNotFound, Response{
			Success: false,
			Error:   "balance not found",
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    balance,
	})
}

// GetBalanceChangesHandler 查询余额变动历史
// GET /api/v1/balance/:chain/:address/changes?start_time=xxx&end_time=xxx
func (h *Handlers) GetBalanceChangesHandler(c *gin.Context) {
	chainName := c.Param("chain")
	userAddress := c.Param("address")

	startTimeStr := c.Query("start_time")
	endTimeStr := c.Query("end_time")

	// 默认查询最近24小时
	endTime := time.Now()
	startTime := endTime.Add(-24 * time.Hour)

	if startTimeStr != "" {
		if t, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
			startTime = t
		}
	}

	if endTimeStr != "" {
		if t, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
			endTime = t
		}
	}

	changes, err := h.balanceService.GetBalanceChanges(c.Request.Context(), chainName, userAddress, startTime, endTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    changes,
	})
}

// GetPointsHandler 查询用户积分
// GET /api/v1/points/:chain/:address
func (h *Handlers) GetPointsHandler(c *gin.Context) {
	chainName := c.Param("chain")
	userAddress := c.Param("address")

	if chainName == "" || userAddress == "" {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Error:   "chain and address are required",
		})
		return
	}

	points, err := h.pointsService.GetUserPoints(c.Request.Context(), chainName, userAddress)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	if points == nil {
		c.JSON(http.StatusNotFound, Response{
			Success: false,
			Error:   "points not found",
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    points,
	})
}

// GetPointsHistoryHandler 查询用户积分历史
// GET /api/v1/points/:chain/:address/history?start_time=xxx&end_time=xxx
func (h *Handlers) GetPointsHistoryHandler(c *gin.Context) {
	chainName := c.Param("chain")
	userAddress := c.Param("address")

	startTimeStr := c.Query("start_time")
	endTimeStr := c.Query("end_time")

	// 默认查询最近7天
	endTime := time.Now()
	startTime := endTime.Add(-7 * 24 * time.Hour)

	if startTimeStr != "" {
		if t, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
			startTime = t
		}
	}

	if endTimeStr != "" {
		if t, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
			endTime = t
		}
	}

	history, err := h.pointsService.GetUserPointsHistory(c.Request.Context(), chainName, userAddress, startTime, endTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    history,
	})
}

// GetLeaderboardHandler 查询积分排行榜
// GET /api/v1/leaderboard/:chain?limit=100
func (h *Handlers) GetLeaderboardHandler(c *gin.Context) {
	chainName := c.Param("chain")

	limitStr := c.DefaultQuery("limit", "100")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 1000 {
		limit = 100
	}

	topUsers, err := h.pointsService.GetTopUsers(c.Request.Context(), chainName, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    topUsers,
	})
}

// TriggerCalculationHandler 手动触发积分计算
// POST /api/v1/admin/calculate/:chain
func (h *Handlers) TriggerCalculationHandler(c *gin.Context) {
	chainName := c.Param("chain")

	if chainName == "" {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Error:   "chain is required",
		})
		return
	}

	err := h.scheduler.TriggerCalculation(c.Request.Context(), chainName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    gin.H{"message": "calculation triggered successfully"},
	})
}

// BackfillPointsHandler 执行积分回溯
// POST /api/v1/admin/backfill/:chain
// Body: {"start_time": "2024-01-01T00:00:00Z", "end_time": "2024-01-02T00:00:00Z"}
func (h *Handlers) BackfillPointsHandler(c *gin.Context) {
	chainName := c.Param("chain")

	var req struct {
		StartTime string `json:"start_time" binding:"required"`
		EndTime   string `json:"end_time" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Error:   "invalid request: " + err.Error(),
		})
		return
	}

	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Error:   "invalid start_time format, use RFC3339",
		})
		return
	}

	endTime, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Error:   "invalid end_time format, use RFC3339",
		})
		return
	}

	// 异步执行回溯（可能耗时较长）
	go func() {
		ctx := c.Request.Context()
		if err := h.scheduler.RunBackfill(ctx, chainName, startTime, endTime); err != nil {
			// 记录错误日志（这里简化处理）
			_ = err
		}
	}()

	c.JSON(http.StatusAccepted, Response{
		Success: true,
		Data:    gin.H{"message": "backfill started"},
	})
}

// CORSMiddleware CORS中间件
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// LoggerMiddleware 日志中间件
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		c.Next()

		duration := time.Since(startTime)

		// 记录请求日志（这里简化处理，实际应使用统一的logger）
		_ = duration
	}
}

// RecoveryMiddleware 恢复中间件
func RecoveryMiddleware() gin.HandlerFunc {
	return gin.Recovery()
}

// SetupRoutes 设置路由
func SetupRoutes(router *gin.Engine, handlers *Handlers) {
	// 中间件
	router.Use(CORSMiddleware())
	router.Use(LoggerMiddleware())
	router.Use(RecoveryMiddleware())

	// 健康检查
	router.GET("/health", handlers.HealthCheckHandler)
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": "my-token-points",
			"version": "1.0.0",
			"status":  "running",
		})
	})

	// API v1
	v1 := router.Group("/api/v1")
	{
		// 余额相关
		v1.GET("/balance/:chain/:address", handlers.GetBalanceHandler)
		v1.GET("/balance/:chain/:address/changes", handlers.GetBalanceChangesHandler)

		// 积分相关
		v1.GET("/points/:chain/:address", handlers.GetPointsHandler)
		v1.GET("/points/:chain/:address/history", handlers.GetPointsHistoryHandler)

		// 排行榜
		v1.GET("/leaderboard/:chain", handlers.GetLeaderboardHandler)

		// 管理接口（生产环境应添加认证）
		admin := v1.Group("/admin")
		{
			admin.POST("/calculate/:chain", handlers.TriggerCalculationHandler)
			admin.POST("/backfill/:chain", handlers.BackfillPointsHandler)
		}
	}
}

// NormalizeAddress 标准化地址（转小写）
func NormalizeAddress(address string) string {
	return strings.ToLower(strings.TrimSpace(address))
}

