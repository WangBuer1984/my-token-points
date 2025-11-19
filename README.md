# My Token Points - 多链代币积分追踪系统

[![Solidity](https://img.shields.io/badge/Solidity-0.8.20-blue)](https://soliditylang.org/)
[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8)](https://go.dev/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-17-336791)](https://www.postgresql.org/)
[![License](https://img.shields.io/badge/License-MIT-green)](LICENSE)

一个功能完整的多链 ERC20 代币事件追踪和积分计算系统，支持 **Sepolia** 和 **Base Sepolia** 测试网。

## 🎯 项目目标

实现一个完整的区块链数据追踪和积分系统：
1. ✅ 部署带 mint 和 burn 功能的 ERC20 合约
2. ✅ 使用 Go 后端追踪合约事件，重建用户余额
3. ✅ 实现 6 区块延迟确认机制
4. ✅ 基于持有时间和余额的精确积分计算
5. ✅ 完整的余额变动历史记录
6. ✅ 支持多链（Sepolia, Base Sepolia）
7. ✅ 支持积分回溯计算（处理服务中断）

## 📊 积分计算示例

```
用户在 15:00 - 0 个 token
用户在 15:10 - 100 个 token (mint)
用户在 15:30 - 200 个 token (transfer in)
在 16:00 计算积分：
  积分 = 100 × 0.05 × (20/60) + 200 × 0.05 × (30/60)
       = 1.667 + 5.0 = 6.667
```

## 🏗️ 技术栈

### 智能合约层
- **Solidity 0.8.20** - 智能合约语言
- **OpenZeppelin** - 安全的合约库
- **Hardhat** - 开发框架

### 后端服务层
- **Go 1.21+** - 后端语言
- **Gin** - Web 框架
- **sqlx** - 数据库工具
- **Cobra** - CLI 框架
- **Viper** - 配置管理（YAML）
- **PostgreSQL 17** - 数据库

### 区块链交互
- **go-ethereum** - 以太坊客户端库
- **Alchemy/Infura** - RPC 节点服务

## 📂 项目结构

```
my-token-points/
├── contracts/              # 智能合约
│   ├── MyToken.sol        # ERC20 合约（带 mint/burn）
│   ├── scripts/           # 部署和测试脚本
│   └── hardhat.config.js  # Hardhat 配置
├── backend/               # Go 后端服务
│   ├── cmd/              # CLI 命令（Cobra）
│   ├── config/           # 配置管理（Viper + YAML）
│   ├── internal/         # 内部包
│   │   ├── model/        # 数据模型
│   │   ├── repository/   # 数据访问层
│   │   ├── service/      # 业务逻辑
│   │   └── api/          # HTTP API
│   └── migrations/       # 数据库迁移
└── docs/                 # 文档
    ├── TECHNICAL_DESIGN.md    # 技术设计文档
    ├── ARCHITECTURE.md         # 系统架构
    └── QUICKSTART.md          # 快速开始
```

## 🚀 快速开始

### 前置要求

- Node.js 18+
- Go 1.21+
- PostgreSQL 17
- MetaMask 钱包

### 1. 部署智能合约

```bash
cd contracts

# 安装依赖
npm install

# 配置环境变量
cp env.example .env
# 编辑 .env 填入：
# - PRIVATE_KEY（从 MetaMask 导出）
# - SEPOLIA_RPC_URL（从 Alchemy 获取）
# - BASE_SEPOLIA_RPC_URL
# - ETHERSCAN_API_KEY（用于验证合约）

# 编译合约
npx hardhat compile

# 部署到 Sepolia
npx hardhat run scripts/deploy.js --network sepolia

# 部署到 Base Sepolia
npx hardhat run scripts/deploy.js --network base_sepolia

# 验证合约（可选）
npx hardhat verify --network sepolia 0x你的合约地址
npx hardhat verify --network base_sepolia 0x你的合约地址
```

### 2. 初始化数据库

```bash
# 创建数据库
createdb token_points_dev

# 运行迁移
psql -d token_points_dev -f backend/migrations/001_init_schema.up.sql
```

### 3. 配置后端

```bash
cd backend

# 编辑配置文件
vim config/dev.yaml
# 填入：
# - 数据库连接信息
# - RPC URLs
# - 合约地址（从部署脚本输出获取）
# - 起始区块号
```

### 4. 启动后端服务

```bash
# 安装依赖
go mod tidy

# 启动所有服务
go run main.go start --env dev
```

## 📖 详细文档

- 📘 [第一阶段完成报告](PHASE1_FINAL_REPORT.md) - 智能合约 + 事件监听
- 📗 [第二阶段完成报告](PHASE2_COMPLETE.md) - 积分计算 + API 服务
- 📕 [快速开始指南](QUICKSTART_PHASE2.md) - 5 分钟上手
- 📙 [数据库设计文档](DATABASE_DESIGN_UPDATE.md) - 数据库架构
- 📝 [Etherscan API V2 更新](ETHERSCAN_API_V2_UPDATE.md) - 最新配置说明

## 🌟 核心特性

### ✅ 智能合约
- ERC20 标准代币
- Mint/Burn 功能
- 自定义事件（TokenMinted, TokenBurned）
- OpenZeppelin 安全库

### ✅ 事件监听
- 实时监听区块链事件
- 批量处理（1000 区块/批）
- 6 区块延迟确认机制
- 断点续传支持

### ✅ 余额重建
- 精确追踪每笔交易
- 记录完整的余额变动历史
- 支持多种变动类型（mint/burn/transfer）
- 余额前后快照

### ✅ 积分计算
- 基于持有时间的精确计算
- 每小时自动计算（定时任务）
- 支持积分回溯（处理中断场景）
- 完整的计算历史审计

### ✅ 多链支持
- 同时支持 Sepolia 和 Base Sepolia
- 通过配置文件轻松添加新链
- 数据通过 chain_name 隔离
- 每条链独立监听和计算

### ✅ API 服务
- RESTful API（Gin 框架）
- 余额查询和历史记录
- 积分查询和历史记录
- 积分排行榜
- 手动触发计算
- 健康检查接口

## 🔧 开发命令

```bash
# 合约相关（在 contracts/ 目录）
npm run compile          # 编译合约
npx hardhat run scripts/deploy.js --network sepolia
npx hardhat run scripts/interact.js --network sepolia

# 后端相关（在 backend/ 目录）
go build -o bin/my-token-points .   # 编译

# 启动服务
./bin/my-token-points start         # 启动所有服务（推荐）
./bin/my-token-points listener      # 仅启动事件监听
./bin/my-token-points calculator    # 仅启动积分计算
./bin/my-token-points api           # 仅启动 API 服务

# API 测试
curl http://localhost:8080/health                        # 健康检查
curl http://localhost:8080/api/v1/points/sepolia/0x...  # 查询积分
curl http://localhost:8080/api/v1/leaderboard/sepolia   # 排行榜
```

## 📊 数据库表

| 表名 | 说明 | 关键字段 |
|------|------|----------|
| user_balances | 用户当前余额 | chain_name, user_address, balance |
| balance_changes | 余额变动历史 | change_type, amount, confirmed |
| user_points | 用户累计积分 | total_points, last_calc_at |
| points_history | 积分计算记录 | balance_snapshot, points_earned |
| sync_state | 区块同步状态 | last_synced_block, status |

## 🔐 安全注意事项

⚠️ **重要提醒**：
- 永远不要提交 `.env` 文件到版本控制
- 只在测试网使用测试账号
- 私钥和 API Keys 妥善保管
- 生产环境使用环境变量管理敏感信息

## 🌐 支持的网络

### 测试网
- ✅ Ethereum Sepolia (ChainID: 11155111)
- ✅ Base Sepolia (ChainID: 84532)

### 主网（计划中）
- 🔮 Ethereum Mainnet
- 🔮 Base Mainnet

## 📈 项目进度

- ✅ 第一阶段：智能合约 + 数据库 + 事件监听 + 余额重建（已完成 100%）
- ✅ 第二阶段：积分计算 + 定时任务 + API 服务（已完成 100%）
- ⏳ 第三阶段：测试 + 监控 + 前端界面（可选）

详见 [PHASE1_FINAL_REPORT.md](PHASE1_FINAL_REPORT.md) 和 [PHASE2_COMPLETE.md](PHASE2_COMPLETE.md)

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📄 许可证

MIT License - 详见 [LICENSE](LICENSE) 文件

## 🔗 相关链接

- [Etherscan (Sepolia)](https://sepolia.etherscan.io)
- [Basescan (Base Sepolia)](https://sepolia.basescan.org)
- [Alchemy](https://www.alchemy.com)
- [Base Documentation](https://docs.base.org)
- [Hardhat Documentation](https://hardhat.org)

---

**开发状态**: 🚧 进行中  
**最后更新**: 2025-11-15
