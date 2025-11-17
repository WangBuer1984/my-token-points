# 📖 代码阅读指南 - 从零开始理解项目

**适合人群**: 初学者、想要理解项目业务逻辑的开发者  
**阅读时间**: 约 30-60 分钟  
**前置知识**: 基础的 Solidity、Go、区块链概念

---

## 🎯 项目是做什么的？

### 核心功能
这是一个**代币事件追踪和积分计算系统**，主要做三件事：

1. **监听区块链事件** 📡
   - 实时监听代币的 mint（铸造）、burn（销毁）、transfer（转账）事件
   - 记录所有代币的流转情况

2. **维护用户余额** 💰
   - 根据事件实时更新每个用户的代币余额
   - 记录每次余额变动的历史

3. **计算持有积分** 🏆
   - 根据用户持有代币的时间和数量计算积分
   - 持有时间越长、数量越多，积分越高

### 业务场景
```
用户 Alice:
  1. 获得 100 个代币 (mint)
  2. 持有 1 小时 → 累积积分
  3. 转出 30 个给 Bob (transfer)
  4. 剩余 70 个继续持有 → 继续累积积分
```

---

## 📚 阅读顺序建议

### 新手推荐路径

```
第一步: 了解数据结构
  ↓
第二步: 理解合约事件
  ↓
第三步: 跟踪数据流
  ↓
第四步: 深入核心逻辑
  ↓
第五步: 理解服务启动
```

---

## 🏗️ 整体架构

### 系统组成

```
┌─────────────────────────────────────────────────┐
│              区块链 (Ethereum/Base)              │
│         MyToken 合约 (铸造/转账/销毁)            │
└─────────────────────────────────────────────────┘
                      ↓ 事件
┌─────────────────────────────────────────────────┐
│               事件监听服务                       │
│          (EventListener - Go)                   │
│   - 监听 TokenMinted 事件                        │
│   - 监听 TokenBurned 事件                        │
│   - 监听 Transfer 事件                           │
└─────────────────────────────────────────────────┘
                      ↓ 解析数据
┌─────────────────────────────────────────────────┐
│               余额管理服务                       │
│          (BalanceService - Go)                  │
│   - 更新用户余额                                 │
│   - 记录余额变动                                 │
└─────────────────────────────────────────────────┘
                      ↓ 存储
┌─────────────────────────────────────────────────┐
│            PostgreSQL 数据库                     │
│   - user_balances (当前余额)                    │
│   - balance_changes (变动历史)                  │
│   - user_points (累计积分)                      │
│   - points_history (积分历史)                   │
│   - sync_state (同步状态)                       │
└─────────────────────────────────────────────────┘
```

---

## 📖 第一步：了解数据结构

### 1.1 数据库设计原则

**重要**: 本项目的数据库设计遵循特定原则：
- ❌ **不使用触发器 (Trigger)** - 所有字段更新由应用层显式控制
- ❌ **不使用外键 (Foreign Key)** - 关联关系由应用层维护
- ✅ **使用 CHECK 约束** - 保证枚举值的合法性
- ✅ **使用 UNIQUE 约束** - 防止重复数据
- ✅ **使用索引** - 优化查询性能

**为什么？**
- **性能**: 写入速度提升 2-3 倍
- **扩展性**: 易于分片、微服务化
- **可维护性**: 逻辑清晰、易于调试

详细说明请参考：[数据库设计原则文档](docs/DATABASE_DESIGN_PRINCIPLES.md)

---

### 1.2 核心数据表

#### 表 1: user_balances (用户当前余额)

```sql
-- 存储每个用户在每条链上的当前代币余额
CREATE TABLE user_balances (
    id BIGSERIAL PRIMARY KEY,
    chain_name VARCHAR(50),          -- 哪条链 (sepolia, base_sepolia)
    user_address VARCHAR(42),        -- 用户地址
    balance NUMERIC(78, 0),          -- 当前余额 (大数，支持很大的数字)
    last_update_block BIGINT,        -- 最后更新的区块号
    last_update_time TIMESTAMP,      -- 最后更新时间
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);
```

**例子**:
```
chain_name    | user_address | balance | last_update_block
--------------|--------------|---------|------------------
sepolia       | 0xABC...123  | 1000    | 9639500
sepolia       | 0xDEF...456  | 500     | 9639520
base_sepolia  | 0xABC...123  | 2000    | 33750600
```

