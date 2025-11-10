# 部署指南

本文档提供详细的部署步骤，帮助你从零开始部署整个系统。

## 一、环境准备

### 1.1 安装必要软件

```bash
# Node.js (建议使用 nvm)
curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.0/install.sh | bash
nvm install 18
nvm use 18

# Go
# macOS
brew install go

# Linux
wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin

# PostgreSQL
# macOS
brew install postgresql@15
brew services start postgresql@15

# Linux (Ubuntu/Debian)
sudo apt update
sudo apt install postgresql postgresql-contrib
sudo systemctl start postgresql
```

### 1.2 获取测试网代币

#### Sepolia测试网
- 访问 https://sepoliafaucet.com/
- 或 https://www.alchemy.com/faucets/ethereum-sepolia
- 输入你的钱包地址获取测试ETH

#### Base Sepolia测试网
- 访问 https://www.coinbase.com/faucets/base-ethereum-sepolia-faucet
- 需要先有Sepolia ETH，然后通过桥接到Base Sepolia

## 二、智能合约部署

### 2.1 配置合约环境

```bash
cd contracts

# 安装依赖
npm install

# 创建环境变量文件
cp .envexample .env
```

### 2.2 编辑 `.env` 文件

```bash
nano .env
```

填入以下信息：

```env
# 你的钱包私钥（不要包含0x前缀）
PRIVATE_KEY=your_private_key_without_0x

# RPC URLs (可以使用免费的公共RPC或注册Alchemy/Infura)
SEPOLIA_RPC_URL=https://rpc.sepolia.org
BASE_SEPOLIA_RPC_URL=https://sepolia.base.org

# API Keys (可选，用于合约验证)
ETHERSCAN_API_KEY=your_etherscan_api_key
BASESCAN_API_KEY=your_basescan_api_key
```

### 2.3 部署合约

```bash
# 编译合约
npm run compile

# 部署到Sepolia
npm run deploy:sepolia

# 如果需要，也可以部署到Base Sepolia
npm run deploy:baseSepolia
```

部署成功后，你会看到类似输出：

```
Deploying contracts with the account: 0x...
Account balance: 1000000000000000000
MyToken deployed to: 0xAbC123...
Network: sepolia
ChainId: 11155111
Deployment info saved to: ../backend/deployments/sepolia.json
```

**重要**: 记下合约地址，后面配置后端时需要用到！

### 2.4 测试合约交互

```bash
# 执行一些测试交易（mint, transfer, burn）
npx hardhat run scripts/interact.js --network sepolia
```

你会看到代币的铸造、转账、销毁等操作。这些操作会产生事件，后端会捕获这些事件。

## 三、数据库设置

### 3.1 创建数据库

```bash
# 使用 createdb 命令
createdb erc20_tracker

# 或使用 psql
psql -U postgres
postgres=# CREATE DATABASE erc20_tracker;
postgres=# \q
```

### 3.2 初始化表结构

```bash
# 执行SQL脚本
psql -U postgres -d erc20_tracker -f backend/database/schema.sql

# 或使用 make 命令
make setup-db
```

### 3.3 验证数据库

```bash
psql -U postgres -d erc20_tracker

erc20_tracker=# \dt
                   List of relations
 Schema |         Name         | Type  |  Owner   
--------+----------------------+-------+----------
 public | balance_changes      | table | postgres
 public | points_calculations  | table | postgres
 public | sync_state           | table | postgres
 public | user_balances        | table | postgres
 public | user_points          | table | postgres
(5 rows)

erc20_tracker=# \q
```

## 四、后端服务配置

### 4.1 配置环境变量

```bash
cd backend

# 复制环境变量模板
cp .envexample .env

# 编辑配置文件
nano .env
```

### 4.2 填写 `.env` 配置

