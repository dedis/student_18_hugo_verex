pragma solidity ^0.4.24;

import "/Users/hugo/student_18_hugo_verex/byzcoin/contracts/Voting/ERC20.sol";
import "/Users/hugo/student_18_hugo_verex/byzcoin/contracts/Voting/SafeMath.sol";

contract VotingToken {
    // Fields
    ERC20 rewardToken;
    bool opened;
    bool closed;
    address[] votingAddresses;
    uint256 numberOrAlternatives;
    address owner;
    string name;
    string symbol;
    uint8 decimals;
    uint256 totalSupply;
    mapping(address => uint256) balances;
    mapping(address => mapping(address => uint256)) allowed;
    address[] participants;

    // Enumerations

    // Constructor
    constructor () public {
    }

    // Public functions
    function mint (address _to, uint256 _amount) public returns (bool) {
        require(onlyOwner(), "error");
        require(!(opened), "error");

        uint256 newBalance = SafeMath.add(balances[_to], _amount);
        totalSupply = SafeMath.add(totalSupply, _amount);
        balances[_to] = newBalance;
        return true;
    }

    function transferOwnership (address newOwner) public {
        require(onlyOwner() && !(newOwner == address(0)), "error");
        owner = newOwner;
    }

    function MAX_NUMBER_OF_ALTERNATIVES () public pure returns (uint256) {
        return 255;
    }

    function open () public {
        require(onlyOwner(), "error");
        require(!(opened), "error");
        opened = true;
    }

    function transfer (address _to, uint256 _value) public returns (bool) {
        require(!(_to == address(0)), "error");
        require(_value <= balances[msg.sender], "error");

        balances[msg.sender] = SafeMath.sub(balances[msg.sender], _value);
        balances[_to] = SafeMath.add(balances[_to], _value);
        _rewardVote(msg.sender, _to, _value);
        return true;
    }

    function approve (address _spender, uint256 _value) public returns (bool) {
        allowed[msg.sender][_spender] = _value;
        return true;
    }

    function REWARD_RATIO () public pure returns (uint256) {
        return 100;
    }

    function allowance (address _owner, address _spender) public view returns (uint256) {
        return allowed[_owner][_spender];
    }

    function close () public {
        require(onlyOwner(), "error");
        require(opened, "error");
        require(!(closed), "error");
        closed = true;
    }

    function balanceOf (address _owner) public view returns (uint256) {
        return balances[_owner];
    }

    function destroy (ERC20[] tokens) public {
        require(onlyOwner(), "error");
        transferToken(tokens, 0);
        selfdestruct(owner);
    }

    function transferFrom (address _from, address _to, uint256 _value) public returns (bool) {
        require(!(_to == address(0)), "error");
        require(_value <= balances[_from], "error");
        require(_value <= allowed[_from][msg.sender], "error");

        balances[_from] = SafeMath.sub(balances[_from], _value);
        balances[_to] = SafeMath.add(balances[_to], _value);
        allowed[_from][msg.sender] = SafeMath.sub(allowed[_from][msg.sender], _value);
        return true;
    }

    function onlyOwner () public view returns (bool) {
        return msg.sender == owner;
    }

    // Private functions
    function _isVotingAddress (address votingAddress) view private returns (bool) {
        return _isVotingAddressFrom(0, votingAddress);
    }

    function _isVotingAddressFrom (uint256 i, address votingAddress) view private returns (bool) {
        if (i >= votingAddresses.length) {
            return false;
        } else {
            if (votingAddresses[i] == votingAddress) {
                return true;
            } else {
                return _isVotingAddressFrom(i + 1, votingAddress);
            }
        }
    }

    function transferToken (ERC20[] tokens, uint256 i) private {
        if (i < tokens.length) {
            tokens[i].transfer(owner, tokens[i].balanceOf(address(this)));
            transferToken(tokens, i + 1);
        }

    }

    function _rewardVote (address _from, address _to, uint256 _value) private returns (bool) {
        if (_isVotingAddress(_to)) {
            assert(opened && !(closed));
            uint256 rewardTokens = SafeMath.div(_value, REWARD_RATIO());
            return rewardToken.transfer(_from, rewardTokens);
        } else {
            return false;
        }
    }


}

