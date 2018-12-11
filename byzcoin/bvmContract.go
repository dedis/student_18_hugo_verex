package byzcoin

import (
	"errors"
	"github.com/dedis/protobuf"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
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
	var es EVMStruct
	var darcID darc.ID
	var EVMStructBuf []byte

	EVMStructBuf, _, darcID, err = cdb.GetValues(inst.InstanceID.Slice())
	if err != nil {
		return
	}




	switch inst.GetType() {

	case byzcoin.SpawnType:
		log.LLvl1("evm was spawned correctly")
		db, _, err := spawnEvm()
		if err != nil{
			return nil, nil, err
		}
		es.RootHash, err = db.Commit(true)
		if err != nil {
			return nil, nil, err
		}
		es.DbBuf = db.Dump()
		esBuf, err := protobuf.Encode(&es)
		instID := inst.DeriveID("")

		scs = []byzcoin.StateChange{
			byzcoin.NewStateChange(byzcoin.Create, instID, ContractBvmID, esBuf, darcID),
		}
		return  scs, cOut, nil

	case byzcoin.InvokeType:

		err = protobuf.Decode(EVMStructBuf, &es)
		if err != nil {
			log.LLvl1(err, EVMStructBuf)
			return
		}
		switch inst.Invoke.Command {

		case "display":
			addressBuf := inst.Invoke.Args.Search("address")
			if addressBuf == nil {
				return nil, nil, errors.New("no address provided")
			}
			db, err := getDB(es)
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
			addressBuf := inst.Invoke.Args.Search("address")
			if addressBuf == nil {
				return nil, nil, errors.New("no address provided")
			}
			value := inst.Invoke.Args.Search("value")
			if value == nil {
				return nil, nil , errors.New("no value provided")
			}
			db, err := getDB(es)
			if err !=nil {
				log.Error()
				return nil, nil, err
			}
			eth, err := strconv.ParseInt(string(value), 10, 64)
			if err !=nil {
				return nil, nil, err
			}
			address := common.HexToAddress(string(addressBuf))
			db.SetBalance(address, big.NewInt(1*eth))
			log.LLvl1(address.Hex(), "balance set", db.GetBalance(address))
			es.RootHash, err = db.Commit(true)
			if err != nil {
				log.Error()
				return nil, nil ,err
			}
			es.DbBuf = db.Dump()
			esBuf, err := protobuf.Encode(&es)
			if err != nil {
				log.Error()
				return nil, nil , err
			}
			scs = []byzcoin.StateChange{
				byzcoin.NewStateChange(byzcoin.Update, inst.InstanceID, ContractBvmID, esBuf, darcID),
			}

		case "transaction":
			db, err := getDB(es)
			if err != nil{
				return nil, nil, err
			}
			txBuffer := inst.Invoke.Args.Search("tx")
			if txBuffer == nil {
				log.LLvl1("no transaction provided in byzcoin transaction")
				return nil, nil, err
			}
			var ethTx types.Transaction
			err = ethTx.UnmarshalJSON(txBuffer)
			if err != nil {
				return nil, nil, err
			}
			transactionReceipt, err := sendTx(&ethTx, db)
			if err != nil {
				log.LLvl1("error issuing transaction:", err)
				return nil, nil, err
			}
			log.LLvl1("successful new transaction:", transactionReceipt.TxHash.Hex())
			log.LLvl1("gas used:", transactionReceipt.CumulativeGasUsed)
			if transactionReceipt.ContractAddress.Hex() != nilAddress.Hex() {
				log.LLvl1("contract deployed at:", transactionReceipt.ContractAddress.Hex())
			}
			es.RootHash, _ = db.Commit(true)
			es.DbBuf = db.Dump()
			esBuf, err := protobuf.Encode(&es)
			if err != nil {
				return nil, nil , err
			}
			scs = []byzcoin.StateChange{
				byzcoin.NewStateChange(byzcoin.Update, inst.InstanceID, ContractBvmID, esBuf, darcID),
			}

		}
		return
	}
	err = errors.New("didn't find any instructions")
	return

}

func sendTx(tx *types.Transaction, db *state.StateDB) (*types.Receipt, error){
	chainconfig := getChainConfig()
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

	receipt, usedGas, err := core.ApplyTransaction(chainconfig, bc, &nilAddress, gp, db, header, tx, ug, config)
	if err !=nil {
		log.Error()
		return nil, err
	}
	return receipt, nil
}


type EVMStruct struct {
	DbBuf []byte
	RootHash common.Hash
}



