package byzcoin

import (
	"github.com/stretchr/testify/require"
	"math/big"
	"strings"
	"testing"

	"github.com/dedis/onet/log"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
)

func TestTokenContract(t *testing.T) {

	canTransfer := func(vm.StateDB, common.Address, *big.Int) bool {
		//log.Println("Verified transfer")
		return true
	}

	transfer := func(vm.StateDB, common.Address, common.Address, *big.Int) {
		log.Lvl3("tried to transfer")
	}

	gethash := func(uint64) common.Hash {
		log.Lvl3("tried to get hash")
		return common.HexToHash("0x0000000000000000000000000000000000000000")
	}

	contractsPath := "/Users/hugo/student_18_hugo_verex/contracts/ModifiedToken/"
	log.LLvl1("test: evm creation and function calls")
	simpleAbi, simpleBin := getSmartContract(contractsPath, "ModifiedToken")


	aPublicKey := common.HexToAddress("0x1111111111111111111111111111111111111111")
	bPublicKey := common.HexToAddress("0x2222222222222222222222222222222222222222")

	//aPublicKey, _ := GenerateKeys()
	//bPublicKey, _ := GenerateKeys()

	accountRef := vm.AccountRef(common.HexToAddress("0x0000000000000000000000000000000000000000"))


	//Cration of abi object
	abi, err := abi.JSON(strings.NewReader(simpleAbi))
	require.Nil(t, err)
	//Helper functions for contract call
	create, err := abi.Pack("create", uint64(12), aPublicKey)
	require.Nil(t, err)
	get, err := abi.Pack("getBalance", aPublicKey)
	require.Nil(t, err)
	send, err := abi.Pack("transfer", aPublicKey, bPublicKey,uint32(1))
	require.Nil(t, err)
	get1, err := abi.Pack("getBalance", bPublicKey)
	require.Nil(t, err)
	get2, err := abi.Pack("getBalance", aPublicKey)
	require.Nil(t, err)
	transferTests, err := abi.Pack("transfer", aPublicKey, bPublicKey, uint32(1))
	require.Nil(t, err)



	//Various setups
	log.Lvl3("DB setup")
	emptyData := []byte{}
	memDb, _ := NewMemDatabase(emptyData)
	sdb, _ := getDB(memDb)
	//sdb.SetBalance(aPublicKey, big.NewInt(1000000000000))
	log.Lvl2("Setting up context")
	ctx := vm.Context{CanTransfer: canTransfer, Transfer: transfer, GetHash: gethash, Origin: aPublicKey, GasPrice: big.NewInt(1), Coinbase: aPublicKey, GasLimit: 10000000000, BlockNumber: big.NewInt(0), Time: big.NewInt(1), Difficulty: big.NewInt(1)}
	log.Lvl2("Setting up & checking VMs")
	bvm := vm.NewEVM(ctx, sdb, getChainConfig(), getVMConfig())
	bvmInterpreter := vm.NewEVMInterpreter(bvm, getVMConfig())
	a := bvmInterpreter.CanRun(common.Hex2Bytes(simpleBin))
	if !a {
		log.Lvl1("Problem setting up VM")
	}

	//Contract deployment
	log.LLvl1("contract creation")
	retContractCreation, addrContract, leftOverGas, err := bvm.Create(accountRef, common.Hex2Bytes(simpleBin), 100000000, big.NewInt(0))
	if err != nil {
		log.LLvl1("contract deployment unsuccessful")
		log.LLvl1("return of contract creation", common.Bytes2Hex(retContractCreation))
		log.Lvl1(err)
	}
	log.LLvl1("successful contract deployment")
	log.LLvl1("new contract address", addrContract.Hex())

	//Contract calls
	log.Lvl3("Contract calls")


	//Call to constructor of contract
	_, _, err = bvm.Call(accountRef, addrContract, create, leftOverGas, big.NewInt(0))
	require.Nil(t, err)
	log.Lvl3("successful token creation")

	//Checking if the account specified received all the tokens
	retBalanceOfAccountA, _, err := bvm.Call(accountRef, addrContract, get, leftOverGas, big.NewInt(0))
	require.Nil(t, err)
	log.Lvl3("successful balance fetch")
	log.Lvl1("the balance of : ", aPublicKey.Hex(), " is ", retBalanceOfAccountA)

	//Sending a token from A to B
	_, _, err = bvm.Call(accountRef, addrContract, send, leftOverGas, big.NewInt(0))
	require.Nil(t, err)
	log.Lvl1("successful send from ", aPublicKey.Hex(), " to ", bPublicKey.Hex())

	//Checking balance of account B
	retBalanceOfAccountB, _, err := bvm.Call(accountRef, addrContract, get1, leftOverGas, big.NewInt(0))
	require.Nil(t, err)
	log.Lvl3(("successful balance fetch"))
	log.Lvl1("the balance of :", bPublicKey.Hex(), " is ", retBalanceOfAccountB)

	//Checking if the other account was updated accordingly
	retBalanceOfAccountA, _, err = bvm.Call(accountRef, addrContract, get2, leftOverGas, big.NewInt(0))
	require.Nil(t, err)
	log.Lvl3("successful balance fetch")
	log.Lvl1("balance of ", aPublicKey.Hex(), " is ", retBalanceOfAccountA)

	//Trying to transfer
	_, _, err = bvm.Call(accountRef, addrContract, transferTests, leftOverGas, big.NewInt(0))
	require.Nil(t, err)
	log.Lvl1("successful transfer")

	retBalanceOfAccountB, _, err = bvm.Call(accountRef, addrContract, get1, leftOverGas, big.NewInt(0))
	require.Nil(t, err)
	log.Lvl3("successful balance fetch")
	log.Lvl1("the balance of :", bPublicKey.Hex(), " is ", retBalanceOfAccountB)

	retBalanceOfAccountA, _, err = bvm.Call(accountRef, addrContract, get2, leftOverGas, big.NewInt(0))
	require.Nil(t, err)
	log.Lvl1("successful balance fetch")
	log.Lvl1("balance of ", aPublicKey.Hex(), " is ", retBalanceOfAccountA)
	log.LLvl1("contract calls passed")
}



