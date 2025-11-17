# 🔧 故障排除指南

本文档记录了常见问题和解决方案。

---

## 问题 1: User2 余额不足

### 错误信息
```
ProviderError: insufficient funds for gas * price + value: have 0 want 36172000397892
```

### 原因
测试账户（User1 或 User2）没有 ETH 用于支付 gas 费用。

### 解决方案 A: 使用水龙头获取测试 ETH（推荐）

#### Sepolia 水龙头
- https://sepoliafaucet.com (需要 Alchemy 账户)
- https://www.infura.io/faucet/sepolia (需要 Infura 账户)
- https://faucets.chain.link/sepolia (需要 GitHub 账户)

#### Base Sepolia 水龙头
- https://www.coinbase.com/faucets/base-ethereum-sepolia-faucet
- https://bridge.base.org/

#### 步骤：
1. 在 MetaMask 中切换到需要充值的账户
2. 复制账户地址
3. 访问水龙头网站
4. 粘贴地址并领取测试 ETH
5. 等待 1-2 分钟确认

### 解决方案 B: 从 Owner 账户转账

使用我们提供的转账脚本：

```bash
cd /Users/rick/myweb3/my-token-points/contracts

# Sepolia
npx hardhat run scripts/fund-accounts.js --network sepolia

# Base Sepolia
npx hardhat run scripts/fund-accounts.js --network base_sepolia
```

这个脚本会：
- 检查所有账户余额
- 给余额低于 0.05 ETH 的账户转 0.1 ETH
- 显示最终余额

---

## 问题 2: Alchemy 免费套餐限制

### 错误信息
```
Under the Free tier plan, you can make eth_getLogs requests with up to a 10 block range.
```

### 原因
Alchemy 免费套餐限制 `eth_getLogs` 查询最多 10,000 个区块。

### 解决方案
✅ **已修复**！我已经更新了 `scripts/interact.js`：
- 自动检测区块范围
- 如果超过 10,000 个区块，会自动分批查询
- 每批之间添加延迟避免频率限制

现在可以正常运行：
```bash
npx hardhat run scripts/interact.js --network sepolia
```

### 如果仍有问题

**选项 A: 升级到 Alchemy 付费套餐**
- 访问 https://www.alchemy.com/pricing
- Growth 套餐支持更大的区块范围

**选项 B: 使用其他 RPC 提供商**
在 `.env` 中更改 RPC URL：
```bash
# Infura
SEPOLIA_RPC_URL=https://sepolia.infura.io/v3/YOUR_INFURA_KEY

# Ankr (免费)
SEPOLIA_RPC_URL=https://rpc.ankr.com/eth_sepolia

# 公共节点（不推荐用于生产）
SEPOLIA_RPC_URL=https://ethereum-sepolia-rpc.publicnode.com
```

---

## 问题 3: 只有一个账户无法运行 interact.js

### 错误信息
```
TypeError: Cannot read properties of undefined (reading 'address')
```

### 原因
`interact.js` 脚本需要至少 3 个账户进行完整测试。

### 解决方案

#### 选项 A: 配置多个测试账户（推荐）

参考 [MULTI_ACCOUNT_SETUP.md](MULTI_ACCOUNT_SETUP.md) 配置指南：

1. 从 MetaMask 导出 2-3 个账户的私钥
2. 在 `.env` 中添加：
   ```bash
   PRIVATE_KEY=0x主账户私钥
   PRIVATE_KEY_USER1=0x测试账户1私钥
   PRIVATE_KEY_USER2=0x测试账户2私钥
   ```
3. 给测试账户充值 0.1-0.5 ETH
4. 运行验证脚本：
   ```bash
   npx hardhat run scripts/test-accounts.js --network sepolia
   ```

#### 选项 B: 修改脚本使用单账户模式

如果只想快速测试，可以修改 `interact.js` 只使用一个账户：

```javascript
// 修改获取账户部分
const [owner] = await hre.ethers.getSigners();  // 只获取一个账户
console.log("\n账户信息:");
console.log("Owner:", owner.address);

// 后续所有 mint、transfer、burn 都使用 owner 账户
// 示例：
await token.mint(owner.address, hre.ethers.parseEther("1000"));
await token.burn(hre.ethers.parseEther("100"));
```

