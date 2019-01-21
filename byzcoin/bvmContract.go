package byzcoin

import (
	"errors"
	"github.com/dedis/protobuf"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"math/big"
	"strconv"

	"github.com/ethereum/go-ethereum/common"

	"github.com/dedis/onet/log"

	"github.com/dedis/cothority/byzcoin"
	"github.com/dedis/cothority/darc"
)

// The value contract can simply store a value in an instance and serves
// mainly as a template for other contracts. It helps show the possibilities
// of the contracts and how to use them at a very simple example.

// ContractKeyValueID denotes a contract that can store and update
// key/value pairs.
var ContractBvmID = "bvm"
var nilAddress = common.HexToAddress("0x0000000000000000000000000000000000000000")
//var accountRef = vm.AccountRef(nilAddress)


var es EVMStruct
var EVMStructBuf []byte
var darcID darc.ID



type contractBvm struct {
	Bvm
}

/*
func contractBvmFromBytes(in []byte) (byzcoin.Contract, error) {
	cv := &contractBvm{}
	err := protobuf.Decode(in, &cv.Bvm)
	if err != nil {
		return nil, err
	}
	return cv, nil
}*/


// ContractKeyValue is a simple key/value storage where you
// can put any data inside as wished.
// It can spawn new keyValue instances and will store all the arguments in
// the data field.
// Existing keyValue instances can be "update"d and deleted.
func (c *contractBvm) Spawn(rst byzcoin.ReadOnlyStateTrie, inst byzcoin.Instruction, coins []byzcoin.Coin) (sc []byzcoin.StateChange, cout []byzcoin.Coin, err error) {
	cout = coins
	memdb, db, _, err := spawnEvm()
	if err != nil{
		return nil, nil, err
	}
	es.RootHash, err = db.Commit(true)
	if err != nil {
		return nil, nil, err
	}
	err = db.Database().TrieDB().Commit(es.RootHash, true)
	if err != nil {
		return nil, nil, err
	}
	es.DbBuf, err = memdb.Dump()
	esBuf, err := protobuf.Encode(&es)
	// Then create a StateChange request with the data of the instance. The
	// InstanceID is given by the DeriveID method of the instruction that allows
	// to create multiple instanceIDs out of a given instruction in a pseudo-
	// random way that will be the same for all nodes.
	sc = []byzcoin.StateChange{
		byzcoin.NewStateChange(byzcoin.Create, inst.DeriveID(""), ContractBvmID, esBuf, darcID),
	}
	return
}

func (c *contractBvm) Invoke(rst byzcoin.ReadOnlyStateTrie, inst byzcoin.Instruction, coins []byzcoin.Coin) (sc []byzcoin.StateChange, cout []byzcoin.Coin, err error) {
	cout = coins
	var darcID darc.ID
	EVMStructBuf, _, _, darcID, err = rst.GetValues(inst.InstanceID.Slice())
	if err != nil {
		return
	}
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
		address := common.HexToAddress(string(addressBuf))
		_, db, err := getDB(es)
		if err !=nil {
			return nil, nil, err
		}
		ret := db.GetBalance(address)
		if ret == big.NewInt(0) {
			log.LLvl1("balance empty")
		}
		log.LLvl1( address.Hex(), "balance", ret)
		return nil, nil, nil

	case "credit":
		addressBuf := inst.Invoke.Args.Search("address")
		if addressBuf == nil {
			return nil, nil, errors.New("no address provided")
		}
		address := common.HexToAddress(string(addressBuf))
		value := inst.Invoke.Args.Search("value")
		if value == nil {
			return nil, nil , errors.New("no value provided")
		}
		eth, err := strconv.ParseInt(string(value), 10, 64)
		if err !=nil {
			return nil, nil, err
		}
		memdb, db, err := getDB(es)
		if err !=nil {
			return nil, nil, err
		}
		db.SetBalance(address, big.NewInt(1*eth))
		log.LLvl1(address.Hex(), "balance set", db.GetBalance(address))
		es.RootHash, err = db.Commit(true)
		if err != nil {
			return nil, nil ,err
		}
		err = db.Database().TrieDB().Commit(es.RootHash, true)
		if err != nil {
			return nil, nil, err
		}
		es.DbBuf, err = memdb.Dump()
		if err != nil {
			return nil, nil, err
		}
		esBuf, err := protobuf.Encode(&es)
		if err != nil {
			return nil, nil , err
		}
		sc = []byzcoin.StateChange{
			byzcoin.NewStateChange(byzcoin.Update, inst.InstanceID,
				ContractBvmID, esBuf, darcID),
		}

	case "transaction":
		memdb, db, err := getDB(es)
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
		log.LLvl1("tx receipt:", transactionReceipt.TxHash.Hex())
		if transactionReceipt.ContractAddress.Hex() != nilAddress.Hex() {
			log.LLvl1("contract deployed at:", transactionReceipt.ContractAddress.Hex())
		}
		es.RootHash, err = db.Commit(true)
		if err != nil {
			return nil, nil, err
		}
		err = db.Database().TrieDB().Commit(es.RootHash, true)
		if err != nil {
			return nil, nil, err
		}
		es.DbBuf, err = memdb.Dump()
		if err != nil {
			return nil, nil, err
		}
		esBuf, err := protobuf.Encode(&es)
		if err != nil {
			return nil, nil , err
		}
		sc = []byzcoin.StateChange{
			byzcoin.NewStateChange(byzcoin.Update, inst.InstanceID,
				ContractBvmID, esBuf, darcID),
		}

	}
	return
}


/*
func (c *contractBvm) Delete(rst byzcoin.ReadOnlyStateTrie, inst byzcoin.Instruction, coins []byzcoin.Coin) (sc []byzcoin.StateChange, cout []byzcoin.Coin, err error) {
	cout = coins

	var darcID darc.ID
	_, _, _, darcID, err = rst.GetValues(inst.InstanceID.Slice())
	if err != nil {
		return
	}
	sc = byzcoin.StateChanges{
		byzcoin.NewStateChange(byzcoin.Remove, inst.InstanceID, ContractBvmID, nil, darcID),
	}
	return
}

*/


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

