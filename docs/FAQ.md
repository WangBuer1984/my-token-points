# 常见问题解答 (FAQ)

## 目录

1. [部署相关](#部署相关)
2. [配置相关](#配置相关)
3. [运行相关](#运行相关)
4. [数据相关](#数据相关)
5. [性能相关](#性能相关)
6. [故障排查](#故障排查)

---

## 部署相关

### Q1: 我需要真实的ETH吗？

**A:** 不需要！系统设计用于测试网（Sepolia、Base Sepolia）。你只需要测试网的测试币，可以从水龙头免费获取。

推荐水龙头：
- Sepolia: https://sepoliafaucet.com/
- Base Sepolia: https://www.coinbase.com/faucets/base-ethereum-sepolia-faucet

### Q2: 部署合约时gas费用是多少？

**A:** 在测试网上，部署MyToken合约大约需要 0.01-0.02 测试ETH（取决于网络拥堵情况）。

### Q3: 可以部署到主网吗？

**A:** 可以，但需要谨慎：
1. 修改 `hardhat.config.js` 添加主网配置
2. 确保私钥安全
3. 准备足够的真实ETH支付gas
4. 建议先在测试网充分测试

### Q4: 支持其他EVM链吗？

**A:** 是的！只要是EVM兼容链，都可以支持。添加步骤：
1. 在 `hardhat.config.js` 添加网络配置
2. 在 `backend/.env` 添加该链的RPC和合约地址
3. 系统会自动识别并监听

---

## 配置相关

### Q5: 如何获取RPC URL？

**A:** 有几个选择：
- **公共RPC**: 免费但可能有速率限制
  - Sepolia: https://rpc.sepolia.org
  - Base Sepolia: https://sepolia.base.org
- **私有节点服务**: 更稳定，建议生产环境使用
  - Alchemy: https://www.alchemy.com/
  - Infura: https://infura.io/
  - QuickNode: https://www.quicknode.com/

### Q6: START_BLOCK应该设置多少？

**A:** 建议设置为合约部署时的区块号：
- 可以从部署输出中查看
- 或在区块浏览器查看合约创建交易的区块号
- 设置正确可以避免扫描无用区块，加快初始同步

### Q7: CONFIRMATION_BLOCKS为什么是6？

**A:** 这是以太坊社区的共识：
- 6个区块后，链重组的概率极低
- 确保数据的最终确定性
- 可以根据链的特性调整（如BSC可能设置为15）

### Q8: POINTS_RATE如何计算？

**A:** `POINTS_RATE=0.05` 表示年化5%的积分率：
- 持有 1000 代币一年 → 获得 50 积分
- 持有 1000 代币一小时 → 获得 50 / 8760 ≈ 0.0057 积分
- 可以根据业务需求调整

---

## 运行相关

### Q9: 后端服务占用多少资源？

**A:** 资源占用很小：
- **CPU**: 通常 < 5%（单核）
- **内存**: ~50-100 MB
- **磁盘**: 主要是数据库，取决于事件数量
- **网络**: 每12秒一次RPC调用，带宽需求低

### Q10: 可以同时监听多条链吗？

**A:** 可以！系统原生支持多链：
- 每条链使用独立的goroutine
- 互不影响
- 共享数据库和代码

### Q11: 服务挂了会丢失数据吗？

**A:** 不会！系统设计了容错机制：
- 同步状态持久化到数据库
- 重启后从上次位置继续
- 不会重复处理或遗漏事件

### Q12: 如何停止服务？

**A:** 优雅关闭：
```bash
# 按 Ctrl+C
# 或发送 SIGTERM
kill -TERM <pid>
```
系统会：
- 停止新的区块扫描
- 等待正在进行的事务完成
- 保存同步状态
- 安全退出

---

## 数据相关

### Q13: 数据库会占用多少空间？

**A:** 取决于交易量，估算：
- 每笔交易 ~1KB（余额变动记录）
- 1万笔交易 ~10MB
- 100万笔交易 ~1GB
- 建议定期备份和归档

### Q14: 如何备份数据？

**A:** PostgreSQL备份：
```bash
# 完整备份
pg_dump -U postgres erc20_tracker > backup.sql

# 压缩备份
pg_dump -U postgres erc20_tracker | gzip > backup.sql.gz

# 恢复
psql -U postgres erc20_tracker < backup.sql
```

### Q15: 如何清理旧数据？

**A:** 谨慎操作，建议先备份：
```sql
-- 删除1年前的积分计算记录
DELETE FROM points_calculations 
WHERE calculation_time < NOW() - INTERVAL '1 year';

-- 归档确认的余额变动（不删除）
-- 可以移到归档表
```

### Q16: 余额和链上不一致怎么办？

**A:** 重新同步：
1. 停止服务
2. 执行 `./scripts/reset_sync.sh`
3. 清空相关数据（可选）
4. 重启服务

---

## 性能相关

### Q17: 初始同步很慢怎么办？

**A:** 几个优化方法：
1. **调整批量大小**: 修改 `listener.go` 中的 `batchSize`
2. **使用更快的RPC**: 付费节点通常更快
3. **并行处理**: 代码已实现，确保网络带宽足够
4. **从最近区块开始**: 设置合理的 `START_BLOCK`

### Q18: 数据库查询慢怎么办？

**A:** 检查索引：
```sql
-- 查看当前索引
\di

-- 添加需要的索引（已在schema.sql中）
CREATE INDEX IF NOT EXISTS idx_balance_changes_user_time 
ON balance_changes(user_address, block_timestamp DESC);
```

### Q19: RPC速率限制怎么办？

**A:** 几个方案：
1. **使用付费RPC**: Alchemy/Infura有更高的速率限制
2. **增加请求间隔**: 修改 `listener.go` 的 ticker 间隔
3. **使用多个RPC**: 实现RPC负载均衡（需要自己实现）

---

## 故障排查

### Q20: 无法连接数据库

**A:** 检查清单：
```bash
# 1. PostgreSQL是否运行？
sudo systemctl status postgresql

# 2. 能否本地连接？
psql -U postgres -d erc20_tracker

# 3. 检查配置
cat backend/.env | grep DB_

# 4. 检查权限
# PostgreSQL的 pg_hba.conf 需要允许本地连接
```

### Q21: RPC连接失败

**A:** 排查步骤：
1. 测试RPC是否可达：
```bash
curl -X POST $SEPOLIA_RPC_URL \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}'
```
2. 检查网络连接
3. 尝试备用RPC
4. 查看日志中的详细错误

### Q22: 事件没有被捕获

**A:** 调试步骤：
1. 确认合约地址正确（小写）
2. 确认链ID匹配
3. 检查起始区块号
4. 查看日志中是否有解析错误
5. 在区块浏览器确认事件确实存在

### Q23: 积分计算不准确

**A:** 检查：
1. 余额变动是否都被标记为 `confirmed=true`
```sql
SELECT confirmed, COUNT(*) FROM balance_changes GROUP BY confirmed;
```
2. 检查 `POINTS_RATE` 配置
3. 查看 `points_calculations` 表的历史记录
4. 验证余额变动的完整性

### Q24: 服务启动后立即退出

**A:** 查看日志：
```bash
go run main.go 2>&1 | tee error.log
```
常见原因：
- 数据库连接失败
- RPC URL配置错误
- 合约地址格式错误
- 环境变量未设置

---

## 高级问题

### Q25: 如何添加HTTP API？

**A:** 参考 `docs/API.md`：
1. 添加 Gin 框架依赖
2. 创建 API 路由
3. 基于现有 Repository 实现

### Q26: 如何实现实时通知？

**A:** 几个方案：
1. **WebSocket**: 在事件处理时推送
2. **Server-Sent Events**: 单向实时推送
3. **Webhook**: 调用外部URL
4. **消息队列**: Redis Pub/Sub 或 RabbitMQ

### Q27: 可以用MySQL替代PostgreSQL吗？

**A:** 可以，但需要修改：
1. 数据类型映射（如 NUMERIC → DECIMAL）
2. SQL语法差异（如 ON CONFLICT）
3. 驱动导入（github.com/go-sql-driver/mysql）

### Q28: 如何实现水平扩展？

**A:** 建议：
1. **数据库**: 读写分离，主从复制
2. **服务**: 每条链部署独立实例
3. **缓存**: 使用Redis缓存热点数据
4. **负载均衡**: Nginx分发HTTP请求

### Q29: 如何监控服务健康？

**A:** 实现健康检查：
```go
// 添加HTTP健康检查端点
r.GET("/health", func(c *gin.Context) {
    // 检查数据库连接
    // 检查区块同步延迟
    // 返回状态
})
```

### Q30: 如何实现自动重启？

**A:** 使用systemd：
```ini
[Unit]
Description=ERC20 Tracker
After=network.target

[Service]
Type=simple
User=youruser
WorkingDirectory=/path/to/backend
ExecStart=/path/to/backend/backend
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

---

## 还有问题？

1. 查看项目文档：README.md, DEPLOYMENT.md
2. 查看代码注释
3. 在GitHub提交Issue
4. 加入社区讨论

---

**提示**: 这个FAQ会持续更新。如果你遇到了其他问题并找到了解决方案，欢迎贡献！