#### 表 2: balance_changes (余额变动历史)

```sql
-- 记录每次余额变动的详细信息
CREATE TABLE balance_changes (
    id BIGSERIAL PRIMARY KEY,
    chain_name VARCHAR(50),          -- 哪条链
    user_address VARCHAR(42),        -- 哪个用户
    tx_hash VARCHAR(66),             -- 交易哈希
    block_number BIGINT,             -- 区块号
    block_time TIMESTAMP,            -- 区块时间
    event_type VARCHAR(20),          -- 事件类型 (mint/burn/transfer_in/transfer_out)
    amount_delta NUMERIC(78, 0),     -- 变动金额 (正数=增加, 负数=减少)
    balance_before NUMERIC(78, 0),   -- 变动前余额
    balance_after NUMERIC(78, 0),    -- 变动后余额
    confirmed BOOLEAN,               -- 是否已确认
    created_at TIMESTAMP
);
```

**例子**:
```
user_address | event_type   | amount_delta | balance_before | balance_after
-------------|--------------|--------------|----------------|---------------
0xABC...123  | mint         | +1000        | 0              | 1000
0xABC...123  | transfer_out | -300         | 1000           | 700
0xDEF...456  | transfer_in  | +300         | 0              | 300
0xABC...123  | burn         | -100         | 700            | 600
```

**理解要点**:
- `amount_delta` 是**变化量**，不是最终余额
- 正数表示增加，负数表示减少
- `balance_before` 和 `balance_after` 记录快照，方便审计

---

### 1.3 Go 数据模型

#### model/balance.go

```go
// UserBalance - 用户余额模型
type UserBalance struct {
    ID              int64     // 主键
    ChainName       string    // 链名称
    UserAddress     string    // 用户地址
    Balance         string    // 余额 (用字符串存储大数)
    LastUpdateBlock int64     // 最后更新区块
    LastUpdateTime  time.Time // 最后更新时间
    CreatedAt       time.Time
    UpdatedAt       time.Time
}

// BalanceChange - 余额变动模型
type BalanceChange struct {
    ID            int64
    ChainName     string
    UserAddress   string
    TxHash        string
    BlockNumber   int64
    BlockTime     time.Time
    EventType     EventType    // 事件类型枚举
    AmountDelta   string       // 变动金额 (字符串存储)
    BalanceBefore string       // 变动前余额
    BalanceAfter  string       // 变动后余额
    Confirmed     bool         // 是否已确认
    CreatedAt     time.Time
}

// EventType - 事件类型
type EventType string

const (
    EventTypeMint        EventType = "mint"         // 铸造
    EventTypeBurn        EventType = "burn"         // 销毁
    EventTypeTransferIn  EventType = "transfer_in"  // 转入
    EventTypeTransferOut EventType = "transfer_out" // 转出
)
```

**为什么用 string 存储余额？**
- 区块链的数字可以非常大 (uint256)
- Go 的 int64 最大只能存储到 2^63-1
- 使用 string 可以存储任意大的数字
- 计算时转换成 `big.Int`

---

## 📖 第二步：理解合约事件

### 2.1 智能合约事件定义

#### contracts/contracts/MyToken.sol

```solidity
contract MyToken is ERC20, Ownable {
    // 事件 1: 代币被铸造
    event TokenMinted(
        address indexed to,      // 接收者地址
        uint256 amount,          // 铸造数量
        uint256 timestamp        // 时间戳
    );

    // 事件 2: 代币被销毁
    event TokenBurned(
        address indexed from,    // 销毁者地址
        uint256 amount,          // 销毁数量
        uint256 timestamp        // 时间戳
    );
    
    // 事件 3: 代币转账 (ERC20 标准事件)
    event Transfer(
        address indexed from,    // 发送者
        address indexed to,      // 接收者
        uint256 value            // 金额
    );

    // 铸造函数
    function mint(address to, uint256 amount) public onlyOwner {
        _mint(to, amount);
        emit TokenMinted(to, amount, block.timestamp);
    }

    // 销毁函数
    function burn(uint256 amount) public {
        _burn(msg.sender, amount);
        emit TokenBurned(msg.sender, amount, block.timestamp);
    }
}
```