```env
# 数据库配置
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=erc20_tracker

# Sepolia 配置
SEPOLIA_RPC_URL=https://rpc.sepolia.org
SEPOLIA_CONTRACT_ADDRESS=0xAbC123...  # 从部署输出中获取
SEPOLIA_START_BLOCK=4500000           # 合约部署的区块号（可选）

# Base Sepolia 配置（如果部署了）
BASE_SEPOLIA_RPC_URL=https://sepolia.base.org
BASE_SEPOLIA_CONTRACT_ADDRESS=0xDeF456...
BASE_SEPOLIA_START_BLOCK=2000000

# 业务配置
CONFIRMATION_BLOCKS=6    # 区块确认数
POINTS_RATE=0.05         # 积分比率（5%年化）
```

**注意事项**：
- `CONTRACT_ADDRESS`: 必须是你部署的合约地址
- `START_BLOCK`: 建议设置为合约部署的区块号，避免扫描无用区块
- 如果只部署了一条链，删除另一条链的配置即可

### 4.3 安装Go依赖

```bash
cd backend
go mod download
```

## 五、启动服务

### 5.1 启动后端服务

```bash
cd backend
go run main.go

# 或使用 make 命令
make run-backend
```

你会看到类似输出：

```
INFO[0000] === ERC20 事件追踪和积分计算系统 ===
INFO[0000] 配置加载成功，监听 1 条链
INFO[0000]   - sepolia (ChainID: 11155111, 合约: 0x...)
INFO[0000] 数据库连接成功
INFO[0000] 启动监听器: sepolia (ChainID: 11155111)
INFO[0000] 启动积分计算定时任务（每小时执行一次）
INFO[0000] 所有服务已启动
INFO[0001] 启动时执行首次积分计算...
INFO[0001] [sepolia] 处理区块: 4500000 - 4501000 (最新: 5000000, 确认: 4999994)
INFO[0001] [sepolia] 发现 5 个事件
INFO[0001] [sepolia] Transfer: from=0x..., to=0x..., amount=100000000000000000000, tx=0x...
...
```

### 5.2 验证服务运行

在另一个终端中，检查数据库中的数据：

```bash
psql -U postgres -d erc20_tracker

-- 查看同步状态
SELECT * FROM sync_state;

-- 查看余额变动
SELECT * FROM balance_changes ORDER BY block_number DESC LIMIT 10;

-- 查看用户余额
SELECT * FROM user_balances;

-- 查看积分
SELECT * FROM user_points;
```

## 六、测试完整流程

### 6.1 执行测试交易

在合约目录执行：

```bash
cd contracts
npx hardhat run scripts/interact.js --network sepolia
```

### 6.2 观察后端日志

你应该能在后端日志中看到：

1. 新区块被扫描
2. 事件被捕获（Transfer、TokenMinted、TokenBurned）
3. 余额被更新
4. 区块确认状态更新

### 6.3 查询数据验证

```sql
-- 查看最近的余额变动
SELECT 
    user_address,
    change_type,
    amount,
    balance_after,
    confirmed,
    block_timestamp
FROM balance_changes 
ORDER BY block_timestamp DESC 
LIMIT 10;

-- 查看用户当前余额
SELECT 
    user_address,
    balance,
    updated_at
FROM user_balances;

-- 查看用户积分
SELECT 
    user_address,
    total_points,
    last_calculated_at
FROM user_points;
```

## 七、生产环境部署建议

### 7.1 使用进程管理器

```bash
# 安装 systemd 服务
sudo nano /etc/systemd/system/erc20-tracker.service
```

