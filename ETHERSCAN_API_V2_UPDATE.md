# Etherscan API V2 更新说明

## 📢 重要变更

**Basescan 已合并到 Etherscan API V2**

这意味着你只需要**一个 Etherscan API Key** 就可以验证所有支持的链（包括 Ethereum、Base、Optimism、Arbitrum 等）的合约。

---

## ✅ 已更新的文件

### 1. `contracts/hardhat.config.js` ⭐
**主要变更**：
- ✅ 统一使用 `ETHERSCAN_API_KEY` 验证所有链
- ✅ 移除了 `BASESCAN_API_KEY` 的引用
- ✅ 添加了 Sourcify 支持（可选的去中心化验证）
- ✅ 更新了 RPC URL 默认值（推荐 Alchemy）

**关键代码**：
```javascript
etherscan: {
  apiKey: {
    sepolia: process.env.ETHERSCAN_API_KEY,
    baseSepolia: process.env.ETHERSCAN_API_KEY,  // 使用同一个 Key
  },
  customChains: [
    {
      network: "baseSepolia",
      chainId: 84532,
      urls: {
        apiURL: "https://api-sepolia.basescan.org/api",
        browserURL: "https://sepolia.basescan.org"
      }
    }
  ]
}
```

### 2. `contracts/env.example` ⭐
**主要变更**：
- ✅ 移除了 `BASESCAN_API_KEY`
- ✅ 添加了详细的注释说明
- ✅ 推荐使用 Alchemy RPC
- ✅ 说明只需要一个 Etherscan API Key

**新的环境变量结构**：
```bash
# 私钥
PRIVATE_KEY=your_private_key_here

# RPC URLs（Alchemy）
SEPOLIA_RPC_URL=https://eth-sepolia.g.alchemy.com/v2/YOUR_ALCHEMY_KEY
BASE_SEPOLIA_RPC_URL=https://sepolia.base.org

# 统一的 Etherscan API Key（用于所有链）
ETHERSCAN_API_KEY=your_etherscan_api_key

# ⚠️ 不再需要 BASESCAN_API_KEY
```

### 3. `backend/config/dev.yaml` ⭐
**主要变更**：
- ✅ 更新了默认 RPC URL（推荐 Alchemy）
- ✅ 添加了区块浏览器配置字段
  - `explorer_url`: 浏览器主页
  - `explorer_api_url`: API 端点

**新增字段**：
```yaml
chains:
  - name: "sepolia"
    chain_id: 11155111
    rpc_url: "https://eth-sepolia.g.alchemy.com/v2/YOUR_KEY"
    # ... 其他配置
    explorer_url: "https://sepolia.etherscan.io"
    explorer_api_url: "https://api-sepolia.etherscan.io/api"

  - name: "base_sepolia"
    chain_id: 84532
    rpc_url: "https://sepolia.base.org"
    # ... 其他配置
    explorer_url: "https://sepolia.basescan.org"
    explorer_api_url: "https://api-sepolia.basescan.org/api"
```

### 4. `backend/config/prod.yaml` ⭐
**主要变更**：
- ✅ 与 dev.yaml 同步更新
- ✅ 添加了区块浏览器配置

### 5. `backend/config/config.go` ⭐
**主要变更**：
- ✅ ChainConfig 结构体添加了新字段：
  - `ExplorerURL string`
  - `ExplorerAPIURL string`

**更新的结构体**：
```go
type ChainConfig struct {
    Name            string
    ChainID         int64
    RPCURL          string
    ContractAddress string
    StartBlock      uint64
    ScanInterval    int
    BatchSize       uint64
    ExplorerURL     string      // 新增
    ExplorerAPIURL  string      // 新增
}
```

---

## 🚀 如何使用更新后的配置

### 第一步：创建 `.env` 文件

```bash
cd contracts
cp env.example .env
```

### 第二步：填写配置

编辑 `contracts/.env`：

```bash
# 1. 从 MetaMask 导出私钥（测试账号）
PRIVATE_KEY=0x你的私钥

# 2. 从 Alchemy 获取 RPC URLs
SEPOLIA_RPC_URL=https://eth-sepolia.g.alchemy.com/v2/你的Alchemy_Key
BASE_SEPOLIA_RPC_URL=https://base-sepolia.g.alchemy.com/v2/你的Base_Key

# 3. 从 Etherscan 申请 API Key（只需要一个）
ETHERSCAN_API_KEY=你的Etherscan_API_Key
```

### 第三步：部署和验证合约

```bash
# 编译
npx hardhat compile

# 部署到 Sepolia
npx hardhat run scripts/deploy.js --network sepolia

# 部署到 Base Sepolia
npx hardhat run scripts/deploy.js --network base_sepolia

# 验证合约（使用同一个 Etherscan Key）
npx hardhat verify --network sepolia 0x你的Sepolia合约地址
npx hardhat verify --network base_sepolia 0x你的Base合约地址
```

