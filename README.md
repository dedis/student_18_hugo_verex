# Verex

## Overview

Verex is an experimental library to deploy solidity smart contracts on the [Byzcoin](https://github.com/dedis/cothority/tree/master/byzcoin) ledger.



### Prerequisites
This tool requires a solidity and a go compiler.

**Go compiler** :

OSX :

`brew install go`

Ubuntu : 

`sudo apt-get install golang-go`

**Solidity compiler** :

using npm

`npm install -g solc`

## Installation

```
$ git clone https://github.com/dedis/student_18_hugo_verex
//...

$ cd student_18_hugo_verex/byzcoin
$ go build
```

## Test 

The current tests will deploy and call the `ModifiedToken.sol` contract.

`go test -run Spawn` will test spawning a byzcoin ledger

`go test -run Display` will display the test account balance

`go test -run Credit` will credit the test account

`go test -run Deploy` will test deploying the contract

`go test -run Transaction` will test the minting of new tokens from the above contract





## Formal verification of smart contracts

It is possible to verify formally smart contract using the [Stainless](https://github.com/epfl-lara/stainless)  library.
For the moment only the MinimumToken is usable. Further integration with the library is planned in the future.

  

 





