pragma solidity ^0.4.19;


contract simple {
  string name;


  function returnAddress(string _param) public view returns (string) {
    name = _param;
    return (name);
  }

  function returnNumber() public pure returns (uint){
    return 4098;
  }

}
