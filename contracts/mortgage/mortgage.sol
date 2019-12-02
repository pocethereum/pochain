pragma solidity 0.4.8;

contract MortageSystem {
    event Mortgage(address indexed to, uint256 value);
    event Redeem(address indexed owner, uint256 value);

    mapping (address => uint256) private _mortgages;
    uint256 private _totalMortgage;
    uint256 private _totalRewarded;

    function MortageSystem() public {
        _totalMortgage = 0xABCD;//INIT VALUE NOT USED
        _totalRewarded = 0xEFAB;//INIT VALUE NOT USED
    }

    function totalMortgage() public constant returns (uint256) {
        return _totalMortgage;
    }

    function mortgageOf(address owner) public constant returns (uint256) {
        return _mortgages[owner];
    }


    function mortgage(address to, uint256 value)payable public returns (bool) {
        if(to == address(0)){
            return false;
        }

        if (msg.sender.send(value)){
             return false;
        }

        _mortgages[to] += value;
        _totalMortgage+=value;

        Mortgage(to, value);
        return true;
    }

    function redeem(address owner, uint256 value) public returns (bool) {
        if (owner.send(value)){
             return false;
        }
        _mortgages[msg.sender] -= value;
        _totalMortgage-=value;

        Redeem(owner, value);
        return true;
    }
}
