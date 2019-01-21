pragma solidity ^0.4.24;


library SafeMath{
    function sub (uint256 a, uint256 b) public pure returns (uint256) {
        require(b <= a, "error");
        return a - b;
    }

    function div (uint256 a, uint256 b) public pure returns (uint256) {
        require(b > 0, "error");
        return a / b;
    }

    function add (uint256 a, uint256 b) public pure returns (uint256) {
        uint256 c = a + b;
        require(c >= a, "error");
        return c;
    }

    function mul (uint256 a, uint256 b) public pure returns (uint256) {
        uint256 c = a * b;
        require(a == 0 || c / a == b, "error");
        return c;
    }


}