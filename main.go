package main

import (
	"fmt"
	"log"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/core/vm/runtime"
)

func main() {

	simple_bin := "608060405234801561001057600080fd5b50610108806100206000396000f3006080604052600436106049576000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff1680633450bd6a14604e578063943640c3146076575b600080fd5b348015605957600080fd5b50606060ca565b6040518082815260200191505060405180910390f35b348015608157600080fd5b50608860d4565b604051808273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200191505060405180910390f35b6000611000905090565b6000339050905600a165627a7a72305820ab0d895bf02f28d1a6446166e2e71913cd6b963c0eea3617ff1bf05b94e747cb0029"

	simple_abi := `[{"constant":true,"inputs":[],"name":"returnNumber","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"pure","type":"function"},{"constant":true,"inputs":[],"name":"returnAddress","outputs":[{"name":"","type":"address"}],"payable":false,"stateMutability":"view","type":"function"}]`

	public_key, _ := getKeys()

	//contract_pk := common.HexToAddress("0xBd770416a3345F91E4B34576cb804a576fa48EB1")

	abi, err := abi.JSON(strings.NewReader(simple_abi))
	if err != nil {
		fmt.Println(err)
	}
	returnAddress, err := abi.Pack("returnAddress")
	if err != nil {
		fmt.Println(err)
	}
	returnNumber, err := abi.Pack("returnNumber")
	if err != nil {
		fmt.Println(err)
	}

	sdb, _ := getDB()

	test := func(vm.StateDB, common.Address, *big.Int) bool {
		return true
	}

	transfer := func(vm.StateDB, common.Address, common.Address, *big.Int) {
		log.Println("tried to transfer")
	}

	gethash := func(uint64) common.Hash {
		return common.HexToHash("0x0000000000000000000000000000000000000000")
	}

	fmt.Println("Setting up balances")

	sdb.SetBalance(common.HexToAddress(public_key), big.NewInt(1000000000000))

	fmt.Println("Sending", sdb.GetBalance(common.HexToAddress(public_key)), "eth to", common.HexToAddress(public_key).Hex())

	fmt.Println("Setting up context")
	ctx := vm.Context{CanTransfer: test, Transfer: transfer, GetHash: gethash, Origin: common.HexToAddress(public_key), GasPrice: big.NewInt(1), Coinbase: common.HexToAddress(public_key), GasLimit: 10000000000, BlockNumber: big.NewInt(0), Time: big.NewInt(1), Difficulty: big.NewInt(1)}

	fmt.Println("Setting up VMs")
	bvm := vm.NewEVM(ctx, sdb, getChainConfig(), getVMConfig())
	bvm1 := runtime.NewEnv(getConfig())

	fmt.Println("...........................")

	fmt.Println("===== Through vm =====")

	accountRef := &vm.AccountRef{}
	ret, addrContract, leftOverGas, err := bvm.Create(accountRef, common.Hex2Bytes(simple_bin), 100000000, big.NewInt(0))
	if err != nil {
		fmt.Println("Contract deployment unsuccessful")
		fmt.Println(err)
	} else {
		fmt.Println("Successful contract deployment")
	}
	fmt.Println("Return of contract", common.Bytes2Hex(ret))
	fmt.Println("Left over gas : ", leftOverGas)
	fmt.Println("Contract address", addrContract.Hex())

	fmt.Println("Contract call")
	ret_call, leftOverGas, err := bvm.Call(accountRef, addrContract, returnAddress, leftOverGas, big.NewInt(0))
	if err != nil {
		fmt.Println("Contract call unsuccessful")
		fmt.Println(err)
	} else {
		fmt.Println("Successful contract call")
	}
	fmt.Println("Return of call", common.Bytes2Hex(ret_call))
	fmt.Println("Left over gas : ", leftOverGas)
	fmt.Println("Nonce contract", sdb.GetNonce(addrContract))

	ret_call1, leftOverGas1, err := bvm.Call(accountRef, addrContract, returnNumber, leftOverGas, big.NewInt(0))
	if err != nil {
		fmt.Println("Contract call unsuccessful")
		fmt.Println(err)
	} else {
		fmt.Println("Successful contract call")
	}
	fmt.Println("Return of call", common.Bytes2Hex(ret_call1))
	fmt.Println("Left over gas : ", leftOverGas1)
	fmt.Println("Nonce contract", sdb.GetNonce(addrContract))

	fmt.Println("===== End vm =====")

	bvmInterpreter := vm.NewEVMInterpreter(bvm, getVMConfig())
	fmt.Println(bvmInterpreter.CanRun(common.Hex2Bytes(simple_bin)))

	bvm1Interpreter := vm.NewEVMInterpreter(bvm1, getVMConfig())
	fmt.Println(bvm1Interpreter.CanRun(common.Hex2Bytes(simple_bin)))

	//func (in *EVMInterpreter) Run(contract *Contract, input []byte, readOnly bool) (ret []byte, err error)

	/*contract := &vm.Contract{
		CallerAddress: common.HexToAddress("0xE420b7546D387039dDaD2741a688CbEBD2578363"),
		Code:          common.Hex2Bytes(minimum_token),
		CodeHash:      common.HexToHash("0x0000000000000000000000000000000000000000"),
		//CodeAddr:      &(common.HexToAddress("0xE420b7546D387039dDaD2741a688CbEBD2578363")),
		Input: nil,
		Gas:   1000000,
	}*/

	//fmt.Println(ret)
	//fmt.Println(addrContract)

	//new_contract := vm.NewContract(accountRef, contract, big.NewInt(10000), 1)

}
