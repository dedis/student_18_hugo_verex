package byzcoin

import (
	"errors"
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
		case "deploy":
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
