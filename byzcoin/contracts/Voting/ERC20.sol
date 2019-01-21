pragma solidity ^0.4.24;


interface ERC20 {

    function transferFrom(address from, address to, uint256 tokens) external returns (bool);
    function totalSupply() external view returns (uint256);
    function approve(address spender, uint256 tokens) external returns (bool);
    function transfer(address to, uint256 tokens) external returns (bool);
    function balanceOf(address tokenOwner) external view returns (uint256);
    function allowance(address tokenOwner, address spender) external view returns (uint256);
}

