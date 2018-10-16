pragma solidity ^0.4.24;


contract MinimumToken {
    // Fields
    mapping(address => uint256) balanceOf;
    uint256 total;
    address[] participants;

    // Enumerations

    // Constructor
    constructor () public {
    }

    // Public functions
    function transferFrom (address from, address to, uint256 amount) public {
        require(!(to == address(0)), "error");
        require(!(from == to), "error");
        require(amount <= balanceOf[from], "error");

        balanceOf[from] = balanceOf[from] - amount;
        balanceOf[to] = balanceOf[to] + amount;

    }

    // Private functions

}
