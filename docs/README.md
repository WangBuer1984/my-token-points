# ERC20 事件追踪和积分计算系统

这是一个完整的区块链事件追踪系统，用于监听ERC20代币的转账、铸造和销毁事件，并基于用户余额自动计算积分。

## 功能特性

### 1. 智能合约
- ✅ 基于OpenZeppelin的ERC20标准实现
- ✅ 支持 `mint` 和 `burn` 功能
- ✅ 发出 `TokenMinted` 和 `TokenBurned` 事件
- ✅ 包含标准的 `Transfer` 事件

### 2. Go后端服务
- ✅ 多链支持（Sepolia、Base Sepolia等）
- ✅ 实时监听合约事件（Transfer、Mint、Burn）
- ✅ 6个区块延迟确认机制，防止区块链回滚
- ✅ 自动重建用户余额
- ✅ 记录所有余额变动历史
- ✅ 每小时自动计算用户积分
- ✅ 基于余额变化的精确积分计算

### 3. 数据库设计
- ✅ 用户总余额表（支持多链）
- ✅ 用户总积分表（支持多链）
- ✅ 余额变动记录表（完整历史）
- ✅ 积分计算历史表
- ✅ 区块同步状态表

## 项目结构

```
my-erc/
├── contracts/                  # 智能合约
│   ├── MyToken.sol            # ERC20合约
│   ├── hardhat.config.js      # Hardhat配置
│   ├── package.json           # Node依赖
│   ├── scripts/
│   │   ├── deploy.js          # 部署脚本
│   │   └── interact.js        # 交互脚本
│   └── .envexample            # 环境变量示例
│
└── backend/                    # Go后端服务
    ├── main.go                # 主程序入口
    ├── go.mod                 # Go依赖
    ├── config/                # 配置管理
    │   └── config.go
    ├── database/              # 数据库
    │   ├── db.go
    │   └── schema.sql         # 数据库表结构
    ├── models/                # 数据模型
    │   └── models.go
    ├── repository/            # 数据访问层
    │   ├── repository.go
    │   └── tx_wrapper.go
    ├── listener/              # 事件监听器
    │   └── listener.go
    ├── service/               # 业务服务
    │   └── points_calculator.go
    ├── contracts/             # 合约Go绑定
    │   └── MyToken.go
    ├── deployments/           # 部署信息
    └── .envexample            # 环境变量示例
```

## 快速开始

### 前置要求

- Node.js 16+ 和 npm/yarn
- Go 1.21+
- PostgreSQL 13+
- 测试网代币（Sepolia ETH 或 Base Sepolia ETH）

### 1. 安装和配置

#### 智能合约部署

```bash
cd contracts

# 安装依赖
npm install

# 复制环境变量文件并填写私钥
cp .envexample .env
# 编辑 .env 文件，填入你的私钥和RPC URL

# 编译合约
npm run compile

# 部署到Sepolia测试网
npm run deploy:sepolia

# 或部署到Base Sepolia测试网
npm run deploy:baseSepolia

# 与合约交互（mint、burn、transfer测试）
npx hardhat run scripts/interact.js --network sepolia
```

#### 数据库设置

```bash
# 创建数据库
createdb erc20_tracker

# 或使用psql
psql -U postgres
CREATE DATABASE erc20_tracker;
\q

# 初始化表结构
psql -U postgres -d erc20_tracker -f backend/database/schema.sql
```

#### 后端服务配置

```bash
cd backend

# 安装Go依赖
go mod download

# 复制环境变量文件
cp .envexample .env

# 编辑 .env 文件，配置：
# - 数据库连接信息
# - RPC URL
# - 合约地址（从部署信息中获取）
# - 起始区块号（可选）
nano .env
```

### 2. 运行服务

```bash
cd backend

# 运行后端服务
go run main.go
```

服务启动后会：
1. 自动连接所有配置的链
2. 从上次同步的区块继续监听（首次从配置的起始区块开始）
3. 监听 Transfer、TokenMinted、TokenBurned 事件
4. 实时更新用户余额
5. 每小时自动计算积分

### 3. 测试功能

在合约目录下运行交互脚本来生成测试事件：

```bash
cd contracts
npx hardhat run scripts/interact.js --network sepolia
```

这个脚本会：
- Mint 代币到测试地址
- 执行 Transfer 转账
- Burn 销毁代币

然后观察后端日志，你会看到：
- 事件被捕获
- 余额被更新
- 区块确认状态更新

## 核心机制说明

### 1. 区块延迟确认

系统默认延迟6个区块后才将余额变动标记为"已确认"。这样做的原因：
- 防止区块链短期分叉导致的数据不一致
- 确保数据的最终一致性
- 未确认的余额变动不会用于积分计算

