pragma solidity ^0.4.19;


contract simple {

  function returnAddress() public view returns (address) {
    return (msg.sender);
  }

  function returnNumber() public pure returns (uint){
    return 4096;
  }

}