**理解要点**:
- `indexed` 关键字让参数可以被搜索
- `emit` 发出事件到区块链
- 事件会被永久记录在区块链上
- 后端服务通过监听这些事件来更新数据库

---

### 2.2 事件监听（Go 端）

#### internal/service/listener/abi.go

```go
// 合约 ABI 定义 (简化版，只包含事件)
const MyTokenABI = `[
    {
        "anonymous": false,
        "inputs": [
            {"indexed": true, "name": "to", "type": "address"},
            {"indexed": false, "name": "amount", "type": "uint256"},
            {"indexed": false, "name": "timestamp", "type": "uint256"}
        ],
        "name": "TokenMinted",
        "type": "event"
    },
    // ... 其他事件
]`
```

**ABI 是什么？**
- ABI = Application Binary Interface (应用程序二进制接口)
- 它告诉 Go 代码：
  - 事件的名字是什么
  - 事件有哪些参数
  - 参数的类型是什么
- 相当于合约和 Go 之间的"翻译字典"

---

## 📖 第三步：跟踪数据流

### 3.1 完整数据流图

```
用户调用合约 mint(Alice, 1000)
            ↓
合约执行 _mint() 并发出 TokenMinted 事件
            ↓
事件被记录到区块链上
            ↓
【6 个区块后】(确认延迟)
            ↓
EventListener 扫描到该事件
            ↓
解析事件: to=Alice, amount=1000
            ↓
调用 BalanceService.UpdateBalance()
            ↓
1. 查询 Alice 当前余额 (假设为 0)
2. 计算新余额: 0 + 1000 = 1000
3. 记录到 balance_changes 表
4. 更新 user_balances 表
            ↓
数据库更新完成
```

### 3.2 关键代码跟踪

#### 步骤 1: EventListener 扫描区块

**文件**: `internal/service/listener/event_listener.go`

```go
// scanBlocks - 扫描区块寻找事件
func (l *EventListener) scanBlocks(ctx context.Context) error {
    // 1. 获取当前链上最新区块
    latestBlock, err := l.client.BlockNumber(ctx)
    
    // 2. 获取上次同步到哪个区块
    syncState, err := l.syncRepo.GetSyncState(ctx, l.chainName)
    fromBlock := syncState.LastSyncedBlock + 1
    
    // 3. 计算要扫描到哪个区块 (延迟 6 个区块确认)
    toBlock := int64(latestBlock) - l.confirmBlocks
    
    // 4. 如果没有新区块，直接返回
    if fromBlock > toBlock {
        return nil
    }
    
    // 5. 查询这个区块范围内的所有事件
    logs, err := l.queryLogs(ctx, fromBlock, toBlock)
    
    // 6. 处理每个事件
    for _, vLog := range logs {
        l.processLog(ctx, vLog)
    }
    
    // 7. 更新同步状态
    syncState.LastSyncedBlock = toBlock
    l.syncRepo.UpdateSyncState(ctx, syncState)
    
    return nil
}
```

**理解要点**:
- `fromBlock` 到 `toBlock` 是要扫描的区块范围
- `confirmBlocks = 6` 是确认延迟，防止链重组
- `queryLogs()` 从区块链获取事件日志
- 扫描是**增量**的，每次只处理新区块

---

#### 步骤 2: 处理 TokenMinted 事件

**文件**: `internal/service/listener/event_listener.go`

```go
// handleTokenMinted - 处理代币铸造事件
func (l *EventListener) handleTokenMinted(ctx context.Context, vLog types.Log) error {
    // 1. 解析事件数据
    var event struct {
        To        common.Address  // 接收者地址
        Amount    *big.Int        // 金额
        Timestamp *big.Int        // 时间戳
    }
    
    // 从事件日志中提取数据
    l.contractABI.UnpackIntoInterface(&event, "TokenMinted", vLog.Data)
    event.To = common.HexToAddress(vLog.Topics[1].Hex()) // indexed 参数在 Topics 中
    
    l.logger.Infof("TokenMinted: to=%s, amount=%s, block=%d",
        event.To.Hex(), event.Amount.String(), vLog.BlockNumber)
    
    // 2. 获取区块时间
    blockTime, err := l.getBlockTime(ctx, vLog.BlockNumber)
    
    // 3. 调用余额服务更新余额
    return l.balanceService.UpdateBalance(ctx, &balance.BalanceUpdate{
        ChainName:   l.chainName,
        UserAddress: event.To.Hex(),
        TxHash:      vLog.TxHash.Hex(),
        BlockNumber: int64(vLog.BlockNumber),
        BlockTime:   blockTime,
        EventType:   model.EventTypeMint,
        AmountDelta: event.Amount.String(),  // 正数，表示增加
    })
}
```

