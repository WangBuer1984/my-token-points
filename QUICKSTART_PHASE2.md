# 🚀 第二阶段快速开始指南

本指南将帮助你快速启动和使用第二阶段新增的功能：积分计算、定时任务和 API 服务。

---

## 📋 前置条件

确保你已经完成第一阶段的设置：

- ✅ PostgreSQL 数据库已创建并运行迁移
- ✅ 智能合约已部署（Sepolia）
- ✅ 配置文件已更新（`backend/config/dev.yaml`）
- ✅ 数据库中已有一些余额数据（通过事件监听获取）

---

## 🔧 配置检查

### 1. 检查数据库连接

```bash
psql -d token_points_dev -c "\dt"
```

应该看到 5 张表：
- `user_balances`
- `balance_changes`
- `user_points`
- `points_history`
- `sync_state`

### 2. 检查配置文件

编辑 `backend/config/dev.yaml`：

```yaml
# 确保这些配置项存在

# API 服务配置
api:
  enabled: true       # 启用 API 服务
  host: "0.0.0.0"
  port: 8080
  mode: "debug"

# 积分计算配置
points:
  enabled: true                    # 启用积分计算
  cron_expression: "0 * * * *"     # 每小时执行
  hourly_rate: 0.05                # 小时利率 5%
  calc_interval: 3600000000000     # 1小时（纳秒）
  enable_backfill: true            # 启用回溯
  backfill_on_startup: true        # 启动时回溯
  backfill_max_days: 30            # 最多回溯30天
```

### 3. 更新 RPC URL

替换配置文件中的 `YOUR_ALCHEMY_KEY`：

```yaml
chains:
  - name: "sepolia"
    rpc_url: "https://eth-sepolia.g.alchemy.com/v2/你的真实密钥"
    contract_address: "0x5CCEC1a2039Dd249B376033feB2d5479482614bb"
```

---

## 🚀 启动服务

### 方式1：启动所有服务（推荐）

这会同时启动事件监听、积分计算和 API 服务：

```bash
cd backend

# 编译（如果还没编译）
go build -o bin/my-token-points .

# 启动所有服务
./bin/my-token-points start --env dev
```

**预期输出**：
```
正在启动服务...
INFO[0000] 启动 my-token-points 服务，环境: dev
✅ 数据库连接成功
启动事件监听服务...
INFO[0000] 启动 sepolia 链的事件监听...
启动积分计算调度器...
启动API服务 (http://0.0.0.0:8080)...
✅ 所有服务启动完成
📊 API服务地址: http://0.0.0.0:8080
📚 健康检查: http://0.0.0.0:8080/health
⏰ 积分计算调度器已启动
```

### 方式2：分别启动服务（开发调试）

#### 终端1：启动事件监听
```bash
./bin/my-token-points listener --env dev
```

#### 终端2：启动积分计算
```bash
./bin/my-token-points calculator --env dev
```

#### 终端3：启动 API 服务
```bash
./bin/my-token-points api --env dev
```

---

## 🧪 测试功能

### 1. 健康检查

```bash
curl http://localhost:8080/health
```

**预期响应**：
```json
{
  "success": true,
  "data": {
    "status": "healthy",
    "timestamp": 1700388000,
    "scheduler": true
  }
}
```

### 2. 查询用户余额

```bash
# 替换为你的地址
curl http://localhost:8080/api/v1/balance/sepolia/0x你的地址
```

**预期响应**：
```json
{
  "success": true,
  "data": {
    "id": 1,
    "chain_name": "sepolia",
    "user_address": "0x你的地址",
    "balance": "1000000000000000000000",
    "last_update_block": 9639500,
    "last_update_time": "2024-11-19T10:00:00Z",
    "created_at": "2024-11-19T08:00:00Z",
    "updated_at": "2024-11-19T10:00:00Z"
  }
}
```

### 3. 查询余额变动历史

```bash
curl "http://localhost:8080/api/v1/balance/sepolia/0x你的地址/changes?start_time=2024-11-01T00:00:00Z&end_time=2024-11-20T00:00:00Z"
```

