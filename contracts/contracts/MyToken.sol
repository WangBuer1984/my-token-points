// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import "@openzeppelin/contracts/access/Ownable.sol";

/**
 * @title MyToken
 * @dev ERC20 代币合约，支持 mint 和 burn 功能
 * @notice 用于积分系统的代币合约，记录所有余额变化事件
 */
contract MyToken is ERC20, Ownable {
    
    /// @notice 代币铸造事件
    /// @param to 接收地址
    /// @param amount 铸造数量
    /// @param timestamp 区块时间戳
    event TokenMinted(address indexed to, uint256 amount, uint256 timestamp);
    
    /// @notice 代币销毁事件
    /// @param from 销毁地址
    /// @param amount 销毁数量
    /// @param timestamp 区块时间戳
    event TokenBurned(address indexed from, uint256 amount, uint256 timestamp);
    
    /**
     * @dev 构造函数
     * @notice 初始化代币名称和符号
     */
    constructor() ERC20("MyToken", "MTK") Ownable(msg.sender) {
        // 初始化完成，不预挖任何代币
    }
    
    /**
     * @notice 铸造代币 (仅限 owner)
     * @param to 接收地址
     * @param amount 铸造数量
     */
    function mint(address to, uint256 amount) public onlyOwner {
        require(to != address(0), "MyToken: mint to zero address");
        require(amount > 0, "MyToken: mint amount must be positive");
        
        _mint(to, amount);
        emit TokenMinted(to, amount, block.timestamp);
    }
    
    /**
     * @notice 销毁代币
     * @param amount 销毁数量
     */
    function burn(uint256 amount) public {
        require(amount > 0, "MyToken: burn amount must be positive");
        require(balanceOf(msg.sender) >= amount, "MyToken: insufficient balance");
        
        _burn(msg.sender, amount);
        emit TokenBurned(msg.sender, amount, block.timestamp);
    }
    
    /**
     * @notice 授权销毁代币
     * @param from 销毁地址
     * @param amount 销毁数量
     */
    function burnFrom(address from, uint256 amount) public {
        require(amount > 0, "MyToken: burn amount must be positive");
        require(from != address(0), "MyToken: burn from zero address");
        
        _spendAllowance(from, msg.sender, amount);
        _burn(from, amount);
        emit TokenBurned(from, amount, block.timestamp);
    }
}

