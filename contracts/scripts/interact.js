const hre = require("hardhat");
const fs = require("fs");
const path = require("path");

async function main() {
  console.log("开始与 MyToken 合约交互...");
  console.log("网络:", hre.network.name);

  // 读取部署信息
  const deploymentFile = path.join(__dirname, "../deployments", `${hre.network.name}.json`);
  if (!fs.existsSync(deploymentFile)) {
    console.error(`❌ 未找到部署信息: ${deploymentFile}`);
    console.log("请先运行部署脚本");
    process.exit(1);
  }

  const deploymentInfo = JSON.parse(fs.readFileSync(deploymentFile, "utf8"));
  const tokenAddress = deploymentInfo.contractAddress;
  console.log("合约地址:", tokenAddress);

  // 获取签名者
  const [owner, user1, user2] = await hre.ethers.getSigners();
  console.log("\n账户信息:");
  console.log("Owner:", owner.address);
  console.log("User1:", user1.address);
  console.log("User2:", user2.address);

  // 连接到合约
  const MyToken = await hre.ethers.getContractFactory("MyToken");
  const token = MyToken.attach(tokenAddress);

  console.log("\n" + "=".repeat(60));
  console.log("测试 1: Mint 代币");
  console.log("=".repeat(60));

  // Mint 代币给 user1
  console.log("\n铸造 100 MTK 给 User1...");
  let tx = await token.mint(user1.address, hre.ethers.parseEther("100"));
  let receipt = await tx.wait();
  console.log("✅ Mint 成功! Tx:", receipt.hash);
  console.log("区块号:", receipt.blockNumber);

  // 检查余额
  let balance = await token.balanceOf(user1.address);
  console.log("User1 余额:", hre.ethers.formatEther(balance), "MTK");

  // Mint 代币给 user2
  console.log("\n铸造 200 MTK 给 User2...");
  tx = await token.mint(user2.address, hre.ethers.parseEther("200"));
  receipt = await tx.wait();
  console.log("✅ Mint 成功! Tx:", receipt.hash);
  console.log("区块号:", receipt.blockNumber);

  balance = await token.balanceOf(user2.address);
  console.log("User2 余额:", hre.ethers.formatEther(balance), "MTK");

  console.log("\n" + "=".repeat(60));
  console.log("测试 2: Transfer 转账");
  console.log("=".repeat(60));

  // User1 转账给 User2
  console.log("\nUser1 转账 30 MTK 给 User2...");
  tx = await token.connect(user1).transfer(user2.address, hre.ethers.parseEther("30"));
  receipt = await tx.wait();
  console.log("✅ 转账成功! Tx:", receipt.hash);
  console.log("区块号:", receipt.blockNumber);

  // 检查余额
  let balance1 = await token.balanceOf(user1.address);
  let balance2 = await token.balanceOf(user2.address);
  console.log("User1 余额:", hre.ethers.formatEther(balance1), "MTK");
  console.log("User2 余额:", hre.ethers.formatEther(balance2), "MTK");

  console.log("\n" + "=".repeat(60));
  console.log("测试 3: Burn 销毁代币");
  console.log("=".repeat(60));

  // User2 销毁代币
  console.log("\nUser2 销毁 50 MTK...");
  tx = await token.connect(user2).burn(hre.ethers.parseEther("50"));
  receipt = await tx.wait();
  console.log("✅ 销毁成功! Tx:", receipt.hash);
  console.log("区块号:", receipt.blockNumber);

  balance2 = await token.balanceOf(user2.address);
  console.log("User2 余额:", hre.ethers.formatEther(balance2), "MTK");

  console.log("\n" + "=".repeat(60));
  console.log("最终余额汇总");
  console.log("=".repeat(60));

  balance1 = await token.balanceOf(user1.address);
  balance2 = await token.balanceOf(user2.address);
  const totalSupply = await token.totalSupply();

  console.log("\nUser1:", hre.ethers.formatEther(balance1), "MTK");
  console.log("User2:", hre.ethers.formatEther(balance2), "MTK");
  console.log("总供应量:", hre.ethers.formatEther(totalSupply), "MTK");

  console.log("\n" + "=".repeat(60));
  console.log("事件日志");
  console.log("=".repeat(60));

  // 获取当前区块
  const latestBlock = await hre.ethers.provider.getBlockNumber();
  const fromBlock = deploymentInfo.blockNumber;
  const blockRange = latestBlock - fromBlock;

  console.log(`\n从区块 ${fromBlock} 到 ${latestBlock} (共 ${blockRange + 1} 个区块)`);

  let allLogs = [];

  // 如果区块范围较大，分批查询（避免 Alchemy 免费套餐限制）
  const MAX_BLOCK_RANGE = 10000; // Alchemy 免费套餐支持的最大范围
  
  if (blockRange <= MAX_BLOCK_RANGE) {
    // 直接查询
    const filter = {
      address: tokenAddress,
      fromBlock: fromBlock,
      toBlock: latestBlock
    };
    allLogs = await hre.ethers.provider.getLogs(filter);
  } else {
    // 分批查询
    console.log(`区块范围较大，将分批查询...`);
    let currentBlock = fromBlock;
    
    while (currentBlock <= latestBlock) {
      const toBlock = Math.min(currentBlock + MAX_BLOCK_RANGE - 1, latestBlock);
      console.log(`  查询区块 ${currentBlock} 到 ${toBlock}...`);
      
      const filter = {
        address: tokenAddress,
        fromBlock: currentBlock,
        toBlock: toBlock
      };
      
      const logs = await hre.ethers.provider.getLogs(filter);
      allLogs = allLogs.concat(logs);
      
      currentBlock = toBlock + 1;
      
      // 添加短暂延迟避免频率限制
      if (currentBlock <= latestBlock) {
        await new Promise(resolve => setTimeout(resolve, 200));
      }
    }
  }

  console.log(`\n找到 ${allLogs.length} 个事件:\n`);

  const logs = allLogs;

  for (let log of logs) {
    try {
      const parsed = token.interface.parseLog(log);
      if (parsed) {
        console.log(`事件: ${parsed.name}`);
        console.log(`  区块: ${log.blockNumber}`);
        console.log(`  交易: ${log.transactionHash}`);
        if (parsed.name === "Transfer") {
          console.log(`  From: ${parsed.args.from}`);
          console.log(`  To: ${parsed.args.to}`);
          console.log(`  Amount: ${hre.ethers.formatEther(parsed.args.value)} MTK`);
        } else if (parsed.name === "TokenMinted") {
          console.log(`  To: ${parsed.args.to}`);
          console.log(`  Amount: ${hre.ethers.formatEther(parsed.args.amount)} MTK`);
          console.log(`  Timestamp: ${parsed.args.timestamp}`);
        } else if (parsed.name === "TokenBurned") {
          console.log(`  From: ${parsed.args.from}`);
          console.log(`  Amount: ${hre.ethers.formatEther(parsed.args.amount)} MTK`);
          console.log(`  Timestamp: ${parsed.args.timestamp}`);
        }
        console.log("");
      }
    } catch (e) {
      // 忽略无法解析的事件
    }
  }

  console.log("✅ 交互测试完成!");
  console.log("\n现在可以启动后端服务来追踪这些事件了");
}

main()
  .then(() => process.exit(0))
  .catch((error) => {
    console.error(error);
    process.exit(1);
  });

