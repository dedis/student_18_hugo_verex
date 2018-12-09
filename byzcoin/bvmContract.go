package byzcoin

import (
	"errors"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	"math/big"
	"strconv"

	"github.com/ethereum/go-ethereum/common"

	"github.com/dedis/onet/log"

	"github.com/dedis/cothority/byzcoin"
	"github.com/dedis/cothority/darc"
)

//ContractBvmID denotes a contract that can deploy and call an Ethereum virtual machine
var ContractBvmID = "bvm"
var nilAddress = common.HexToAddress("0x0000000000000000000000000000000000000000")
var accountRef = vm.AccountRef(nilAddress)

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

		case "display":
			memDB, err := NewMemDatabase(memDBBuff)
			if err != nil {
				log.LLvl1("problem generating DB")
				return nil, nil, err
			}
			addressBuf := inst.Invoke.Args.Search("address")
			if addressBuf == nil {
				return nil, nil, errors.New("no address provided")
			}
			db, err := getDB(memDB)
			if err !=nil {
				return nil, nil, err
			}
			address := common.HexToAddress(string(addressBuf))
			ret := db.GetBalance(address)
			db.Commit(true)
			//log.LLvl1("db", db)
			if ret == big.NewInt(0) {
				log.LLvl1("object not found")
			}
			log.LLvl1( address.Hex(), "balance", ret)
			return nil, nil, nil


		case "credit":
			memDB, err := NewMemDatabase(memDBBuff)
			if err != nil {
				log.LLvl1("problem generating DB")
				return nil, nil, err
			}
			addressBuf := inst.Invoke.Args.Search("address")
			if addressBuf == nil {
				return nil, nil, errors.New("no address provided")
			}
			value := inst.Invoke.Args.Search("value")
			if value == nil {
				return nil, nil , errors.New("no value provided")
			}
			db, err := getDB(memDB)
			if err !=nil {
				return nil, nil, err
			}
			eth, err := strconv.ParseInt(string(value), 10, 64)
			if err !=nil {
				return nil, nil, err
			}
			address := common.HexToAddress(string(addressBuf))
			db.SetBalance(address, big.NewInt(1e9*eth))
			_, err = db.Commit(true)
			if err != nil {
				return nil, nil ,err
			}
			//CreditAccount(db, , eth)
			dbBuf, err := memDB.Dump()
			if err != nil {
				return nil, nil, err
			}
			//log.LLvl1("dump", dbBuf)
			scs = []byzcoin.StateChange{
				byzcoin.NewStateChange(byzcoin.Update, inst.InstanceID, ContractBvmID, dbBuf, darcID),
			}

		case "deploy":
			memDB, err := NewMemDatabase(memDBBuff)
			if err != nil {
				return nil, nil, err
			}
			gasLimit := uint64(1e18)
			value := big.NewInt(0)
			gasPrice := big.NewInt(0)
			bytecode := inst.Invoke.Args.Search("bytecode")
			if bytecode == nil {
				log.LLvl1("no data provided in transaction")
				return nil, nil, err
			}
			deployTx := types.NewContractCreation(0, value, gasLimit,  gasPrice, bytecode)
			deployReceipt, err := sendTx(deployTx, memDB)
			contractAddress := deployReceipt.ContractAddress
			log.LLvl1("contract deployed at: ", contractAddress.Hex())
			if err != nil {
				return nil, nil, err
			}
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
			gasLimit := uint64(1e18)
			value := big.NewInt(0)
			gasPrice := big.NewInt(0)
			contractAddress := inst.Invoke.Args.Search("contractAddress")
			method := inst.Invoke.Args.Search("method")
			newTx := types.NewTransaction(0, common.HexToAddress(string(contractAddress)), value ,gasLimit, gasPrice, method)
			methodCallReceipt, err := sendTx(newTx, memDB)
			if err != nil {
				log.LLvl1("error minting", err)
				return nil, nil, err
			}
			log.LLvl1("call to contract, gas used", methodCallReceipt.CumulativeGasUsed)
			dbBuf, err := memDB.Dump()
			if err != nil {
				return nil, nil, err
			}
			scs = []byzcoin.StateChange{
				byzcoin.NewStateChange(byzcoin.Update, inst.InstanceID, ContractBvmID, dbBuf, darcID),
			}
		}
		return
	}
	err = errors.New("didn't find any instructions")
	return

}

func sendTx(tx *types.Transaction, memDB *MemDatabase) (*types.Receipt, error){
	chainconfig := getChainConfig()
	statedb, _ := getDB(memDB)
	config := getVMConfig()


	//// GasPool tracks the amount of gas available during execution of the transactions in a block.
	gp := new(core.GasPool).AddGas(uint64(1e18))
	usedGas := uint64(0)
	ug := &usedGas

	// ChainContext supports retrieving headers and consensus parameters from the
	// current blockchain to be used during transaction processing.
	var bc core.ChainContext

	// Header represents a block header in the Ethereum blockchain.
	var header  *types.Header
	header = &types.Header{
		Number: big.NewInt(0),
		Difficulty: big.NewInt(0),
		ParentHash: common.Hash{0},
		Time: big.NewInt(0),
	}


	//Transaction signing
	privateKey := "a33fca62081a2665454fe844a8afbe8e2e02fb66af558e695a79d058f9042f0d"
	private, err := crypto.HexToECDSA(privateKey)
	if err != nil {
		return nil, err
	}
	address := crypto.PubkeyToAddress(private.PublicKey)
	statedb.SetBalance(address, big.NewInt(1e18))
	var signer1 types.Signer = types.HomesteadSigner{}
	tx, err = types.SignTx(tx, signer1, private)
	if err != nil {
		return nil, err
	}
	receipt, usedGas, err := core.ApplyTransaction(chainconfig, bc, &nilAddress, gp, statedb, header, tx, ug, config)
	if err !=nil {
		log.LLvl1("issue applying tx:", err)
		return nil, err
	}
	return receipt, nil
}



