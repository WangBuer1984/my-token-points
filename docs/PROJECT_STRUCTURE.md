# 项目结构说明

```
my-erc/
│
├── README.md                   # 项目主文档
├── QUICKSTART.md              # 快速开始指南
├── DEPLOYMENT.md              # 详细部署指南
├── PROJECT_STRUCTURE.md       # 本文件
├── Makefile                   # 快捷命令
├── .gitignore                 # Git忽略文件
│
├── contracts/                 # 智能合约目录
│   ├── MyToken.sol           # ERC20合约源码
│   ├── hardhat.config.js     # Hardhat配置
│   ├── package.json          # Node.js依赖
│   ├── .envexample           # 环境变量示例
│   ├── scripts/              # 部署和交互脚本
│   │   ├── deploy.js        # 部署脚本
│   │   └── interact.js      # 测试交互脚本
│   ├── artifacts/           # 编译产物（生成）
│   ├── cache/               # Hardhat缓存（生成）
│   └── node_modules/        # Node依赖（生成）
│
├── backend/                  # Go后端服务
│   ├── main.go              # 程序入口
│   ├── go.mod               # Go依赖管理
│   ├── .envexample          # 环境变量示例
│   │
│   ├── config/              # 配置管理
│   │   └── config.go        # 配置加载和管理
│   │
│   ├── database/            # 数据库相关
│   │   ├── db.go           # 数据库连接
│   │   └── schema.sql      # 表结构定义
│   │
│   ├── models/              # 数据模型
│   │   └── models.go        # 实体定义
│   │
│   ├── repository/          # 数据访问层
│   │   ├── repository.go    # CRUD操作
│   │   └── tx_wrapper.go    # 事务包装器
│   │
│   ├── listener/            # 事件监听器
│   │   └── listener.go      # 区块链事件监听
│   │
│   ├── service/             # 业务服务
│   │   └── points_calculator.go  # 积分计算服务
│   │
│   ├── contracts/           # 合约Go绑定
│   │   └── MyToken.go       # 自动生成的合约绑定
│   │
│   └── deployments/         # 部署信息（生成）
│       ├── sepolia.json
│       └── base_sepolia.json
│
├── scripts/                 # 实用脚本
│   ├── check_db.sh         # 数据库检查脚本
│   ├── query_user.sh       # 用户数据查询脚本
│   ├── reset_sync.sh       # 重置同步状态脚本
│   └── generate_abi.sh     # 生成Go绑定脚本
│
└── docs/                    # 文档目录
    ├── API.md              # API接口设计（待实现）
    └── ARCHITECTURE.md     # 系统架构文档
```

## 目录说明

### 根目录

- **README.md**: 项目主要文档，包含功能介绍、使用说明等
- **QUICKSTART.md**: 5分钟快速开始指南
- **DEPLOYMENT.md**: 详细的部署步骤和配置说明
- **Makefile**: 提供便捷的命令（如 `make deploy-sepolia`）
- **.gitignore**: 定义不被Git追踪的文件

### contracts/ - 智能合约

存放Solidity智能合约和相关配置：

- **MyToken.sol**: ERC20代币合约，包含mint/burn功能
- **hardhat.config.js**: Hardhat框架配置，定义网络和编译选项
- **package.json**: Node.js依赖，包含Hardhat和OpenZeppelin
- **scripts/deploy.js**: 自动部署脚本
- **scripts/interact.js**: 测试交互脚本（mint、burn、transfer）

### backend/ - Go后端服务

核心的区块链事件追踪和积分计算服务：

#### 主要文件
- **main.go**: 程序入口，启动所有服务
- **go.mod**: Go模块依赖管理

#### config/ - 配置管理
- 从环境变量加载配置
- 支持多链配置
- 管理数据库连接参数

#### database/ - 数据库
- **db.go**: 数据库连接管理
- **schema.sql**: 数据库表结构SQL

#### models/ - 数据模型
- 定义所有实体结构（余额、积分、变动记录等）

#### repository/ - 数据访问层
- 封装所有数据库操作
- 提供事务支持
- CRUD接口实现

#### listener/ - 事件监听器
- 连接区块链节点
- 监听和解析合约事件
- 实现6区块延迟确认机制
- 更新用户余额

#### service/ - 业务服务
- **points_calculator.go**: 积分计算服务
- 定时任务（每小时）
- 基于余额变化精确计算积分