配置方法：
```env
CONFIRMATION_BLOCKS=6
```

### 2. 积分计算逻辑

积分基于**用户的余额持有时间**计算：

- **公式**: `积分 = 余额 × 持有时间（小时）× 年化比率 / (365 × 24)`
- **年化比率**: 默认 0.05，即持有100代币一年可获得5积分
- **精确计算**: 系统记录所有余额变动，按时间段分段计算积分

例如：
- 用户持有 1000 代币 1 小时，获得积分：
  ```
  1000 × 1 × 0.05 / 8760 ≈ 0.0057 积分
  ```

配置方法：
```env
POINTS_RATE=0.05
```

### 3. 多链支持

系统可以同时监听多条链，每条链的数据独立存储：

```env
# Sepolia配置
SEPOLIA_RPC_URL=https://rpc.sepolia.org
SEPOLIA_CONTRACT_ADDRESS=0x...
SEPOLIA_START_BLOCK=4500000

# Base Sepolia配置
BASE_SEPOLIA_RPC_URL=https://sepolia.base.org
BASE_SEPOLIA_CONTRACT_ADDRESS=0x...
BASE_SEPOLIA_START_BLOCK=2000000
```

## 数据库表说明

### user_balances - 用户余额表
存储每个用户在各条链上的当前余额。

### user_points - 用户积分表
存储每个用户在各条链上的累计积分。

### balance_changes - 余额变动记录表
记录所有余额变动（转入、转出、铸造、销毁），包括：
- 变动类型
- 变动金额
- 变动前后余额
- 交易哈希
- 区块号和时间戳
- 确认状态

### points_calculations - 积分计算历史表
记录每次积分计算的详细信息：
- 计算时间
- 余额快照
- 获得的积分

### sync_state - 同步状态表
记录每条链的同步进度：
- 最后同步区块
- 最后确认区块

## 查询示例

### 查询用户余额
```sql
SELECT * FROM user_balances 
WHERE chain_name = 'sepolia' AND user_address = '0x...';
```

### 查询用户积分
```sql
SELECT * FROM user_points 
WHERE chain_name = 'sepolia' AND user_address = '0x...';
```

### 查询用户余额变动历史
```sql
SELECT * FROM balance_changes 
WHERE chain_name = 'sepolia' AND user_address = '0x...'
ORDER BY block_timestamp DESC;
```

### 查询积分计算历史
```sql
SELECT * FROM points_calculations 
WHERE chain_name = 'sepolia' AND user_address = '0x...'
ORDER BY calculation_time DESC;
```

### 查看同步状态
```sql
SELECT * FROM sync_state;
```

## 监控和日志

系统使用 logrus 进行日志记录，关键日志包括：

- 服务启动和配置信息
- 新区块扫描进度
- 事件捕获详情（Transfer、Mint、Burn）
- 余额更新
- 区块确认状态
- 积分计算结果

日志级别可在 `main.go` 中调整：
```go
log.SetLevel(log.DebugLevel) // 更详细的日志
log.SetLevel(log.InfoLevel)  // 标准日志
```

## 常见问题

### Q1: 如何处理历史区块？

在 `.env` 中设置 `START_BLOCK`：
```env
SEPOLIA_START_BLOCK=4500000
```

系统会从这个区块开始扫描所有历史事件。

### Q2: 如何添加新链？

1. 在 `.env` 中添加新链配置
2. 重启服务
3. 系统会自动创建该链的同步状态

### Q3: 数据不一致怎么办？

1. 停止服务
2. 重置同步状态：
   ```sql
   UPDATE sync_state SET last_synced_block = 0 WHERE chain_name = 'sepolia';
   ```
3. 清空相关数据（可选）
4. 重启服务重新同步

### Q4: 如何优化大量历史数据的扫描速度？

系统已经实现了批量处理，每次扫描最多1000个区块。你可以：
1. 使用更快的RPC节点
2. 调整批量大小（修改 `listener.go` 中的 `batchSize`）
3. 使用归档节点确保历史数据完整

## 安全建议

1. **私钥管理**: 永远不要提交 `.env` 文件到版本控制
2. **数据库安全**: 使用强密码，限制访问权限
3. **RPC节点**: 使用可信的RPC服务或自己运行节点
4. **权限控制**: 合约的 mint 功能只有 owner 可以调用

## 技术栈

- **智能合约**: Solidity 0.8.20, OpenZeppelin, Hardhat
- **后端**: Go 1.21+
- **数据库**: PostgreSQL 13+
- **区块链交互**: go-ethereum (Geth)
- **定时任务**: robfig/cron
- **日志**: logrus

## 许可证

MIT License

## 贡献

欢迎提交 Issue 和 Pull Request！

## 联系方式

如有问题，请提交 GitHub Issue。

