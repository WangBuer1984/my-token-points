require("@nomicfoundation/hardhat-toolbox");
require("dotenv").config();

// 配置多账户支持
function getAccounts() {
  const accounts = [];
  
  // 主账户（必需）
  if (process.env.PRIVATE_KEY) {
    accounts.push(process.env.PRIVATE_KEY);
  }
  
  // 测试账户1（可选）
  if (process.env.PRIVATE_KEY_USER1) {
    accounts.push(process.env.PRIVATE_KEY_USER1);
  }
  
  // 测试账户2（可选）
  if (process.env.PRIVATE_KEY_USER2) {
    accounts.push(process.env.PRIVATE_KEY_USER2);
  }
  
  return accounts;
}

/** @type import('hardhat/config').HardhatUserConfig */
module.exports = {
  solidity: {
    version: "0.8.20",
    settings: {
      optimizer: {
        enabled: true,
        runs: 200
      }
    }
  },
  
  networks: {
    // Sepolia 测试网
    sepolia: {
      url: process.env.SEPOLIA_RPC_URL || 
           "https://eth-sepolia.g.alchemy.com/v2/demo",
      accounts: getAccounts(), // 支持多账户
      chainId: 11155111
    },
    
    // Base Sepolia 测试网
    base_sepolia: {
      url: process.env.BASE_SEPOLIA_RPC_URL || 
           "https://sepolia.base.org",
      accounts: getAccounts(), // 支持多账户
      chainId: 84532,
      gasPrice: 1000000000 // 1 gwei
    }
  },
  
  // ⭐ Etherscan API V2 配置（统一验证所有链）
  // Basescan 已合并到 Etherscan，只需要一个 API Key
  etherscan: {
    apiKey: {
      // 所有网络使用同一个 Etherscan API Key
      sepolia: process.env.ETHERSCAN_API_KEY || "",
      baseSepolia: process.env.ETHERSCAN_API_KEY || "",  // 使用同一个 Key
    },
    customChains: [
      {
        network: "baseSepolia",
        chainId: 84532,
        urls: {
          // Base Sepolia 由 Etherscan 管理
          apiURL: "https://api-sepolia.basescan.org/api",
          browserURL: "https://sepolia.basescan.org"
        }
      }
    ]
  },
  
  // 可选：使用去中心化的 Sourcify 验证
  sourcify: {
    enabled: true
  },
  
  paths: {
    sources: "./contracts",
    tests: "./test",
    cache: "./cache",
    artifacts: "./artifacts"
  }
};

