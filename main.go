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

	contracts_path := "/Users/hugo/student_18_hugo_verex/contracts/"

	simple_abi, simple_bin := getSC(contracts_path, "ModifiedToken")

	A_public_key, _ := getKeys()
	B_public_key, _ := getKeys1()

	accountRef := vm.AccountRef(common.HexToAddress(A_public_key))

	abi, err := abi.JSON(strings.NewReader(simple_abi))
	if err != nil {
		fmt.Println(err)
	}

	create, err := abi.Pack("create", big.NewInt(4096), common.HexToAddress(A_public_key))
	if err != nil {
		fmt.Println(err)
	}

	get, err := abi.Pack("getBalance", common.HexToAddress(A_public_key))
	if err != nil {
		fmt.Println(err)
	}

	send, err := abi.Pack("send", common.HexToAddress(A_public_key), common.HexToAddress(B_public_key), big.NewInt(16))
	if err != nil {
		fmt.Println(err)
	}

	get1, err := abi.Pack("getBalance", common.HexToAddress(B_public_key))
	if err != nil {
		fmt.Println(err)
	}
	get2, err := abi.Pack("getBalance", common.HexToAddress(A_public_key))
	if err != nil {
		fmt.Println(err)
	}

	transfer_test, err := abi.Pack("transfer", common.HexToAddress(B_public_key), 10)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("DB setup")

	sdb, _ := getDB()

	canTransfer := func(vm.StateDB, common.Address, *big.Int) bool {
		//log.Println("Verified transfer")
		return true
	}

	transfer := func(vm.StateDB, common.Address, common.Address, *big.Int) {
		//log.Println("tried to transfer")
	}

	gethash := func(uint64) common.Hash {
		log.Println("tried to get hash")
		return common.HexToHash("0x0000000000000000000000000000000000000000")
	}

	sdb.SetBalance(common.HexToAddress(A_public_key), big.NewInt(1000000000000))

	fmt.Println("Setting up context")
	ctx := vm.Context{CanTransfer: canTransfer, Transfer: transfer, GetHash: gethash, Origin: common.HexToAddress(A_public_key), GasPrice: big.NewInt(1), Coinbase: common.HexToAddress(A_public_key), GasLimit: 10000000000, BlockNumber: big.NewInt(0), Time: big.NewInt(1), Difficulty: big.NewInt(1)}

	fmt.Println("Setting up & checking VMs")
	bvm := vm.NewEVM(ctx, sdb, getChainConfig(), getVMConfig())
	bvm1 := runtime.NewEnv(getConfig())
	bvmInterpreter := vm.NewEVMInterpreter(bvm, getVMConfig())
	bvm1Interpreter := vm.NewEVMInterpreter(bvm1, getVMConfig())
	a := bvmInterpreter.CanRun(common.Hex2Bytes(simple_bin))
	b := bvm1Interpreter.CanRun(common.Hex2Bytes(simple_bin))
	if !a || !b {
		fmt.Println("Problem setting up vms")
	}

	fmt.Println("======== Contract creation ========")
	ret, addrContract, leftOverGas, err := bvm.Create(accountRef, common.Hex2Bytes(simple_bin), 100000000, big.NewInt(0))
	if err != nil {
		fmt.Println("Contract deployment unsuccessful")
		fmt.Println("Return of contract creation", common.Bytes2Hex(ret))
		fmt.Println(err)
	} else {
		fmt.Println("- Successful contract deployment")
		fmt.Println("- New contract address", addrContract.Hex())
	}

	fmt.Println("======== Contract call ========")
	_, leftOverGas, err = bvm.Call(accountRef, addrContract, create, leftOverGas, big.NewInt(0))
	if err != nil {
		fmt.Println("token creation unsuccessful")
		fmt.Println(err)
	} else {
		fmt.Println("- Successful token creation")
	}
	get_call, leftOverGas, err := bvm.Call(accountRef, addrContract, get, leftOverGas, big.NewInt(0))
	if err != nil {
		fmt.Println("get unsuccessful")
		fmt.Println(err)
	} else {
		fmt.Println("- Successful balance fetch")
		fmt.Println("The balance of : ", A_public_key, " is ", get_call)
	}

	_, leftOverGas, err = bvm.Call(accountRef, addrContract, send, leftOverGas, big.NewInt(0))
	if err != nil {
		fmt.Println("send unsuccessful")
		fmt.Println(err)
	} else {
		fmt.Println("- Successful send from ", A_public_key, " to ", B_public_key)
	}

	get1_call, leftOverGas, err := bvm.Call(accountRef, addrContract, get1, leftOverGas, big.NewInt(0))
	if err != nil {
		fmt.Println("get unsuccessful")
		fmt.Println(err)
		fmt.Println("Left over gas : ", leftOverGas)
	} else {
		fmt.Println("- Successful balance fetch")
		fmt.Println("The balance of :", B_public_key, " is ", get1_call)
	}
	get11_call, leftOverGas, err := bvm.Call(accountRef, addrContract, get2, leftOverGas, big.NewInt(0))
	if err != nil {
		fmt.Println("get unsuccessful")
		fmt.Println("Left over gas : ", leftOverGas)
		fmt.Println(err)
	} else {
		fmt.Println("- Successful balance fetch")
		fmt.Println("Balance of ", A_public_key, " is ", get11_call)
	}

	transfer_call, leftOverGas, err := bvm.Call(accountRef, addrContract, transfer_test, leftOverGas, big.NewInt(0))
	if err != nil {
		fmt.Println("transfer unsuccessful")
		fmt.Println("Left over gas : ", leftOverGas)
		fmt.Println(err)
	} else {
		fmt.Println("Successful transfer")
		fmt.Println("Return of call", common.Bytes2Hex(transfer_call))
	}

}