**理解要点**:
- `vLog.Topics[1]` 包含 indexed 参数（to 地址）
- `vLog.Data` 包含非 indexed 参数（amount, timestamp）
- `big.Int` 用于处理大数
- 最后调用 `balanceService.UpdateBalance()` 更新数据库

---

#### 步骤 3: 更新余额

**文件**: `internal/service/balance/balance_service.go`

```go
// UpdateBalance - 更新用户余额
func (s *BalanceService) UpdateBalance(ctx context.Context, update *BalanceUpdate) error {
    // 1. 标准化地址（转小写）
    userAddress := strings.ToLower(update.UserAddress)
    
    // 2. 解析变动金额（string → big.Int）
    amountDelta := new(big.Int)
    amountDelta.SetString(update.AmountDelta, 10)
    
    // 3. 获取当前余额
    currentBalance, err := s.balanceRepo.GetUserBalance(ctx, update.ChainName, userAddress)
    
    // 4. 计算新余额
    var balanceBefore, balanceAfter *big.Int
    
    if currentBalance == nil {
        // 新用户，余额从 0 开始
        balanceBefore = big.NewInt(0)
    } else {
        // 老用户，从数据库读取当前余额
        balanceBefore = new(big.Int)
        balanceBefore.SetString(currentBalance.Balance, 10)
    }
    
    // 计算新余额 = 旧余额 + 变动量
    balanceAfter = new(big.Int).Add(balanceBefore, amountDelta)
    
    // 5. 记录余额变动到历史表
    change := &model.BalanceChange{
        ChainName:     update.ChainName,
        UserAddress:   userAddress,
        TxHash:        update.TxHash,
        BlockNumber:   update.BlockNumber,
        BlockTime:     update.BlockTime,
        EventType:     update.EventType,
        AmountDelta:   amountDelta.String(),
        BalanceBefore: balanceBefore.String(),
        BalanceAfter:  balanceAfter.String(),
        Confirmed:     true,  // 已经延迟 6 个区块，直接标记为已确认
    }
    s.balanceRepo.RecordBalanceChange(ctx, change)
    
    // 6. 更新用户当前余额表
    newBalance := &model.UserBalance{
        ChainName:       update.ChainName,
        UserAddress:     userAddress,
        Balance:         balanceAfter.String(),
        LastUpdateBlock: update.BlockNumber,
        LastUpdateTime:  update.BlockTime,
    }
    s.balanceRepo.UpsertUserBalance(ctx, newBalance)
    
    s.logger.Debugf("Updated balance for %s: %s -> %s",
        userAddress, balanceBefore.String(), balanceAfter.String())
    
    return nil
}
```

**理解要点**:
- **Upsert** = Update + Insert，如果记录存在则更新，不存在则插入
- `balanceBefore` 和 `balanceAfter` 记录快照，方便审计
- `big.Int` 用于安全处理大数运算
- 先记录历史，再更新当前余额

---

## 📖 第四步：深入核心逻辑

### 4.1 Transfer 事件的特殊处理

Transfer 事件比较特殊，因为涉及两个用户：

