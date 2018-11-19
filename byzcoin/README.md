# Solidity smart contracts on Byzcoin

## Byzcoin Virtual Machine instructions

- `Spawn` Instantiate a new ledger for testing the BVM
- `Invoke:deploy` deploys a contract using as argument the compiled bytecode
- `Ìnvoke:call` calls a deployed contract using a contract address, an [ABI](https://solidity.readthedocs.io/en/develop/abi-spec.html), the name of the function
to call and the arguments of said function



## Files

The following files are in this directory:


- `bvmContract.go` defines the byzcoin contract that interacts with the Ethereum Virtual Machine
- `database.go` redefines the ethereum database functions to be compatible with Byzcoin
- `params.go` defines the parameter of a BVM
- `keys.go` helper methods for Ethereum key management 
- `service.go` only serves to register the contract with ByzCoin. If you
want to give more power to your service, be sure to look at the
[../service](service example).
- `proto.go` has the definitions that will be translated into protobuf

