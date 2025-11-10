我# API 文档（待实现）

本文档描述了可以添加到系统中的HTTP API接口。

## 概述

当前版本是纯后台服务，如果需要对外提供数据查询接口，可以添加HTTP服务器。

## 建议的API端点

### 1. 用户余额查询

**GET** `/api/v1/balance/:chain/:address`

查询指定链上用户的当前余额。

**响应示例**：
```json
{
  "chain_name": "sepolia",
  "chain_id": 11155111,
  "user_address": "0x1234...",
  "balance": "1000000000000000000000",
  "balance_formatted": "1000.0",
  "updated_at": "2024-01-15T10:30:00Z"
}
```

### 2. 用户积分查询

**GET** `/api/v1/points/:chain/:address`

查询指定链上用户的累计积分。

**响应示例**：
```json
{
  "chain_name": "sepolia",
  "chain_id": 11155111,
  "user_address": "0x1234...",
  "total_points": 45.678,
  "last_calculated_at": "2024-01-15T10:00:00Z"
}
```

### 3. 余额变动历史

**GET** `/api/v1/history/:chain/:address?limit=50&offset=0`

查询用户的余额变动历史。

**响应示例**：
```json
{
  "chain_name": "sepolia",
  "user_address": "0x1234...",
  "total": 156,
  "items": [
    {
      "id": 1001,
      "change_type": "transfer_in",
      "amount": "100000000000000000000",
      "amount_formatted": "100.0",
      "balance_before": "900000000000000000000",
      "balance_after": "1000000000000000000000",
      "tx_hash": "0xabc...",
      "block_number": 4500123,
      "block_timestamp": "2024-01-15T09:45:30Z",
      "confirmed": true
    }
  ]
}
```

### 4. 积分计算历史

**GET** `/api/v1/points-history/:chain/:address?limit=50&offset=0`

查询用户的积分计算历史。

**响应示例**：
```json
{
  "chain_name": "sepolia",
  "user_address": "0x1234...",
  "total": 720,
  "items": [
    {
      "id": 5001,
      "calculation_time": "2024-01-15T10:00:00Z",
      "balance_snapshot": "1000000000000000000000",
      "balance_formatted": "1000.0",
      "points_earned": 0.0057,
      "created_at": "2024-01-15T10:00:01Z"
    }
  ]
}
```

### 5. 链同步状态

**GET** `/api/v1/sync-status`

查询所有链的同步状态。

**响应示例**：
```json
{
  "chains": [
    {
      "chain_name": "sepolia",
      "chain_id": 11155111,
      "last_synced_block": 5000000,
      "last_confirmed_block": 4999994,
      "sync_lag": 6,
      "updated_at": "2024-01-15T10:30:15Z"
    }
  ]
}
```

### 6. 统计数据

**GET** `/api/v1/stats/:chain`

查询链上的统计数据。

**响应示例**：
```json
{
  "chain_name": "sepolia",
  "chain_id": 11155111,
  "total_users": 1234,
  "total_transactions": 56789,
  "total_supply": "1000000000000000000000000",
  "total_supply_formatted": "1000000.0",
  "total_points_distributed": 12345.67,
  "last_24h_transactions": 89,
  "last_24h_volume": "50000000000000000000000",
  "updated_at": "2024-01-15T10:30:15Z"
}
```

## 实现建议

### 使用 Gin 框架

```go
package main

import (
    "github.com/gin-gonic/gin"
    // ... 其他导入
)

func main() {
    // ... 现有的初始化代码 ...

    // 创建HTTP服务器
    r := gin.Default()
    
    // 注册路由
    api := r.Group("/api/v1")
    {
        api.GET("/balance/:chain/:address", getBalance)
        api.GET("/points/:chain/:address", getPoints)
        api.GET("/history/:chain/:address", getHistory)
        api.GET("/points-history/:chain/:address", getPointsHistory)
        api.GET("/sync-status", getSyncStatus)
        api.GET("/stats/:chain", getStats)
    }
    
    // 启动HTTP服务器
    go r.Run(":8080")
    
    // ... 现有的服务启动代码 ...
}
```

### Handler 示例

```go
func getBalance(c *gin.Context) {
    chain := c.Param("chain")
    address := c.Param("address")
    
    balance, err := repo.GetUserBalance(chain, address)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }
    
    // 格式化余额（假设18位小数）
    balanceFloat := new(big.Float).SetInt(balance)
    balanceFloat.Quo(balanceFloat, big.NewFloat(1e18))
    balanceFormatted, _ := balanceFloat.Float64()
    
    c.JSON(200, gin.H{
        "chain_name": chain,
        "user_address": address,
        "balance": balance.String(),
        "balance_formatted": balanceFormatted,
    })
}
```

## WebSocket 支持

如果需要实时推送数据，可以添加WebSocket支持：

**WS** `/ws/events/:chain/:address`

订阅用户的实时事件。

**消息格式**：
```json
{
  "type": "balance_change",
  "chain_name": "sepolia",
  "user_address": "0x1234...",
  "change_type": "transfer_in",
  "amount": "100000000000000000000",
  "new_balance": "1100000000000000000000",
  "tx_hash": "0xabc...",
  "timestamp": "2024-01-15T10:30:45Z"
}
```

## 安全建议

1. **速率限制**：使用中间件限制API调用频率
2. **认证**：如果需要，添加JWT或API Key认证
3. **CORS**：配置合适的跨域策略
4. **输入验证**：验证所有输入参数（地址格式、链名称等）
5. **缓存**：使用Redis缓存热点数据

## 下一步

要实现这些API，需要：

1. 在 `go.mod` 中添加依赖：
```bash
go get github.com/gin-gonic/gin
```

2. 创建 `api` 包，实现路由和处理器

3. 在 `repository` 中添加相应的查询方法

4. 更新 `main.go` 启动HTTP服务器