---

## 📊 对比：更新前 vs 更新后

### 环境变量对比

| 更新前 | 更新后 |
|--------|--------|
| `ETHERSCAN_API_KEY` | `ETHERSCAN_API_KEY` ✅ |
| `BASESCAN_API_KEY` ❌ | _已移除_ |

### API Key 申请

| 更新前 | 更新后 |
|--------|--------|
| 需要从 2 个网站申请 | 只需要从 1 个网站申请 ✅ |
| Etherscan.io + Basescan.org | 仅 Etherscan.io |

### 配置复杂度

| 更新前 | 更新后 |
|--------|--------|
| 为每条链配置不同的 Key | 所有链使用同一个 Key ✅ |
| 管理多个 API Key | 管理一个 API Key ✅ |

---

## 🔗 相关链接

### 官方文档
- **Etherscan API V2**: https://docs.etherscan.io/v/etherscan-v2/
- **Base on Etherscan**: https://docs.base.org/tools/block-explorers#basescan
- **Hardhat Verify Plugin**: https://hardhat.org/hardhat-runner/plugins/nomicfoundation-hardhat-verify

### 申请 API Keys
- **Etherscan**: https://etherscan.io → Sign In → API Keys → + Add
- **Alchemy**: https://www.alchemy.com → Dashboard → Create App

### 获取测试 ETH
- **Sepolia Faucet**: https://sepoliafaucet.com
- **Base Sepolia Faucet**: https://www.coinbase.com/faucets/base-ethereum-sepolia-faucet

---

## ⚠️ 注意事项

### 1. 不要混淆两种 API Key

| API Key | 用途 | 申请地址 |
|---------|------|----------|
| Alchemy Key | 连接区块链节点（RPC） | Alchemy.com |
| Etherscan Key | 验证合约源码 | Etherscan.io |

**它们是不同的！不能互换使用！**

### 2. 旧的 Basescan API Key 怎么办？

如果你之前申请了 Basescan API Key：
- ✅ 可以继续使用（暂时）
- ✅ 建议迁移到统一的 Etherscan Key
- ⚠️ Basescan 最终会完全废弃

### 3. 配置文件安全

```bash
# ⚠️ 永远不要提交这些文件到 Git
contracts/.env
backend/.env

# ✅ 确保 .gitignore 包含
.env
*.env
!*.env.example
```

---

## ✅ 更新检查清单

部署前确保：

- [ ] ✅ 已更新 `contracts/hardhat.config.js`
- [ ] ✅ 已更新 `contracts/env.example`
- [ ] ✅ 已创建 `contracts/.env` 并填入正确的值
- [ ] ✅ 已从 Etherscan.io 申请 API Key
- [ ] ✅ 已从 Alchemy.com 获取 RPC URLs
- [ ] ✅ 已获取测试 ETH（Sepolia 和 Base Sepolia）
- [ ] ✅ 测试编译成功：`npx hardhat compile`
- [ ] ✅ 准备部署到两条链

---

## 🎯 快速开始

```bash
# 1. 准备环境
cd contracts
cp env.example .env
# 编辑 .env 填入你的 Keys

# 2. 安装依赖
npm install

# 3. 编译合约
npx hardhat compile

# 4. 部署到 Sepolia
npx hardhat run scripts/deploy.js --network sepolia

# 5. 部署到 Base Sepolia
npx hardhat run scripts/deploy.js --network base_sepolia

# 6. 验证合约（等待几个区块后）
npx hardhat verify --network sepolia 0xYourSepoliaAddress
npx hardhat verify --network base_sepolia 0xYourBaseAddress

# 7. 更新后端配置
# 编辑 backend/config/dev.yaml，填入部署的合约地址

# 8. 启动后端服务
cd ../backend
go run main.go start --env dev
```

---

## 📝 更新日志

- **2025-11-15**: 更新配置以支持 Etherscan API V2
  - 移除 BASESCAN_API_KEY
  - 统一使用 ETHERSCAN_API_KEY
  - 添加区块浏览器配置字段
  - 更新默认 RPC URLs

---

## 💬 常见问题

### Q: 为什么要统一 API Key？
A: Etherscan 收购了 Basescan，现在统一管理所有链的区块浏览器。这简化了配置和管理。

### Q: 旧的配置还能用吗？
A: 暂时可以，但建议尽快迁移到新配置，因为 Basescan 的独立 API 最终会废弃。

### Q: 如果验证失败怎么办？
A: 
1. 确保使用了正确的 Etherscan API Key（不是 Alchemy Key）
2. 等待更多区块确认（约 3-5 个区块）
3. 检查编译器版本和优化设置是否匹配

### Q: Sourcify 是什么？
A: Sourcify 是去中心化的合约验证服务，免费且不需要 API Key。在 hardhat.config.js 中已启用。

---

**更新完成！** 🎉

现在你可以使用更简单的配置来部署和验证多链合约了。

