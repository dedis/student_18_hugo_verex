pragma solidity ^0.4.20;

contract ModifiedToken {
    /* This creates an array with all balances */
    mapping (address => uint256) public balanceOf;

    /* Initializes contract with initial supply tokens to the creator of the contract */
    function ModifiedToken(
        uint256 initialSupply,
        address toGiveTo
        ) public {
        balanceOf[toGiveTo] = initialSupply;              // Give the creator all initial tokens
    }

    function send(address _to, uint amount) public returns (bool){
      balanceOf[_to] = amount;
      return true;
    }

    /* Send coins */
    function transfer(address _to, uint256 _value) public returns (bool success) {
        require(balanceOf[msg.sender] >= _value);           // Check if the sender has enough
        require(balanceOf[_to] + _value >= balanceOf[_to]); // Check for overflows
        balanceOf[msg.sender] -= _value;                    // Subtract from the sender
        balanceOf[_to] += _value;                           // Add the same to the recipient
        return true;
    }

    function getBalance(address account) public view returns (uint256){
      return balanceOf[account];
    }
}
