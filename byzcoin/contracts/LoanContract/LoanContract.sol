pragma solidity ^0.4.24;

import "/Users/hugo/student_18_hugo_verex/byzcoin/contracts/LoanContract/ERC20Token.sol";

contract LoanContract {
    // Fields
    address borrower;
    uint256 wantedAmount;
    uint256 premiumAmount;
    uint256 tokenAmount;
    string tokenName;
    ERC20Token tokenContractAddress;
    uint256 daysToLend;
    State currentState;
    uint256 start;
    address lender;
    State[] visitedStates;

    // Enumerations
    enum State{
        WaitingForPayback,
        Finished,
        Default,
        WaitingForData,
        WaitingForLender
    }


    // Constructor
    constructor () public {
    }

    // Public functions
    function lend () public payable {
        if (!(msg.sender == borrower)) {
            if (currentState == State.WaitingForLender && msg.value >= wantedAmount) {
                lender = msg.sender;
                borrower.transfer(wantedAmount);

                currentState = State.WaitingForPayback;
                start = now;
            }

        }

    }

    function checkTokens () public {
        if (currentState == State.WaitingForData) {
            uint256 balance = tokenContractAddress.balanceOf(address(this));
            if (balance >= tokenAmount) {

                currentState = State.WaitingForLender;
            }

        }

    }

    function requestDefault () public {
        if (currentState == State.WaitingForPayback) {
            require(now > start + daysToLend, "error");
            require(msg.sender == lender, "error");
            uint256 balance = tokenContractAddress.balanceOf(address(this));
            tokenContractAddress.transfer(lender, balance);

            currentState = State.Default;
        }

    }

    function payback () public payable {
        require(address(this).balance >= msg.value, "error");
        require(msg.value >= premiumAmount + wantedAmount, "error");
        require(msg.sender == lender, "error");
        if (currentState == State.WaitingForPayback) {
            lender.transfer(msg.value);
            uint256 balance = tokenContractAddress.balanceOf(address(this));
            tokenContractAddress.transfer(borrower, balance);

            currentState = State.Finished;
        }

    }

    // Private functions

}