```ini
[Unit]
Description=ERC20 Event Tracker
After=network.target postgresql.service

[Service]
Type=simple
User=your_user
WorkingDirectory=/path/to/my-erc/backend
ExecStart=/path/to/my-erc/backend/backend
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

```bash
# 启用并启动服务
sudo systemctl enable erc20-tracker
sudo systemctl start erc20-tracker
sudo systemctl status erc20-tracker
```

### 7.2 配置日志轮转

```bash
sudo nano /etc/logrotate.d/erc20-tracker
```

```
/var/log/erc20-tracker/*.log {
    daily
    rotate 7
    compress
    delaycompress
    missingok
    notifempty
}
```

### 7.3 数据库备份

```bash
# 创建备份脚本
cat > backup_db.sh << 'EOF'
#!/bin/bash
BACKUP_DIR="/path/to/backups"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
pg_dump -U postgres erc20_tracker | gzip > $BACKUP_DIR/erc20_tracker_$TIMESTAMP.sql.gz
# 保留最近7天的备份
find $BACKUP_DIR -name "erc20_tracker_*.sql.gz" -mtime +7 -delete
EOF

chmod +x backup_db.sh

# 添加到 crontab (每天凌晨2点执行)
crontab -e
0 2 * * * /path/to/backup_db.sh
```

### 7.4 监控设置

建议监控以下指标：
- 服务运行状态
- 区块同步延迟
- 数据库连接状态
- 磁盘空间使用
- 内存使用情况

### 7.5 安全建议

1. **防火墙配置**：只开放必要的端口
2. **数据库安全**：使用强密码，限制远程访问
3. **私钥管理**：使用环境变量，不要提交到代码仓库
4. **RPC节点**：使用私有节点或付费服务，避免速率限制
5. **定期更新**：及时更新依赖包和系统补丁

## 八、故障排查

### 8.1 常见问题

#### 问题：无法连接到数据库

```bash
# 检查 PostgreSQL 是否运行
sudo systemctl status postgresql

# 检查连接配置
psql -U postgres -h localhost -p 5432 -d erc20_tracker
```

#### 问题：RPC连接失败

- 检查网络连接
- 验证RPC URL是否正确
- 考虑使用备用RPC节点
- 检查是否有速率限制

#### 问题：事件没有被捕获

- 确认合约地址配置正确
- 检查起始区块号配置
- 查看日志中的错误信息
- 验证合约确实有交易发生

#### 问题：积分计算不准确

- 检查 `POINTS_RATE` 配置
- 确认余额变动已被标记为 `confirmed=true`
- 查看 `points_calculations` 表的历史记录
- 验证余额变动记录的完整性

### 8.2 重新同步

如果数据出现问题，可以重新同步：

```sql
-- 清空数据（谨慎操作！）
TRUNCATE balance_changes, user_balances, user_points, points_calculations;

-- 重置同步状态
UPDATE sync_state SET last_synced_block = 0, last_confirmed_block = 0;

-- 或设置到特定区块
UPDATE sync_state 
SET last_synced_block = 4500000, last_confirmed_block = 4500000
WHERE chain_name = 'sepolia';
```

然后重启服务，系统会从指定区块重新开始同步。

## 九、性能优化

### 9.1 数据库索引

schema.sql 中已经包含了必要的索引，如果查询变慢，可以添加更多索引：

```sql
-- 为常用查询添加组合索引
CREATE INDEX idx_balance_changes_user_time 
ON balance_changes(user_address, block_timestamp DESC);

CREATE INDEX idx_balance_changes_chain_confirmed 
ON balance_changes(chain_name, confirmed, block_number);
```

### 9.2 批处理大小调整

如果历史数据很多，可以调整 `listener.go` 中的批处理大小：

```go
// 默认是 1000，可以根据需要调整
batchSize := uint64(2000)
```

### 9.3 数据库连接池

在 `database/db.go` 中调整连接池大小：

```go
db.SetMaxOpenConns(50)  // 增加最大连接数
db.SetMaxIdleConns(10)  // 增加空闲连接数
```

## 十、下一步

现在你已经成功部署了整个系统！接下来可以：

1. 添加API接口来查询余额和积分
2. 创建前端界面展示数据
3. 添加更多的事件类型
4. 实现通知功能（如余额变动提醒）
5. 添加更复杂的积分规则

祝你使用愉快！如有问题，请查看 README.md 或提交 Issue。