### 4. 查询用户积分

```bash
curl http://localhost:8080/api/v1/points/sepolia/0x你的地址
```

**预期响应**：
```json
{
  "success": true,
  "data": {
    "id": 1,
    "chain_name": "sepolia",
    "user_address": "0x你的地址",
    "total_points": 1234.5678,
    "last_calc_at": "2024-11-19T10:00:00Z",
    "created_at": "2024-11-19T08:00:00Z",
    "updated_at": "2024-11-19T10:00:00Z"
  }
}
```

### 5. 查询积分历史

```bash
curl "http://localhost:8080/api/v1/points/sepolia/0x你的地址/history?start_time=2024-11-01T00:00:00Z&end_time=2024-11-20T00:00:00Z"
```

**预期响应**：
```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "chain_name": "sepolia",
      "user_address": "0x你的地址",
      "calc_period_start": "2024-11-19T09:00:00Z",
      "calc_period_end": "2024-11-19T10:00:00Z",
      "balance_snapshot": [
        {
          "balance": "1000000000000000000000",
          "start_time": "2024-11-19T09:00:00Z",
          "end_time": "2024-11-19T10:00:00Z"
        }
      ],
      "points_earned": 50.0,
      "calculation_type": "normal",
      "created_at": "2024-11-19T10:00:05Z"
    }
  ]
}
```

### 6. 查询积分排行榜

```bash
curl "http://localhost:8080/api/v1/leaderboard/sepolia?limit=10"
```

**预期响应**：
```json
{
  "success": true,
  "data": [
    {
      "user_address": "0x1111...",
      "total_points": 10000.0,
      "last_calc_at": "2024-11-19T10:00:00Z"
    },
    {
      "user_address": "0x2222...",
      "total_points": 5000.0,
      "last_calc_at": "2024-11-19T10:00:00Z"
    }
  ]
}
```

---

## 🎯 管理功能

### 手动触发积分计算

如果不想等待定时任务，可以手动触发：

```bash
curl -X POST http://localhost:8080/api/v1/admin/calculate/sepolia
```

**预期响应**：
```json
{
  "success": true,
  "data": {
    "message": "calculation triggered successfully"
  }
}
```

### 执行积分回溯

回溯计算指定时间段的积分：

```bash
curl -X POST http://localhost:8080/api/v1/admin/backfill/sepolia \
  -H "Content-Type: application/json" \
  -d '{
    "start_time": "2024-11-01T00:00:00Z",
    "end_time": "2024-11-19T00:00:00Z"
  }'
```

**预期响应**：
```json
{
  "success": true,
  "data": {
    "message": "backfill started"
  }
}
```

---

## 📊 实际使用场景

### 场景1：查看自己的积分

1. 确保你的地址有一些代币余额（通过合约 mint 或 transfer）
2. 等待事件监听服务同步数据（几分钟）
3. 手动触发积分计算：
   ```bash
   curl -X POST http://localhost:8080/api/v1/admin/calculate/sepolia
   ```
4. 查询你的积分：
   ```bash
   curl http://localhost:8080/api/v1/points/sepolia/0x你的地址
   ```

### 场景2：查看排行榜

```bash
curl "http://localhost:8080/api/v1/leaderboard/sepolia?limit=10"
```

### 场景3：分析积分历史

```bash
curl "http://localhost:8080/api/v1/points/sepolia/0x你的地址/history?start_time=2024-11-01T00:00:00Z&end_time=2024-11-20T00:00:00Z"
```

---

## 🔍 监控和调试

### 查看日志

日志会输出到控制台，使用 `debug` 级别可以看到详细信息：

```
INFO[0000] Starting event listener for sepolia
DEBUG[0001] Scanning blocks from 9639419 to 9639500
DEBUG[0002] Found 5 events
DEBUG[0003] Processing Transfer event...
DEBUG[0004] Updated balance for 0x1234...
INFO[0005] Calculating points for 50 users
DEBUG[0006] Calculated points for 0x1234...: 123.45
```

### 检查数据库

#### 查看用户余额
```sql
SELECT * FROM user_balances WHERE chain_name = 'sepolia' LIMIT 10;
```

