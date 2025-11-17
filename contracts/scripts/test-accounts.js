const hre = require("hardhat");

async function main() {
  console.log("=".repeat(60));
  console.log("测试账户配置验证");
  console.log("=".repeat(60));
  console.log("\n网络:", hre.network.name);
  console.log("链 ID:", hre.network.config.chainId);
  console.log("");

  const signers = await hre.ethers.getSigners();
  
  if (signers.length === 0) {
    console.log("❌ 未找到任何账户！");
    console.log("\n请检查 .env 文件中的 PRIVATE_KEY 配置");
    process.exit(1);
  }

  console.log(`找到 ${signers.length} 个账户:\n`);

  const roles = ["Owner (部署者)", "User1 (测试账户1)", "User2 (测试账户2)"];
  let totalBalance = 0n;

  for (let i = 0; i < signers.length; i++) {
    const signer = signers[i];
    const address = signer.address;
    const balance = await hre.ethers.provider.getBalance(address);
    const ethBalance = hre.ethers.formatEther(balance);
    
    totalBalance += balance;

    console.log(`【${roles[i] || `账户 ${i}`}】`);
    console.log(`  地址: ${address}`);
    console.log(`  余额: ${ethBalance} ETH`);
    
    if (balance === 0n) {
      console.log(`  ⚠️  警告: 余额为 0，无法支付 gas 费用`);
    } else if (balance < hre.ethers.parseEther("0.01")) {
      console.log(`  ⚠️  警告: 余额较低，建议充值`);
    } else {
      console.log(`  ✅ 余额充足`);
    }
    console.log("");
  }

  console.log("=".repeat(60));
  console.log("配置总结");
  console.log("=".repeat(60));
  console.log(`总账户数: ${signers.length}`);
  console.log(`总余额:   ${hre.ethers.formatEther(totalBalance)} ETH\n`);

  // 给出配置建议
  if (signers.length === 1) {
    console.log("📝 当前配置: 单账户模式");
    console.log("   - 可以运行部署脚本");
    console.log("   - 建议添加测试账户以运行完整的交互测试\n");
    console.log("💡 如需多账户测试:");
    console.log("   1. 在 .env 中添加:");
    console.log("      PRIVATE_KEY_USER1=0x...");
    console.log("      PRIVATE_KEY_USER2=0x...");
    console.log("   2. 给测试账户充值测试 ETH");
    console.log("   3. 重新运行此脚本验证");
  } else if (signers.length === 2) {
    console.log("📝 当前配置: 双账户模式");
    console.log("   - 可以测试基本的转账功能");
    console.log("   - 建议添加第三个账户以运行完整测试\n");
    console.log("💡 添加第三个账户:");
    console.log("   在 .env 中添加: PRIVATE_KEY_USER2=0x...");
  } else if (signers.length >= 3) {
    console.log("✅ 完美配置: 多账户模式");
    console.log("   - 可以运行完整的交互测试");
    console.log("   - 支持 mint、transfer、burn 等所有操作\n");
    console.log("🚀 下一步:");
    console.log("   npx hardhat run scripts/interact.js --network " + hre.network.name);
  }

  console.log("\n" + "=".repeat(60));
  console.log("水龙头链接（获取测试 ETH）");
  console.log("=".repeat(60));
  
  if (hre.network.name === "sepolia") {
    console.log("Sepolia 水龙头:");
    console.log("  - https://sepoliafaucet.com");
    console.log("  - https://www.infura.io/faucet/sepolia");
    console.log("  - https://faucets.chain.link/sepolia");
  } else if (hre.network.name === "base_sepolia") {
    console.log("Base Sepolia 水龙头:");
    console.log("  - https://www.coinbase.com/faucets/base-ethereum-sepolia-faucet");
    console.log("  - https://bridge.base.org/");
  }
  
  console.log("\n✅ 账户配置验证完成！");
}

main()
  .then(() => process.exit(0))
  .catch((error) => {
    console.error(error);
    process.exit(1);
  });

