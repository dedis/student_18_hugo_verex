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
			db, err := getDB1(memDB)
			if err !=nil {
				return nil, nil, err
			}
			log.LLvl1("db", db)
			address := common.HexToAddress(string(addressBuf))
			ret := db.GetBalance(address)
			log.LLvl1("db", db)
			db.Commit(true)
			log.LLvl1("db", db)
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
			db, err := getDB1(memDB)
			if err !=nil {
				return nil, nil, err
			}
			eth, err := strconv.ParseInt(string(value), 10, 64)
			if err !=nil {
				return nil, nil, err
			}
			address := common.HexToAddress(string(addressBuf))
			db.AddBalance(address, big.NewInt(1e9*eth))
			_, err = db.Commit(true)
			if err != nil {
				return nil, nil ,err
			}

			//CreditAccount(db, , eth)
			dbBuf, err := memDB.Dump()
			if err != nil {
				return nil, nil, err
			}
			log.LLvl1("dump", dbBuf)
			scs = []byzcoin.StateChange{
				byzcoin.NewStateChange(byzcoin.Update, inst.InstanceID, ContractBvmID, dbBuf, darcID),
			}

		case "call":
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

		case "transaction":
			memDB, err := NewMemDatabase(memDBBuff)
			if err != nil {
				return nil, nil, err
			}

			dbBuf, err := memDB.Dump()
			if err != nil {
				return nil, nil, err
			}
			toAddressBuf := inst.Invoke.Args.Search("to")
			if toAddressBuf == nil {
				log.LLvl1(err)
				return nil, nil, err
			}
			//var toAddress common.Address
			toAddress := common.HexToAddress(string(toAddressBuf))
			data := inst.Invoke.Args.Search("data")
			if data == nil {
				log.LLvl1("no data provided in transaction")
				data = []byte{}
			}
			tx := types.NewTransaction(0, toAddress, big.NewInt(1), 100000, big.NewInt(2000000), data)
			err = sendTransactionHelper(tx)
			if err != nil {
				return nil, nil, err
			}
			scs = []byzcoin.StateChange{
				byzcoin.NewStateChange(byzcoin.Update, inst.InstanceID, ContractBvmID, dbBuf, darcID),
			}

		}
		//log.LLvl1(err, scs)

		return
	}

	err = errors.New("didn't find any instructions")
	return

}

func sendTransactionHelper(tx *types.Transaction) error{
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

	//address, private := GenerateKeys()
	//Transaction signing
	privateKey := "a33fca62081a2665454fe844a8afbe8e2e02fb66af558e695a79d058f9042f0d"
	private, err := crypto.HexToECDSA(privateKey)
	if err != nil {
		return nil
	}
	address := crypto.PubkeyToAddress(private.PublicKey)
	//CreditAccount(statedb, address, 10000)
	//

	var signer1 types.Signer = types.HomesteadSigner{}
	tx, err = types.SignTx(tx, signer1, private)
	if err != nil {
		return err
	}
	usedGas := uint64(0)
	ug := &usedGas
	receipt, usedGas, err := core.ApplyTransaction(chainconfig, bc, &address, gp, statedb, header, tx, ug, config)
	if err !=nil {
		return err
	}
	log.LLvl1(receipt)
	return nil
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

