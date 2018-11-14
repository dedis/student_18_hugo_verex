Navigation: [DEDIS](https://github.com/dedis/doc/tree/master/README.md) ::
[../README.md](Cothority Template) ::
ByzCoin Example

# Using solidity with Byzcoin

## Byzcoin Virtual Machine

- `Spawn` Instantiate a new ethereum virtual machine
- `Invoke:creditAccount` Credits the account passed in argument with the value indicated
- `Invoke:deployContract` Deploys a contract 
- `Ìnvoke:methodCall` Call a specific method at the contract address

Both of these options are protected by the darc where the value will be stored.


## Files

The following files are in this directory:

- `service.go` only serves to register the contract with ByzCoin. If you
want to give more power to your service, be sure to look at the
[../service](service example).
- `bvmContract.go` defines the contract
- `proto.go` has the definitions that will be translated into protobuf

### A word on ProtoBuf

Usually you start protobuf with a `.proto` file and then translate it to
different languages. Because we're using the `dedis/protobuf` library,
our definitions reside in the `proto.go` files and are translated using
`proto.awk` into `.proto` files. These files are then used to create the
java and javascript definitions.
