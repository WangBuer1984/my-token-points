# 快速开始指南

这是一个5分钟快速开始指南，帮助你快速部署和运行系统。

## 前提条件

- ✅ 已安装 Node.js 16+
- ✅ 已安装 Go 1.21+
- ✅ 已安装 PostgreSQL
- ✅ 有一个包含测试币的钱包

## 步骤1: 安装合约依赖

```bash
cd contracts
npm install
```

## 步骤2: 配置合约环境

```bash
cp .envexample .env
nano .env
```

填入你的私钥：
```env
PRIVATE_KEY=your_private_key_without_0x
SEPOLIA_RPC_URL=https://rpc.sepolia.org
```

## 步骤3: 部署合约

```bash
npm run compile
npm run deploy:sepolia
```

**重要**: 记下输出的合约地址！

## 步骤4: 设置数据库

```bash
# 创建数据库
createdb erc20_tracker

# 初始化表结构
psql -d erc20_tracker -f ../backend/database/schema.sql
```

## 步骤5: 配置后端

```bash
cd ../backend
cp .envexample .env
nano .env
```

填入配置：
```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=erc20_tracker

SEPOLIA_RPC_URL=https://rpc.sepolia.org
SEPOLIA_CONTRACT_ADDRESS=0x... # 步骤3的合约地址
SEPOLIA_START_BLOCK=4500000    # 可选
```

## 步骤6: 安装Go依赖并运行

```bash
go mod download
go run main.go
```

## 步骤7: 测试（新终端）

```bash
cd contracts
npx hardhat run scripts/interact.js --network sepolia
```

观察后端日志，你会看到事件被捕获！

## 验证数据

```bash
psql -d erc20_tracker

-- 查看余额变动
SELECT * FROM balance_changes ORDER BY block_number DESC LIMIT 5;

-- 查看用户余额
SELECT * FROM user_balances;

-- 查看同步状态
SELECT * FROM sync_state;
```

## 使用Makefile（可选）

```bash
# 查看所有命令
make help

# 部署合约
make deploy-sepolia

# 设置数据库
make setup-db

# 运行后端
make run-backend
```

## 完成！

现在你的系统正在运行，它会：
- ✅ 自动监听区块链事件
- ✅ 实时更新用户余额
- ✅ 每小时计算积分
- ✅ 6个区块后确认数据

## 下一步

查看完整文档：
- **README.md** - 完整功能介绍
- **DEPLOYMENT.md** - 详细部署指南
- **docs/API.md** - API接口设计

## 遇到问题？

1. 检查日志输出
2. 运行 `./scripts/check_db.sh` 检查数据库
3. 确认合约地址配置正确
4. 验证RPC连接正常

祝你使用愉快！