#### 查看用户积分
```sql
SELECT * FROM user_points WHERE chain_name = 'sepolia' ORDER BY total_points DESC LIMIT 10;
```

#### 查看积分历史
```sql
SELECT 
  user_address, 
  calc_period_start, 
  calc_period_end, 
  points_earned, 
  calculation_type 
FROM points_history 
WHERE chain_name = 'sepolia' 
ORDER BY calc_period_start DESC 
LIMIT 20;
```

#### 查看同步状态
```sql
SELECT * FROM sync_state;
```

---

## 🐛 常见问题

### 1. API 返回 404

**问题**：`curl http://localhost:8080/api/v1/balance/...` 返回 404

**解决**：
- 检查 API 服务是否启动：`curl http://localhost:8080/health`
- 检查端口是否正确：配置文件中的 `api.port` 是否为 8080
- 检查 URL 格式是否正确

### 2. 积分为 0

**问题**：查询积分返回 0 或 null

**可能原因**：
1. 还没有运行过积分计算
2. 用户没有余额历史
3. 时间还没到整点（定时任务每小时执行）

**解决**：
```bash
# 手动触发计算
curl -X POST http://localhost:8080/api/v1/admin/calculate/sepolia

# 然后再查询
curl http://localhost:8080/api/v1/points/sepolia/0x你的地址
```

### 3. 数据库连接失败

**问题**：`failed to connect to database`

**解决**：
- 检查 PostgreSQL 是否运行：`psql -d token_points_dev`
- 检查配置文件中的数据库连接信息
- 检查密码是否正确

### 4. 编译失败

**问题**：`go build` 报错

**解决**：
```bash
# 清理并重新下载依赖
go clean -modcache
GOPROXY=https://proxy.golang.org,direct go mod tidy
go build -o bin/my-token-points .
```

### 5. 端口被占用

**问题**：`bind: address already in use`

**解决**：
```bash
# 查找占用端口的进程
lsof -i :8080

# 杀死进程
kill -9 <PID>

# 或修改配置文件中的端口
vim config/dev.yaml  # 修改 api.port
```

---

## 📚 API 文档

完整的 API 端点列表：

### 基础接口
- `GET /` - 服务信息
- `GET /health` - 健康检查

### 余额接口
- `GET /api/v1/balance/:chain/:address` - 查询余额
- `GET /api/v1/balance/:chain/:address/changes` - 余额历史

### 积分接口
- `GET /api/v1/points/:chain/:address` - 查询积分
- `GET /api/v1/points/:chain/:address/history` - 积分历史

### 排行榜
- `GET /api/v1/leaderboard/:chain?limit=100` - 积分排行榜

### 管理接口
- `POST /api/v1/admin/calculate/:chain` - 手动触发计算
- `POST /api/v1/admin/backfill/:chain` - 执行回溯计算

---

## 🎓 下一步

1. **测试完整流程**：
   - 从合约 mint 一些代币
   - 等待事件监听同步
   - 手动触发积分计算
   - 查询你的积分

2. **监控定时任务**：
   - 观察每小时的自动计算
   - 检查积分历史记录

3. **尝试回溯功能**：
   - 执行历史数据回溯
   - 验证积分计算的正确性

4. **开发前端界面**（可选）：
   - 使用 React/Vue 创建 UI
   - 调用 API 展示数据
   - 添加图表和可视化

---

## 💡 提示

- 积分计算基于**持有时间**和**余额**，持有越久、余额越多，积分越高
- 默认利率是 **5%/小时**，可以在配置文件中调整
- 积分每小时自动计算一次，也可以手动触发
- 所有查询都支持多链（sepolia, base_sepolia 等）
- API 响应统一格式：`{success, data, error}`

---

**🎉 恭喜！你已经成功启动了第二阶段的所有功能！**

如有问题，请查看：
- [PHASE2_COMPLETE.md](PHASE2_COMPLETE.md) - 完整的功能文档
- [README.md](README.md) - 项目概览
- 日志输出 - 查看详细的运行信息