---

## 问题 4: npm EPERM 权限错误

### 错误信息
```
npm error code EPERM
npm error syscall open
npm error errno -1
```

### 原因
沙箱环境限制了对某些系统文件的访问。

### 解决方案

**直接使用 npx 命令**（推荐）：
```bash
npx hardhat run scripts/interact.js --network sepolia
```

而不是通过 npm scripts 运行。

---

## 问题 5: 合约验证失败

### 错误信息
```
You are using a deprecated V1 endpoint
```

### 说明
这只是一个警告，不影响功能。合约已在 Sourcify 上成功验证。

### 解决方案

如果你想消除警告，可以等待 Hardhat 插件更新，或者只使用 Sourcify 验证（已自动启用）。

查看验证结果：
- Sourcify: https://repo.sourcify.dev/contracts/full_match/{chainId}/{address}/
- Etherscan: https://sepolia.etherscan.io/address/{address}#code

---

## 问题 6: 部署后无法读取合约信息

### 错误信息
```
Error: could not decode result data (value="0x", ...)
```

### 原因
合约刚部署完，RPC 节点可能还没完全同步状态。

### 解决方案

**已修复**！部署脚本已添加错误处理。这个错误不影响部署成功。

验证合约是否真的部署成功：
```bash
# 方法 1: 在区块浏览器查看
https://sepolia.etherscan.io/address/你的合约地址

# 方法 2: 使用 Hardhat Console
npx hardhat console --network sepolia
> const MyToken = await ethers.getContractFactory("MyToken");
> const token = MyToken.attach("你的合约地址");
> await token.name();  // 应该返回 "MyToken"
```

---

## 问题 7: RPC 连接失败

### 错误信息
```
Error: could not detect network
```

### 原因
RPC URL 配置错误或 API Key 无效。

### 解决方案

1. **检查 `.env` 配置**：
   ```bash
   cat contracts/.env | grep RPC_URL
   ```

2. **测试 RPC 连接**：
   ```bash
   curl -X POST YOUR_RPC_URL \
     -H "Content-Type: application/json" \
     -d '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}'
   ```

3. **使用备用 RPC**：
   ```bash
   # Sepolia 公共节点
   SEPOLIA_RPC_URL=https://ethereum-sepolia-rpc.publicnode.com
   
   # Base Sepolia 官方节点
   BASE_SEPOLIA_RPC_URL=https://sepolia.base.org
   ```

---

## 🆘 获取帮助

如果问题仍未解决：

1. **查看完整错误日志**：
   ```bash
   npx hardhat run scripts/interact.js --network sepolia 2>&1 | tee error.log
   ```

2. **检查账户配置**：
   ```bash
   npx hardhat run scripts/test-accounts.js --network sepolia
   ```

3. **验证环境变量**：
   ```bash
   # 确保 .env 文件存在且配置正确
   ls -la contracts/.env
   cat contracts/.env
   ```

4. **查看相关文档**：
   - [MULTI_ACCOUNT_SETUP.md](MULTI_ACCOUNT_SETUP.md) - 多账户配置
   - [SETUP_COMPLETE.md](../SETUP_COMPLETE.md) - 环境配置
   - [README.md](../README.md) - 项目主页

---

## 📝 快速诊断命令

运行这些命令快速诊断问题：

```bash
cd /Users/rick/myweb3/my-token-points/contracts

# 1. 检查账户配置
npx hardhat run scripts/test-accounts.js --network sepolia

# 2. 检查部署信息
cat deployments/sepolia.json

# 3. 给账户充值
npx hardhat run scripts/fund-accounts.js --network sepolia

# 4. 运行完整测试
npx hardhat run scripts/interact.js --network sepolia

# 5. 使用 Hardhat Console 手动测试
npx hardhat console --network sepolia
```

---

**最后更新**: 2025-11-16  
**适用版本**: Hardhat ^2.19.0, Ethers ^6.15.0

