pragma solidity ^0.4.23;




contract simple {
  string[] public arr;
  int public counter = 0;

  function paint(string new_word) returns (string) {

      arr.push(picture(msg.sender, counter, _ipfsAddress));
      return true;


  }
}
