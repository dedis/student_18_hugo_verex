package byzcoin

import (
	"errors"
	"github.com/dedis/cothority/byzcoin"
	"github.com/dedis/cothority/darc"
	"github.com/dedis/protobuf"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"math/big"
	"strconv"

	"github.com/dedis/onet/log"
)


var ContractBvmID = "bvm"
var nilAddress = common.HexToAddress("0x0000000000000000000000000000000000000000")

type contractBvm struct {
	byzcoin.BasicContract
	ES
}

func contractBvmFromBytes(in []byte) (byzcoin.Contract, error) {
	cv := &contractBvm{}
	err := protobuf.Decode(in, &cv.ES)
	if err != nil {
		return nil, err
	}
	return cv, nil
}

func (c *contractBvm) Spawn(rst byzcoin.ReadOnlyStateTrie, inst byzcoin.Instruction, coins []byzcoin.Coin) (sc []byzcoin.StateChange, cout []byzcoin.Coin, err error) {
	cout = coins
	es := c.ES
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
		byzcoin.NewStateChange(byzcoin.Create, inst.DeriveID(""), ContractBvmID, esBuf, darc.ID(inst.InstanceID.Slice())),
	}
	return
}

func (c *contractBvm) Invoke(rst byzcoin.ReadOnlyStateTrie, inst byzcoin.Instruction, coins []byzcoin.Coin) (sc []byzcoin.StateChange, cout []byzcoin.Coin, err error) {
	cout = coins
	var darcID darc.ID
	_, _, _, darcID, err = rst.GetValues(inst.InstanceID.Slice())
	if err != nil {
		return
	}
	es := c.ES
	switch inst.Invoke.Command {
	case "display":
		log.Lvl2("displaying account value")
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
			log.LLvl1(address.Hex(), "balance empty")
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
		if err != nil {
			return nil, nil, err
		}
		db.SetBalance(address, big.NewInt(1e18*eth))
		log.LLvl1(address.Hex(), "balance credited", db.GetBalance(address), "wei")
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
		log.LLvl1("tx status:", transactionReceipt.Status, "- 0 failed, 1 successful.")
		log.LLvl1("tx receipt:", transactionReceipt.TxHash.Hex())
		log.LLvl1("cumulative gas used:", transactionReceipt.CumulativeGasUsed, transactionReceipt.GasUsed)
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
	default :
		err = errors.New("Contract can only display, credit and receive transactions")
		return

	}
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



type ES struct {
	DbBuf []byte
	RootHash common.Hash
}

