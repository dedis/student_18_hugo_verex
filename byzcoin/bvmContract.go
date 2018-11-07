package byzcoin

import (
	"errors"
	"fmt"

	"github.com/dedis/protobuf"

	"github.com/dedis/onet/log"

	"github.com/dedis/cothority/byzcoin"
	"github.com/dedis/cothority/darc"
)

/*
spawn: create a new EVM and initialize the database structure
invoke:createAccount
invoke:sendCommand
invoke:mintCoins - send that many coins directly to the account, out of nowhere
*/

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

	switch inst.GetType() {
	case byzcoin.SpawnType:
		fmt.Println("Spawning..")
		//var bvm *vm.EVM
		memDBBuff := []byte{}
		cs := NewContractStruct(inst.Spawn.Args)
		memDBBuff, _ = protobuf.Encode(&cs)
		memDB, er := NewMemDatabase(memDBBuff)
		if er != nil {
			log.LLvl1("Problem generating DB")
		}
		_, memDB = spawnEvm(memDB)

		fmt.Println("The mem db", memDB)

		dbBuff, _ := memDB.Dump()
		fmt.Println("The db buffer", dbBuff)
		fmt.Println("And stop here")

		//TO DO : memDb.Dump
		//cs := NewContractStruct(inst.Spawn.Args)
		//var memDBBuff []byte
		//memDBBuff, err = protobuf.Encode(&cs)

		instID := inst.DeriveID("")
		scs = []byzcoin.StateChange{
			byzcoin.NewStateChange(byzcoin.Create, instID, ContractBvmID, memDBBuff, darcID),
		}
		return

	case byzcoin.InvokeType:
		//create db out of csbuf
		switch inst.Invoke.Command {

		case "createAccount":

		case "deployContract":

		case "mintCoin":

		}

		scs = []byzcoin.StateChange{
			byzcoin.NewStateChange(byzcoin.Update, inst.InstanceID, ContractBvmID, memDBBuff, darcID),
		}
		return
	}

	err = errors.New("didn't find any instructions")
	return

}

//NewContractStruct :
func NewContractStruct(args byzcoin.Arguments) KeyValueData {
	cs := KeyValueData{}
	for _, kv := range args {
		cs.Storage = append(cs.Storage, KeyValue{kv.Name, kv.Value})
	}
	return cs

}
