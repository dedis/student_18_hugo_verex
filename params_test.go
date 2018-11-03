package main

import (
	"fmt"
	"log"
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/core/vm/runtime"
)

func TestTokenContract(t *testing.T) {
	contractsPath := "/Users/hugo/student_18_hugo_verex/contracts/"

	simpleAbi, simpleBin := getSmartContract(contractsPath, "ModifiedToken")

	aPublicKey, _ := getKeys()
	bPublicKey, _ := getKeys1()

	accountRef := vm.AccountRef(common.HexToAddress(aPublicKey))

	abi, err := abi.JSON(strings.NewReader(simpleAbi))
	if err != nil {
		fmt.Println(err)
	}

	create, err := abi.Pack("create", big.NewInt(4096), common.HexToAddress(aPublicKey))
	if err != nil {
		fmt.Println(err)
	}

	get, err := abi.Pack("getBalance", common.HexToAddress(aPublicKey))
	if err != nil {
		fmt.Println(err)
	}

	send, err := abi.Pack("transfer", common.HexToAddress(aPublicKey), common.HexToAddress(bPublicKey), big.NewInt(16))
	if err != nil {
		fmt.Println(err)
	}

	get1, err := abi.Pack("getBalance", common.HexToAddress(bPublicKey))
	if err != nil {
		fmt.Println(err)
	}
	get2, err := abi.Pack("getBalance", common.HexToAddress(aPublicKey))
	if err != nil {
		fmt.Println(err)
	}

	transferTests, err := abi.Pack("transfer", common.HexToAddress(aPublicKey), common.HexToAddress(bPublicKey), big.NewInt(16))
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

	sdb.SetBalance(common.HexToAddress(aPublicKey), big.NewInt(1000000000000))

	fmt.Println("Setting up context")
	ctx := vm.Context{CanTransfer: canTransfer, Transfer: transfer, GetHash: gethash, Origin: common.HexToAddress(aPublicKey), GasPrice: big.NewInt(1), Coinbase: common.HexToAddress(aPublicKey), GasLimit: 10000000000, BlockNumber: big.NewInt(0), Time: big.NewInt(1), Difficulty: big.NewInt(1)}

	fmt.Println("Setting up & checking VMs")
	bvm := vm.NewEVM(ctx, sdb, getChainConfig(), getVMConfig())
	bvm1 := runtime.NewEnv(getConfig())
	bvmInterpreter := vm.NewEVMInterpreter(bvm, getVMConfig())
	bvm1Interpreter := vm.NewEVMInterpreter(bvm1, getVMConfig())
	a := bvmInterpreter.CanRun(common.Hex2Bytes(simpleBin))
	b := bvm1Interpreter.CanRun(common.Hex2Bytes(simpleBin))
	if !a || !b {
		fmt.Println("Problem setting up vms")
	}

	fmt.Println("======== Contract creation ========")
	ret, addrContract, leftOverGas, err := bvm.Create(accountRef, common.Hex2Bytes(simpleBin), 100000000, big.NewInt(0))
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
	getCall, leftOverGas, err := bvm.Call(accountRef, addrContract, get, leftOverGas, big.NewInt(0))
	if err != nil {
		fmt.Println("get unsuccessful")
		fmt.Println(err)
	} else {
		fmt.Println("- Successful balance fetch")
		fmt.Println("The balance of : ", aPublicKey, " is ", common.Bytes2Hex(getCall))
	}

	_, leftOverGas, err = bvm.Call(accountRef, addrContract, send, leftOverGas, big.NewInt(0))
	if err != nil {
		fmt.Println("send unsuccessful")
		fmt.Println(err)
	} else {
		fmt.Println("- Successful send from ", aPublicKey, " to ", bPublicKey)
	}
	get1Call, leftOverGas, err := bvm.Call(accountRef, addrContract, get1, leftOverGas, big.NewInt(0))
	if err != nil {
		fmt.Println("get unsuccessful")
		fmt.Println(err)
		fmt.Println("Left over gas : ", leftOverGas)
	} else {
		fmt.Println("- Successful balance fetch")
		fmt.Println("The balance of :", bPublicKey, " is ", common.Bytes2Hex(get1Call))
	}
	get11Calls, leftOverGas, err := bvm.Call(accountRef, addrContract, get2, leftOverGas, big.NewInt(0))
	if err != nil {
		fmt.Println("get unsuccessful")
		fmt.Println("Left over gas : ", leftOverGas)
		fmt.Println(err)
	} else {
		fmt.Println("- Successful balance fetch")
		fmt.Println("Balance of ", aPublicKey, " is ", common.Bytes2Hex(get11Calls))
	}

	_, leftOverGas, err = bvm.Call(accountRef, addrContract, transferTests, leftOverGas, big.NewInt(0))
	if err != nil {
		fmt.Println("transfer unsuccessful")
		fmt.Println("Left over gas : ", leftOverGas)
		fmt.Println(err)
	} else {
		fmt.Println("Successful transfer")
	}
	get1Call, leftOverGas, err = bvm.Call(accountRef, addrContract, get1, leftOverGas, big.NewInt(0))
	if err != nil {
		fmt.Println("get unsuccessful")
		fmt.Println(err)
		fmt.Println("Left over gas : ", leftOverGas)
	} else {
		fmt.Println("- Successful balance fetch")
		fmt.Println("The balance of :", bPublicKey, " is ", common.Bytes2Hex(get1Call))
	}
	get11Calls, leftOverGas, err = bvm.Call(accountRef, addrContract, get2, leftOverGas, big.NewInt(0))
	if err != nil {
		fmt.Println("get unsuccessful")
		fmt.Println("Left over gas : ", leftOverGas)
		fmt.Println(err)
	} else {
		fmt.Println("- Successful balance fetch")
		fmt.Println("Balance of ", aPublicKey, " is ", common.Bytes2Hex(get11Calls))
	}

}