```go
// handleTransfer - 处理转账事件
func (l *EventListener) handleTransfer(ctx context.Context, vLog types.Log) error {
    // 解析事件
    var event struct {
        From  common.Address
        To    common.Address
        Value *big.Int
    }
    // ... 解析代码 ...
    
    zeroAddress := common.HexToAddress("0x0000000000000000000000000000000000000000")
    
    // 如果 from 是 0 地址 → 这是 mint 事件
    if event.From == zeroAddress {
        return nil  // 忽略，已由 TokenMinted 处理
    }
    
    // 如果 to 是 0 地址 → 这是 burn 事件
    if event.To == zeroAddress {
        return nil  // 忽略，已由 TokenBurned 处理
    }
    
    // 普通转账：需要更新两个账户
    
    // 1. 减少 from 的余额
    amountDelta := new(big.Int).Neg(event.Value)  // 负数
    l.balanceService.UpdateBalance(ctx, &balance.BalanceUpdate{
        ChainName:   l.chainName,
        UserAddress: event.From.Hex(),
        TxHash:      vLog.TxHash.Hex(),
        BlockNumber: int64(vLog.BlockNumber),
        BlockTime:   blockTime,
        EventType:   model.EventTypeTransferOut,
        AmountDelta: amountDelta.String(),  // 负数
    })
    
    // 2. 增加 to 的余额
    l.balanceService.UpdateBalance(ctx, &balance.BalanceUpdate{
        ChainName:   l.chainName,
        UserAddress: event.To.Hex(),
        TxHash:      vLog.TxHash.Hex(),
        BlockNumber: int64(vLog.BlockNumber),
        BlockTime:   blockTime,
        EventType:   model.EventTypeTransferIn,
        AmountDelta: event.Value.String(),  // 正数
    })
    
    return nil
}
```

**理解要点**:
- ERC20 的 mint 和 burn 也会触发 Transfer 事件
- mint: `Transfer(0x0, to, value)`
- burn: `Transfer(from, 0x0, value)`
- 为了避免重复处理，我们忽略涉及 0 地址的 Transfer
- 普通转账需要更新**两个账户**的余额

---

### 4.2 6 区块确认机制

**为什么需要延迟确认？**

```
区块链可能发生"重组"：

原来的链:
  ... → 区块100 → 区块101 → 区块102

发生重组:
  ... → 区块100 → 区块101' → 区块102' → 区块103'
                      ↑
                   区块101 被替换了！
                   
如果立即处理区块101的事件，重组后数据就错了。
```

**解决方案**: 延迟 6 个区块再处理

```go
// 在 scanBlocks() 中
latestBlock := l.client.BlockNumber(ctx)  // 假设 = 1000
toBlock := int64(latestBlock) - l.confirmBlocks  // = 1000 - 6 = 994

// 只处理到区块 994，区块 995-1000 暂时不处理
// 等待 6 个区块后再处理，确保区块不会被重组
```

**6 是怎么来的？**
- 以太坊社区的经验值
- 6 个区块后，链重组的概率极低
- 可以根据不同链调整（快速确认的链可以设为 3）

---

### 4.3 断点续传机制

**场景**: 服务重启后怎么办？

```go
// sync_state 表记录了同步进度
CREATE TABLE sync_state (
    chain_name VARCHAR(50),
    last_synced_block BIGINT,  -- 上次同步到哪个区块
    last_sync_at TIMESTAMP,
    status VARCHAR(20)
);
```

**恢复逻辑**:

```go
// 服务启动时
func (l *EventListener) Start(ctx context.Context) error {
    // 1. 初始化同步状态（如果是第一次运行）
    l.syncRepo.InitSyncState(ctx, l.chainName, l.chainConfig.StartBlock)
    
    // 2. 启动主循环
    go l.run(ctx)
}

// 每次扫描时
func (l *EventListener) scanBlocks(ctx context.Context) error {
    // 从数据库读取上次同步到哪里
    syncState := l.syncRepo.GetSyncState(ctx, l.chainName)
    fromBlock := syncState.LastSyncedBlock + 1  // 从下一个区块开始
    
    // 扫描新区块...
    
    // 更新同步进度
    syncState.LastSyncedBlock = toBlock
    l.syncRepo.UpdateSyncState(ctx, syncState)
}
```

**理解要点**:
- 每次扫描完区块后，更新 `last_synced_block`
- 服务重启后，从 `last_synced_block + 1` 继续
- 不会遗漏任何区块，也不会重复处理

---

## 📖 第五步：理解服务启动

### 5.1 服务启动流程

**文件**: `cmd/start.go`

