const hre = require("hardhat");
const fs = require("fs");
const path = require("path");

async function main() {
  console.log("开始部署 MyToken 合约...");
  console.log("网络:", hre.network.name);

  // 获取部署账户
  const [deployer] = await hre.ethers.getSigners();
  console.log("部署账户:", deployer.address);

  // 获取账户余额
  const balance = await hre.ethers.provider.getBalance(deployer.address);
  console.log("账户余额:", hre.ethers.formatEther(balance), "ETH");

  // 部署合约
  const MyToken = await hre.ethers.getContractFactory("MyToken");
  console.log("正在部署合约...");
  
  const token = await MyToken.deploy();
  await token.waitForDeployment();

  const tokenAddress = await token.getAddress();
  console.log("✅ MyToken 合约已部署到:", tokenAddress);

  // 获取部署交易信息
  const deployTx = token.deploymentTransaction();
  const deployReceipt = await deployTx.wait();
  
  console.log("部署交易哈希:", deployTx.hash);
  console.log("部署区块号:", deployReceipt.blockNumber);
  console.log("Gas 使用:", deployReceipt.gasUsed.toString());

  // 验证合约信息
  const name = await token.name();
  const symbol = await token.symbol();
  const decimals = await token.decimals();
  const owner = await token.owner();

  console.log("\n合约信息:");
  console.log("名称:", name);
  console.log("符号:", symbol);
  console.log("精度:", decimals);
  console.log("Owner:", owner);

  // 保存部署信息
  const deploymentInfo = {
    network: hre.network.name,
    chainId: hre.network.config.chainId,
    contractAddress: tokenAddress,
    deployerAddress: deployer.address,
    transactionHash: deployTx.hash,
    blockNumber: deployReceipt.blockNumber,
    gasUsed: deployReceipt.gasUsed.toString(),
    timestamp: new Date().toISOString(),
    contractInfo: {
      name,
      symbol,
      decimals: Number(decimals), // 转换 BigInt 为 Number
      owner
    }
  };

  // 创建部署信息目录
  const deploymentsDir = path.join(__dirname, "../deployments");
  if (!fs.existsSync(deploymentsDir)) {
    fs.mkdirSync(deploymentsDir, { recursive: true });
  }

  // 保存到文件
  const filename = `${hre.network.name}.json`;
  const filepath = path.join(deploymentsDir, filename);
  fs.writeFileSync(filepath, JSON.stringify(deploymentInfo, null, 2));
  console.log(`\n部署信息已保存到: ${filepath}`);

  // 验证合约 (仅在支持的网络上)
  if (hre.network.name !== "hardhat" && hre.network.name !== "localhost") {
    console.log("\n等待区块确认后验证合约...");
    console.log("请稍等几个区块后手动执行:");
    console.log(`npx hardhat verify --network ${hre.network.name} ${tokenAddress}`);
  }

  console.log("\n✅ 部署完成!");
  console.log("\n下一步:");
  console.log("1. 复制合约地址到后端配置文件");
  console.log("2. 运行交互脚本测试: npm run interact:" + hre.network.name);
}

main()
  .then(() => process.exit(0))
  .catch((error) => {
    console.error(error);
    process.exit(1);
  });

