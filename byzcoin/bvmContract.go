package byzcoin

import (
	"errors"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"math/big"

	"github.com/ethereum/go-ethereum/common"

	"github.com/dedis/onet/log"

	"github.com/dedis/cothority/byzcoin"
	"github.com/dedis/cothority/darc"
)

//ContractBvmID denotes a contract that can deploy and call an Ethereum virtual machine
var ContractBvmID = "bvm"

func contractBvm(cdb byzcoin.CollectionView, inst byzcoin.Instruction, cIn []byzcoin.Coin) (scs []byzcoin.StateChange, cOut []byzcoin.Coin, err error) {

	cOut = cIn
	err = inst.VerifyDarcSignature(cdb)
	if err != nil {
		return
	}
	//Data to be stored in the memDB must be kept in the memDBBuffer
	var memDBBuff []byte
	var darcID darc.ID
	memDBBuff, _, darcID, err = cdb.GetValues(inst.InstanceID.Slice())
	if err != nil {
		return
	}

	//Ethereum
	nilAddress := common.HexToAddress("0x0000000000000000000000000000000000000000")
	accountRef := vm.AccountRef(nilAddress)

	switch inst.GetType() {

	case byzcoin.SpawnType:
		memDB, _ := NewMemDatabase([]byte{})
		if err != nil {
			return nil, nil, err
		}
		log.LLvl1("evm was spawned correctly")
		spawnEvm(memDB)
		dbBuff, _ := memDB.Dump()
		instID := inst.DeriveID("")
		scs = []byzcoin.StateChange{
			byzcoin.NewStateChange(byzcoin.Create, instID, ContractBvmID, dbBuff, darcID),
		}
		return

	case byzcoin.InvokeType:
		switch inst.Invoke.Command {
		case "printAccount":

		case "creditAccount":


		case "deploy":
			//remove deploy
			memDB, err := NewMemDatabase(memDBBuff)
			if err != nil {
				log.LLvl1("problem generating DB")
				return nil, nil, err
			}
			bvm, err := spawnEvm(memDB)
			if err != nil {
				log.LLvl1("problem generating EVM")
				return nil, nil, err
			}
			bytecode := inst.Invoke.Args.Search("bytecode")
			if bytecode == nil {
				return nil, nil, errors.New("no bytecode provided")
			}
			//same here
			_, contractAddress, leftOverGas, err := bvm.Create(accountRef, bytecode, 100000000, big.NewInt(0))
			if err != nil {
				return nil, nil, err
			}
			log.LLvl1("successful contract deployment at", contractAddress.Hex(), ". Left over gas: ", leftOverGas)
			dbBuf, err := memDB.Dump()
			if err != nil {
				return nil, nil, err
			}
			scs = []byzcoin.StateChange{
				byzcoin.NewStateChange(byzcoin.Update, inst.InstanceID, ContractBvmID, dbBuf, darcID),
			}
		case "transaction":
			memDB, err := NewMemDatabase(memDBBuff)
			if err != nil {
				return nil, nil, err
			}
			//Instantiation of BVM
			bvm, err := spawnEvm(memDB)
			if err != nil {
				return nil, nil, err
			}

			//Creating the transaction that will be sent to the bvm
			transaction, addrContract, err := getAbiCall(inst)
			if err !=nil {
				return nil,nil, err
			}
			//Sending the transaction to the bvm
			addressOfContract := common.HexToAddress(string(addrContract))
			//send transaction here instead of using evm.Call
			_, leftOverGas, err := bvm.Call(accountRef, addressOfContract, transaction, 100000000, big.NewInt(0))
			if err != nil {
				return nil, nil, err
			}
			log.LLvl1("Successful method call at address :", addressOfContract.Hex(), " left over gas: ", leftOverGas)
			//Saving state changes in DB
			dbBuf, err := memDB.Dump()
			if err != nil {
				return nil, nil, err
			}
			scs = []byzcoin.StateChange{
				byzcoin.NewStateChange(byzcoin.Update, inst.InstanceID, ContractBvmID, dbBuf, darcID),
			}
		}
		//is this call useful?
		scs = []byzcoin.StateChange{
			byzcoin.NewStateChange(byzcoin.Update, inst.InstanceID, ContractBvmID, memDBBuff, darcID),
		}
		return
	}

	err = errors.New("didn't find any instructions")
	return

}

func sendTransactionHelper(author *common.Address){
	chainconfig := getChainConfig()
	memDB, _ := NewMemDatabase([]byte{})
	statedb, _ := getDB(memDB)
	config := getVMConfig()

	//// GasPool tracks the amount of gas available during execution of the transactions in a block.
	gp := new(core.GasPool).AddGas(uint64(800000))

	// ChainContext supports retrieving headers and consensus parameters from the
	// current blockchain to be used during transaction processing.
	var bc core.ChainContext

	// Header represents a block header in the Ethereum blockchain.
	var header  *types.Header
	header = &types.Header{
		Number: big.NewInt(0),
		Difficulty: big.NewInt(0),
		ParentHash: common.Hash{0x00},
		Time: big.NewInt(0),
	}

	//Ethereum transaction
	nilAddress := common.HexToAddress("0x0000000000000000000000000000000000000000")
	var tx *types.Transaction

	address, private := GenerateKeys()
	CreditAccount(statedb, address, 10000)
	tx = types.NewTransaction(0, nilAddress, big.NewInt(1), 100000, big.NewInt(2000000), []byte{})
	//func SignTx(tx *Transaction, s Signer, prv *ecdsa.PrivateKey) (*Transaction, error) {
	//var signer types.Signer = types.FrontierSigner{}
	var signer1 types.Signer = types.HomesteadSigner{}
	//signer1 := types.NewEIP155Signer(big.NewInt(0))
	tx, err := types.SignTx(tx, signer1, private)
	if err != nil {
		log.LLvl1(err)
	}
	usedGas := uint64(0)
	ug := &usedGas
	receipt, usedGas, err := core.ApplyTransaction(chainconfig, bc, &address, gp, statedb, header, tx, ug, config)
	if err !=nil{
		log.LLvl1(err)
	}
	log.LLvl1(receipt)
}

//createArgumentParser creates a transaction for the create method of modifiedToken
func getAbiCall(inst byzcoin.Instruction) (abiPack []byte, contractAddress []byte,  err error) {
	arguments := inst.Invoke.Args
	if len(arguments)<2{
		log.LLvl1("Please provide at least a contract address and the abi call.")
		return nil, nil, err
	}
	contractAddressBuf := inst.Invoke.Args.Search("contractAddress")
	if contractAddressBuf == nil {
		log.LLvl1(err)
		return nil, nil, err
	}
	abiBuf := inst.Invoke.Args.Search("abiCall")
	if abiBuf == nil {
		log.LLvl1(err)
		return nil, nil, err
	}

	return abiBuf, contractAddressBuf, nil
}

