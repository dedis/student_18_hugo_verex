package byzcoin

import (
	"math/big"
	"strings"
	"testing"

	"github.com/dedis/onet/log"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
)

func TestTokenContract(t *testing.T) {
	contractsPath := "/Users/hugo/student_18_hugo_verex/contracts/"
	log.LLvl1("evm testing")
	simpleAbi, simpleBin := getSmartContract(contractsPath, "ModifiedToken")
	aPublicKey, _ := GenerateKeys()
	bPublicKey, _ := GenerateKeys()
	accountRef := vm.AccountRef(aPublicKey)
	abi, err := abi.JSON(strings.NewReader(simpleAbi))
	if err != nil {
		log.Lvl1(err)
	}
	create, err := abi.Pack("create", big.NewInt(4096), aPublicKey)
	if err != nil {
		log.Lvl1(err)
	}
	get, err := abi.Pack("getBalance", aPublicKey)
	if err != nil {
		log.Lvl1(err)
	}

	send, err := abi.Pack("transfer", aPublicKey, bPublicKey, big.NewInt(16))
	if err != nil {
		log.Lvl1(err)
	}

	get1, err := abi.Pack("getBalance", bPublicKey)
	if err != nil {
		log.Lvl1(err)
	}
	get2, err := abi.Pack("getBalance", aPublicKey)
	if err != nil {
		log.Lvl1(err)
	}

	transferTests, err := abi.Pack("transfer", aPublicKey, bPublicKey, big.NewInt(16))
	if err != nil {
		log.Lvl1(err)
	}

	log.Lvl1("DB setup")
	emptyData := []byte{}
	memDb, _ := NewMemDatabase(emptyData)
	sdb, _ := getDB(memDb)

	canTransfer := func(vm.StateDB, common.Address, *big.Int) bool {
		//log.Println("Verified transfer")
		return true
	}

	transfer := func(vm.StateDB, common.Address, common.Address, *big.Int) {
		log.Lvl1("tried to transfer")
	}

	gethash := func(uint64) common.Hash {
		log.Lvl1("tried to get hash")
		return common.HexToHash("0x0000000000000000000000000000000000000000")
	}

	sdb.SetBalance(aPublicKey, big.NewInt(1000000000000))

	log.Lvl1("Setting up context")
	ctx := vm.Context{CanTransfer: canTransfer, Transfer: transfer, GetHash: gethash, Origin: aPublicKey, GasPrice: big.NewInt(1), Coinbase: aPublicKey, GasLimit: 10000000000, BlockNumber: big.NewInt(0), Time: big.NewInt(1), Difficulty: big.NewInt(1)}

	log.Lvl1("Setting up & checking VMs")
	bvm := vm.NewEVM(ctx, sdb, getChainConfig(), getVMConfig())

	bvmInterpreter := vm.NewEVMInterpreter(bvm, getVMConfig())

	a := bvmInterpreter.CanRun(common.Hex2Bytes(simpleBin))

	if !a {
		log.Lvl1("Problem setting up VM")
	}
	log.LLvl1("contract creation")
	ret, addrContract, leftOverGas, err := bvm.Create(accountRef, common.Hex2Bytes(simpleBin), 100000000, big.NewInt(0))
	if err != nil {
		log.LLvl1("contract deployment unsuccessful")
		log.LLvl1("return of contract creation", common.Bytes2Hex(ret))
		log.Lvl1(err)
	} else {
		log.LLvl1("successful contract deployment")
		log.LLvl1("new contract address", addrContract.Hex())
	}

	log.LLvl1("some contract calls")
	_, leftOverGas, err = bvm.Call(accountRef, addrContract, create, leftOverGas, big.NewInt(0))
	if err != nil {
		log.LLvl1("token creation unsuccessful")
		log.LLvl1(err)
	} else {
		log.Lvl1("successful token creation")
	}
	getCall, leftOverGas, err := bvm.Call(accountRef, addrContract, get, leftOverGas, big.NewInt(0))
	if err != nil {
		log.Lvl1("get unsuccessful")
		log.Lvl1(err)
	} else {
		log.Lvl1("successful balance fetch")
		log.Lvl1("the balance of : ", aPublicKey.Hex(), " is ", common.Bytes2Hex(getCall))
	}

	_, leftOverGas, err = bvm.Call(accountRef, addrContract, send, leftOverGas, big.NewInt(0))
	if err != nil {
		log.Lvl1("send unsuccessful")
		log.Lvl1(err)
	} else {
		log.Lvl1("successful send from ", aPublicKey.Hex(), " to ", bPublicKey)
	}
	get1Call, leftOverGas, err := bvm.Call(accountRef, addrContract, get1, leftOverGas, big.NewInt(0))
	if err != nil {
		log.Lvl1("get unsuccessful")
		log.Lvl1(err)
		log.Lvl1("left over gas : ", leftOverGas)
	} else {
		log.Lvl1(("successful balance fetch"))
		log.Lvl1("the balance of :", bPublicKey.Hex(), " is ", common.Bytes2Hex(get1Call))
	}
	get11Calls, leftOverGas, err := bvm.Call(accountRef, addrContract, get2, leftOverGas, big.NewInt(0))
	if err != nil {
		log.Lvl1("get unsuccessful")
		log.Lvl1("Left over gas : ", leftOverGas)
		log.Lvl1(err)
	} else {
		log.Lvl1("successful balance fetch")
		log.Lvl1("balance of ", aPublicKey.Hex(), " is ", common.Bytes2Hex(get11Calls))
	}

	_, leftOverGas, err = bvm.Call(accountRef, addrContract, transferTests, leftOverGas, big.NewInt(0))
	if err != nil {
		log.Lvl1("transfer unsuccessful")
		log.Lvl1("left over gas : ", leftOverGas)
		log.Lvl1(err)
	} else {
		log.Lvl1("successful transfer")
	}
	get1Call, leftOverGas, err = bvm.Call(accountRef, addrContract, get1, leftOverGas, big.NewInt(0))
	if err != nil {
		log.Lvl1("get unsuccessful")
		log.Lvl1(err)
		log.Lvl1("leftover gas : ", leftOverGas)
	} else {
		log.Lvl1("successful balance fetch")
		log.Lvl1("the balance of :", bPublicKey.Hex(), " is ", common.Bytes2Hex(get1Call))
	}
	get11Calls, leftOverGas, err = bvm.Call(accountRef, addrContract, get2, leftOverGas, big.NewInt(0))
	if err != nil {
		log.Lvl1("get unsuccessful")
		log.Lvl1("Left over gas : ", leftOverGas)
		log.Lvl1(err)
	} else {
		log.Lvl1("successful balance fetch")
		log.Lvl1("balance of ", aPublicKey.Hex(), " is ", common.Bytes2Hex(get11Calls))
	}
	log.LLvl1("contract calls passed")
	log.LLvl1("end of evm testing")
}