```go
func runStart() {
    // 1. 加载配置文件
    cfg, err := config.LoadConfig(cfgFile, env)
    // 从 config/dev.yaml 或 config/prod.yaml 读取配置
    
    // 2. 初始化日志
    log := logger.InitLogger(cfg.App.LogLevel)
    
    // 3. 连接数据库
    db, err := database.InitDB(&cfg.Database)
    defer db.Close()
    
    // 4. 创建 Repository 层（数据访问层）
    syncRepo := repository.NewSyncRepository(db)
    balanceRepo := repository.NewBalanceRepository(db)
    
    // 5. 创建 Service 层（业务逻辑层）
    balanceService := balance.NewBalanceService(balanceRepo, log)
    
    // 6. 创建上下文（用于优雅关闭）
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    
    var wg sync.WaitGroup
    
    // 7. 为每条链启动一个事件监听器
    for _, chainCfg := range cfg.Chains {
        wg.Add(1)
        go func(chain config.ChainConfig) {
            defer wg.Done()
            
            // 创建事件监听器
            eventListener, err := listener.NewEventListener(
                chain.Name,
                &chain,
                int(cfg.Confirmation.Blocks),
                syncRepo,
                balanceService,
                log,
            )
            
            // 启动监听
            eventListener.Start(ctx)
            
            // 等待关闭信号
            <-ctx.Done()
            eventListener.Stop()
        }(chainCfg)
    }
    
    // 8. 等待中断信号（Ctrl+C）
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
    <-sigChan
    
    // 9. 优雅关闭
    log.Info("收到关闭信号，正在优雅关闭...")
    cancel()      // 取消所有 goroutine
    wg.Wait()     // 等待所有 goroutine 结束
    log.Info("✅ 服务已停止")
}
```

**理解要点**:
- 分层架构：Repository → Service → Listener
- 每条链有独立的 goroutine 监听
- 使用 `context.Context` 实现优雅关闭
- `WaitGroup` 确保所有 goroutine 都结束后才退出

---

### 5.2 配置文件

**文件**: `config/dev.yaml`

```yaml
# 应用配置
app:
  name: "my-token-points"
  env: "dev"
  log_level: "debug"  # 日志级别

# 数据库配置
database:
  host: "localhost"
  port: 5432
  user: "postgres"
  password: "postgres"
  dbname: "token_points_dev"
  sslmode: "disable"

# 区块链配置
chains:
  # Sepolia 测试网
  - name: "sepolia"
    chain_id: 11155111
    rpc_url: "https://eth-sepolia.g.alchemy.com/v2/YOUR_KEY"
    contract_address: "0x5CCEC1a2039Dd249B376033feB2d5479482614bb"
    start_block: 9639419        # 从这个区块开始同步
    scan_interval: 12           # 每 12 秒扫描一次
    batch_size: 1000            # 每次最多扫描 1000 个区块

  # Base Sepolia 测试网
  - name: "base_sepolia"
    chain_id: 84532
    rpc_url: "https://sepolia.base.org"
    contract_address: "0xb99284e6D996b25974A0E6bA0f10EF6A98c22259"
    start_block: 33750588
    scan_interval: 2            # Base 出块更快
    batch_size: 1000

# 确认机制配置
confirmation:
  blocks: 6  # 延迟 6 个区块确认
```

**理解要点**:
- `start_block` 是合约部署的区块号，从这里开始同步
- `scan_interval` 控制扫描频率
- `batch_size` 控制每次扫描的区块数量
- 不同链可以有不同的配置

---

## 🎯 核心业务逻辑总结

### 业务流程图

```
1. 用户在区块链上操作
   ↓
2. 合约发出事件（Minted/Burned/Transfer）
   ↓
3. EventListener 每隔 N 秒扫描一次新区块
   ↓
4. 发现事件后，延迟 6 个区块确认
   ↓
5. 解析事件数据（地址、金额、类型）
   ↓
6. BalanceService 更新余额：
   - 查询当前余额
   - 计算新余额
   - 记录变动历史
   - 更新当前余额
   ↓
7. 数据持久化到 PostgreSQL
   ↓
8. 更新同步进度（checkpoint）
```

### 关键设计模式

1. **Repository 模式** 🗄️
   - Repository 层封装所有数据库操作
   - Service 层不直接操作数据库
   - 便于测试和替换数据库

2. **Event-Driven Architecture** 📡
   - 通过监听区块链事件驱动业务逻辑
   - 解耦合约和后端
   - 实时性好