#### contracts/ - 合约绑定
- Go语言的合约绑定文件
- 用于解析事件和调用合约

### scripts/ - 实用脚本

Shell脚本工具：

- **check_db.sh**: 检查数据库连接和数据统计
- **query_user.sh**: 快速查询用户数据
- **reset_sync.sh**: 重置同步状态（谨慎使用）
- **generate_abi.sh**: 从编译产物生成Go绑定

### docs/ - 文档

详细的技术文档：

- **API.md**: HTTP API接口设计（可选功能）
- **ARCHITECTURE.md**: 系统架构和设计详解

## 配置文件

### contracts/.env
```env
PRIVATE_KEY=...              # 部署账户私钥
SEPOLIA_RPC_URL=...          # RPC节点URL
ETHERSCAN_API_KEY=...        # Etherscan API密钥（可选）
```

### backend/.env
```env
# 数据库配置
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=...
DB_NAME=erc20_tracker

# 区块链配置
SEPOLIA_RPC_URL=...
SEPOLIA_CONTRACT_ADDRESS=... # 部署后的合约地址
SEPOLIA_START_BLOCK=...      # 起始区块号

# 业务配置
CONFIRMATION_BLOCKS=6        # 确认区块数
POINTS_RATE=0.05            # 积分比率
```

## 生成文件

以下目录/文件会在运行过程中自动生成：

- `contracts/artifacts/` - 合约编译产物
- `contracts/cache/` - Hardhat缓存
- `contracts/node_modules/` - Node.js依赖
- `backend/deployments/` - 部署信息JSON
- `backend/backend` - Go编译后的二进制文件

## 数据库表

系统使用5个主要表：

1. **user_balances** - 用户当前余额
2. **user_points** - 用户累计积分
3. **balance_changes** - 余额变动历史记录
4. **points_calculations** - 积分计算历史
5. **sync_state** - 区块同步状态

详细结构请查看 `backend/database/schema.sql`。

## 依赖关系

### 智能合约依赖
```
Node.js 16+ → Hardhat → Solidity 0.8.20 → OpenZeppelin
```

### 后端服务依赖
```
Go 1.21+ → PostgreSQL 13+
├── go-ethereum (Geth)
├── robfig/cron (定时任务)
├── logrus (日志)
└── lib/pq (PostgreSQL驱动)
```

## 数据流

```
区块链 → Listener → Repository → Database
                 ↓
            Points Calculator → Database
```

## 开发流程

1. **修改合约**: 编辑 `contracts/MyToken.sol`
2. **重新编译**: `cd contracts && npm run compile`
3. **重新部署**: `npm run deploy:sepolia`
4. **更新配置**: 修改 `backend/.env` 中的合约地址
5. **重启服务**: 重启Go服务

## 扩展点

### 添加新链
1. 在 `backend/.env` 添加新链配置
2. 重启服务

### 添加API
1. 引入Web框架（如Gin）
2. 在 `backend/` 创建 `api/` 目录
3. 基于 `repository` 实现接口

### 添加新事件
1. 修改合约添加新事件
2. 在 `listener.go` 添加事件处理逻辑
3. 更新数据库schema（如需要）

## 常用命令

```bash
# 合约相关
make install-contracts      # 安装依赖
make compile-contracts      # 编译合约
make deploy-sepolia        # 部署到Sepolia
make interact-sepolia      # 测试交互

# 数据库相关
make setup-db              # 初始化数据库
./scripts/check_db.sh      # 检查数据库
./scripts/query_user.sh sepolia 0x... # 查询用户

# 后端相关
make run-backend           # 运行服务
make build-backend         # 编译服务
```

## 文件大小参考

- **代码总量**: ~3000行
  - Solidity: ~100行
  - Go: ~2500行
  - JavaScript: ~300行
  - SQL: ~100行

- **核心文件**:
  - `listener.go`: ~400行（事件监听核心）
  - `repository.go`: ~250行（数据访问）
  - `points_calculator.go`: ~200行（积分计算）
  - `MyToken.sol`: ~60行（智能合约）

## 更新日志

保持项目更新：
- 定期更新依赖包
- 关注安全漏洞
- 备份数据库
- 监控服务运行状态

---

有关具体实现细节，请查看代码注释和 `docs/ARCHITECTURE.md`。

