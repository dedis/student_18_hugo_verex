package byzcoin

import (
	"errors"
	"math/big"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/core/vm"

	"github.com/ethereum/go-ethereum/accounts/abi"

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
	publicKey := common.HexToAddress("0x0000000000000000000000000000000000000000")
	accountRef := vm.AccountRef(publicKey)

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
		//create db out of csbuf
		switch inst.Invoke.Command {
		case "creditAccount":
			memDB, err := NewMemDatabase(memDBBuff)
			if err != nil {
				log.LLvl1("problem generating DB")
				return nil, nil, err
			}
			pubBuf := inst.Invoke.Args.Search("publicKey")
			if pubBuf == nil {
				return nil, nil, errors.New("no public key provided")
			}
			valueBuf := inst.Invoke.Args.Search("value")
			if valueBuf == nil {
				return nil, nil, err
			}
			value, err := strconv.ParseInt(string(valueBuf), 10, 64)
			if err != nil {
				return nil, nil, err
			}
			db, err := getDB(memDB)
			if err != nil {
				return nil, nil, err
			}
			CreditAccount(db, common.BytesToAddress(pubBuf), value)
			log.LLvl1("credited account", string(pubBuf), " with ", value, " eth")
			dbBuf, err := memDB.Dump()
			if err != nil {
				return nil, nil, err
			}
			scs = []byzcoin.StateChange{
				byzcoin.NewStateChange(byzcoin.Update, inst.InstanceID, ContractBvmID, dbBuf, darcID),
			}

		case "deployContract":
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
			_, contractAddress, _, err := bvm.Create(accountRef, bytecode, 100000000, big.NewInt(0))
			if err != nil {
				return nil, nil, err
			}
			log.LLvl1("successful contract deployment at", contractAddress.Hex())
			dbBuf, err := memDB.Dump()
			if err != nil {
				return nil, nil, err
			}
			scs = []byzcoin.StateChange{
				byzcoin.NewStateChange(byzcoin.Update, inst.InstanceID, ContractBvmID, dbBuf, darcID),
			}

		case "callMethod":
			memDB, err := NewMemDatabase(memDBBuff)
			if err != nil {
				return nil, nil, err
			}
			pubBuf := inst.Invoke.Args.Search("publicKey")
			if pubBuf == nil {
				return nil, nil, errors.New("no public key provided")
			}
			bvm, err := spawnEvm(memDB)
			if err != nil {
				return nil, nil, err
			}
			contractAddressBuf := inst.Invoke.Args.Search("contractAddress")
			if contractAddressBuf == nil {
				log.LLvl1(err)
				return nil, nil, err
			}
			abiBuf := inst.Invoke.Args.Search("abi")
			if abiBuf == nil {
				log.LLvl1(err)
				return nil, nil, err
			}
			methodBuf := inst.Invoke.Args.Search("method")
			if methodBuf == nil {
				log.LLvl1(err)
				return nil, nil, err
			}
			gasBuf := inst.Invoke.Args.Search("gas")
			if gasBuf == nil {
				log.LLvl1(err)
				return nil, nil, err
			}
			addrContract := common.BytesToAddress(inst.Invoke.Args.Search("contractAddress"))
			gas, err := strconv.ParseUint(string(gasBuf), 10, 64)
			if err != nil {
				return nil, nil, err
			}
			abi, err := abi.JSON(strings.NewReader(string(abiBuf)))
			if err != nil {
				return nil, nil, err
			}
			create, err := abi.Pack(string(methodBuf), big.NewInt(45), common.BytesToAddress(contractAddressBuf))
			if err != nil {
				return nil, nil, err
			}
			_, _, err = bvm.Call(accountRef, addrContract, create, gas, big.NewInt(0))
			if err != nil {
				return nil, nil, err

			}
			log.LLvl1("Successful", string(methodBuf), "method call at address :", string(contractAddressBuf))
		}

		scs = []byzcoin.StateChange{
			byzcoin.NewStateChange(byzcoin.Update, inst.InstanceID, ContractBvmID, memDBBuff, darcID),
		}
		return
	}

	err = errors.New("didn't find any instructions")
	return

}