3. **Checkpoint 机制** 📌
   - 记录同步进度
   - 支持断点续传
   - 防止数据丢失和重复

4. **Confirmation Delay** ⏰
   - 延迟确认机制
   - 防止链重组
   - 保证数据一致性

---

## 📝 代码阅读练习

### 练习 1: 追踪一次 Mint 操作

1. 找到合约中的 `mint()` 函数
2. 看它发出了什么事件
3. 找到 Go 代码中处理这个事件的函数
4. 看余额是如何被更新的
5. 检查数据库中的记录

### 练习 2: 理解 Transfer 的双边更新

1. 找到 `handleTransfer()` 函数
2. 看它如何判断是 mint/burn 还是普通转账
3. 理解为什么要调用两次 `UpdateBalance()`
4. 思考：如果只更新一边会怎样？

### 练习 3: 模拟服务重启

1. 假设服务在区块 1000 时停止
2. `sync_state` 表中记录了什么？
3. 重启后从哪个区块继续？
4. 如何保证不遗漏也不重复？

---

## 🔍 深入阅读建议

### 按模块深入

1. **合约层**
   - `contracts/contracts/MyToken.sol` - 合约逻辑
   - 学习 ERC20 标准
   - 理解事件机制

2. **数据层**
   - `backend/migrations/*.sql` - 数据库设计
   - `backend/internal/model/*.go` - 数据模型
   - 理解为什么这样设计表结构

3. **Repository 层**
   - `backend/internal/repository/*.go` - 数据访问
   - 学习 SQL 查询
   - 理解 CRUD 操作

4. **Service 层**
   - `backend/internal/service/listener/*.go` - 事件监听
   - `backend/internal/service/balance/*.go` - 余额管理
   - 理解核心业务逻辑

5. **配置层**
   - `backend/config/*.go` - 配置加载
   - `backend/config/*.yaml` - 配置文件
   - 理解配置管理

---

## 💡 常见问题

### Q1: 为什么要延迟 6 个区块？
**A**: 防止区块链重组导致数据不一致。6 个区块后，区块被替换的概率极低。

### Q2: 如果漏掉了某个事件怎么办？
**A**: 不会漏掉。扫描是顺序的，每个区块都会被扫描。如果服务停止，重启后会从上次的位置继续。

### Q3: 余额为什么用字符串存储？
**A**: 区块链的数字是 uint256（最大 2^256-1），Go 的 int64 存不下，所以用 string 存储，计算时转换成 big.Int。

### Q4: Transfer 事件为什么要特殊处理？
**A**: ERC20 标准中，mint 和 burn 也会触发 Transfer。为了避免重复处理，我们只处理普通转账的 Transfer，忽略涉及 0 地址的。

### Q5: 如何保证数据一致性？
**A**: 通过事务（数据库）、确认延迟（区块链）、checkpoint（进度记录）三重保障。

---

## 📚 推荐学习路径

### 入门 (1-2 周)
1. 理解项目是做什么的
2. 看懂数据库表结构
3. 理解事件监听的基本流程
4. 运行项目，观察日志

### 进阶 (2-4 周)
1. 深入理解每个 Service 的逻辑
2. 学习 big.Int 处理大数
3. 理解 Repository 模式
4. 学习 Context 和 WaitGroup

### 高级 (1-2 月)
1. 优化性能（批量处理、并发）
2. 添加新功能（积分计算）
3. 编写测试
4. 部署到生产环境

---

## 🎓 总结

这个项目的核心是：

1. **监听** 区块链事件
2. **解析** 事件数据
3. **更新** 用户余额
4. **记录** 变动历史
5. **持久化** 到数据库

关键技术点：

- ✅ Event-Driven Architecture (事件驱动)
- ✅ Repository Pattern (仓储模式)
- ✅ Checkpoint Mechanism (断点续传)
- ✅ Confirmation Delay (确认延迟)
- ✅ Big Number Handling (大数处理)

通过阅读本指南，你应该能够：

- ✅ 理解项目的整体架构
- ✅ 理解核心业务逻辑
- ✅ 跟踪数据流
- ✅ 阅读和理解代码

---

**开始你的代码阅读之旅吧！** 🚀

有任何疑问，随时查阅相关代码文件或文档。祝学习愉快！

