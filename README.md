# Verex

## Overview




Verex is a tool to deploy solidity smart-contracts generated with the formal verification [Stainless](https://github.com/epfl-lara/stainless) library  to the [Byzcoin](https://github.com/dedis/cothority/tree/master/byzcoin) ledger.

## Installation

```
$ git clone https://github.com/dedis/student_18_hugo_verex
//...
$ cd student_18_hugo_verex/byzcoin
```


## From Stainless..

Install [Smart](https://github.com/epfl-lara/smart) which is the current solidity interpreter of the [LARA](https://github.com/epfl-lara) laboratory.
```
$ git clone https://github.com/epfl-lara/smart.git
Cloning into 'smart'...
// ...
$ cd smart
$ sbt clean universal:stage  
```
You can then create a symbolic link (e.g. for Linux & Mac OS-X) to have access 
to a ``stainless`` command-line. 

```bash
 ln -s frontends/scalac/target/universal/stage/bin/stainless-scalac stainless
```
Once the setup is done, you can now create a new smart contract using the  [`Candy`](frontends/benchmarks/smartcontracts/valid/Candy.scala) contract as a template. 

If you wish to use one of the [existing](frontends/benchmarks/smartcontracts/valid)  contracts, you can define a constructor with the desired parameters directly with Scala.  
 
Stainless is able to verify that the assertions written in the contract are indeed valid : 

```./stainless *.scala ``` or adding the ``` --strict-arithmetic``` flag to verify overflows issues. 

For more details on formal verification refer of course to the [Stainless](https://github.com/epfl-lara/stainless) and [Smart](https://github.com/epfl-lara/smart) repositories.


## ..to Solidity

Copy your Scala code into a new folder with the contract name into the [contracts](byzcoin/contracts) folder.

Run 

```./stainless *.scala --solidity ```

which will produce a new solidity file with your contract name. 

Install a [Solidity compiler](https://solidity.readthedocs.io/en/v0.4.24/installing-solidity.html).

Run 

``` 
solcjs *.sol --bin
solcjs *.sol --abi
```

Which will generate a `ContractName_sol_ContractName.bin` and `ContractName_sol_ContractName.abi` file containing the bytecode and the [ABI](https://solidity.readthedocs.io/en/develop/abi-spec.html) of the smart contract.
 
## .. to Byzcoin

you can now test your contract deployment on the Byzcoin ledger by refering to the bvmContract_test.go and to the detailed [README.md](byzcoin/README.md) of the Byzcoin Virtual Machine.

#### Acknowledgments
Project was made in conjunction with DEDIS and LARA laboratories at EPFL. 




 






  

 





