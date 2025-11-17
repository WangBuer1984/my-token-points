const hre = require("hardhat");

async function main() {
  console.log("=".repeat(60));
  console.log("给测试账户转账 ETH");
  console.log("=".repeat(60));
  console.log("\n网络:", hre.network.name);
  console.log("");

  const signers = await hre.ethers.getSigners();
  
  if (signers.length < 2) {
    console.log("❌ 只找到 1 个账户，无需转账");
    console.log("请在 .env 中配置 PRIVATE_KEY_USER1 和 PRIVATE_KEY_USER2");
    process.exit(0);
  }

  const [owner, user1, user2] = signers;
  
  // 获取 Owner 余额
  const ownerBalance = await hre.ethers.provider.getBalance(owner.address);
  console.log("【Owner 账户】");
  console.log("地址:", owner.address);
  console.log("余额:", hre.ethers.formatEther(ownerBalance), "ETH");
  
  if (ownerBalance < hre.ethers.parseEther("0.5")) {
    console.log("\n❌ Owner 余额不足，无法进行转账");
    console.log("请先给 Owner 账户充值测试 ETH");
    process.exit(1);
  }
  
  console.log("\n" + "-".repeat(60));

  // 转账金额
  const fundAmount = hre.ethers.parseEther("0.1");
  const minBalance = hre.ethers.parseEther("0.05");

  // 给 User1 转账
  if (user1) {
    const balance1 = await hre.ethers.provider.getBalance(user1.address);
    console.log("\n【User1 账户】");
    console.log("地址:", user1.address);
    console.log("余额:", hre.ethers.formatEther(balance1), "ETH");
    
    if (balance1 < minBalance) {
      console.log(`余额不足，转账 ${hre.ethers.formatEther(fundAmount)} ETH...`);
      try {
        const tx1 = await owner.sendTransaction({
          to: user1.address,
          value: fundAmount
        });
        console.log("交易哈希:", tx1.hash);
        console.log("等待确认...");
        await tx1.wait();
        
        const newBalance1 = await hre.ethers.provider.getBalance(user1.address);
        console.log("✅ 转账成功! 新余额:", hre.ethers.formatEther(newBalance1), "ETH");
      } catch (error) {
        console.log("❌ 转账失败:", error.message);
      }
    } else {
      console.log("✅ 余额充足，跳过转账");
    }
  }

  // 给 User2 转账
  if (user2) {
    const balance2 = await hre.ethers.provider.getBalance(user2.address);
    console.log("\n【User2 账户】");
    console.log("地址:", user2.address);
    console.log("余额:", hre.ethers.formatEther(balance2), "ETH");
    
    if (balance2 < minBalance) {
      console.log(`余额不足，转账 ${hre.ethers.formatEther(fundAmount)} ETH...`);
      try {
        const tx2 = await owner.sendTransaction({
          to: user2.address,
          value: fundAmount
        });
        console.log("交易哈希:", tx2.hash);
        console.log("等待确认...");
        await tx2.wait();
        
        const newBalance2 = await hre.ethers.provider.getBalance(user2.address);
        console.log("✅ 转账成功! 新余额:", hre.ethers.formatEther(newBalance2), "ETH");
      } catch (error) {
        console.log("❌ 转账失败:", error.message);
      }
    } else {
      console.log("✅ 余额充足，跳过转账");
    }
  }

  console.log("\n" + "=".repeat(60));
  console.log("最终余额");
  console.log("=".repeat(60));

  // 显示最终余额
  for (let i = 0; i < signers.length; i++) {
    const signer = signers[i];
    const balance = await hre.ethers.provider.getBalance(signer.address);
    const role = i === 0 ? "Owner" : `User${i}`;
    console.log(`${role}: ${hre.ethers.formatEther(balance)} ETH`);
  }

  console.log("\n✅ 资金准备完成!");
  console.log("\n🚀 现在可以运行交互脚本:");
  console.log(`   npx hardhat run scripts/interact.js --network ${hre.network.name}`);
}

main()
  .then(() => process.exit(0))
  .catch((error) => {
    console.error(error);
    process.exit(1);
  });

