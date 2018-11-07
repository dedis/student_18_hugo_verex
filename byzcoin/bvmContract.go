package byzcoin

import (
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"

	"github.com/ethereum/go-ethereum/core/vm"

	"github.com/ethereum/go-ethereum/common"

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
		//fmt.Println("The mem db", memDB)
		dbBuff, _ := memDB.Dump()
		//fmt.Println("The db buffer", dbBuff)
		instID := inst.DeriveID("")
		scs = []byzcoin.StateChange{
			byzcoin.NewStateChange(byzcoin.Create, instID, ContractBvmID, dbBuff, darcID),
		}
		return

	case byzcoin.InvokeType:
		//create db out of csbuf
		switch inst.Invoke.Command {

		case "deployContract":
			memDBBuff := []byte{}
			cs := NewContractStruct(inst.Invoke.Args)
			memDBBuff, _ = protobuf.Encode(&cs)
			memDB, er := NewMemDatabase(memDBBuff)
			if er != nil {
				log.LLvl1("Problem generating DB")
			}
			bvm, memDB := spawnEvm(memDB)
			bytecode := inst.Invoke.Args[0].Name
			accountRef := vm.AccountRef(common.HexToAddress("0x2afd357E96a3aCbcd01615681C1D7e3398d5fb61"))

			ret, contractAddress, _, err := bvm.Create(accountRef, common.Hex2Bytes(bytecode), 100000000, big.NewInt(0))
			//fmt.Println("The mem db", memDB)
			if err != nil {
				fmt.Println("Contract deployment unsuccessful")
				fmt.Println("Return of contract creation", common.Bytes2Hex(ret))
				fmt.Println(err)
			} else {
				fmt.Println("- Successful contract deployment at", contractAddress.Hex())
				//fmt.Println("- New contract address", contractAddress.Hex())
			}
			memDB.Dump()

		case "callMethod":
			fmt.Println("hey")

			publicKey := common.HexToAddress("0x2afd357E96a3aCbcd01615681C1D7e3398d5fb61")
			accountRef := vm.AccountRef(publicKey)
			memDBBuff := []byte{}
			cs := NewContractStruct(inst.Invoke.Args)
			memDBBuff, _ = protobuf.Encode(&cs)
			memDB, er := NewMemDatabase(memDBBuff)
			if er != nil {
				log.LLvl1("Problem generating DB")
			}
			bvm, memDB := spawnEvm(memDB)
			contractABI := inst.Invoke.Args[0].Name
			methodCall := inst.Invoke.Args[1].Name
			fmt.Println(methodCall)
			addrContract := common.HexToAddress(inst.Invoke.Args[2].Name)
			maxGas := inst.Invoke.Args[3].Name
			u64, err := strconv.ParseUint(maxGas, 10, 64)
			fmt.Println(methodCall)

			abi, err := abi.JSON(strings.NewReader(contractABI))
			if err != nil {
				fmt.Println(err)
			}
			create, err := abi.Pack(methodCall, big.NewInt(4096), publicKey)
			if err != nil {
				fmt.Println(err)
			}

			_, _, err = bvm.Call(accountRef, addrContract, create, u64, big.NewInt(0))
			if err != nil {
				log.LLvl1("Calling", methodCall, "failed")
				log.LLvl1(err)
			} else {
				log.LLvl1("Successful", methodCall, "method call")
			}

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
