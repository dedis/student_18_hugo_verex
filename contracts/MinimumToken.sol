pragma solidity ^0.4.24;

//No overflow check! Check 0xc20cC13f6efe457b515eAab43EAE2cdD6B28583a on rinkeby network

contract MinimumToken {
    // Fields
    mapping(address => uint32) balanceOf;
    uint32 total;
    address[] participants;

    // Enumerations

    // Constructor
    constructor () public {
      balanceOf[msg.sender] = 4294967295;
      balanceOf[0xe745E7ceA88A02a1Fabd4aE591371eF50BFDc099] = 1;
    }

    // Public functions
    function transferFrom (address from, address to, uint32 amount) public {
        require(!(to == address(0)), "error");
        require(!(from == to), "error");
        require(amount <= balanceOf[from], "error");

        balanceOf[from] = balanceOf[from] - amount;
        balanceOf[to] = balanceOf[to] + amount;

    }

    function getBalance(address _account) public view returns (uint32){
      return balanceOf[_account];
    }

    // Private functions

}
