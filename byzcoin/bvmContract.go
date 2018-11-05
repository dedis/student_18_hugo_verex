package byzcoin

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"

	"github.com/ethereum/go-ethereum/core/vm"

	"github.com/dedis/cothority/byzcoin"
	"github.com/dedis/cothority/darc"
	"github.com/dedis/protobuf"
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

	var csBuf []byte
	var darcID darc.ID
	csBuf, _, darcID, err = cdb.GetValues(inst.InstanceID.Slice())
	if err != nil {
		return
	}
	var bvm *vm.EVM
	contractsPath := "/Users/hugo/student_18_hugo_verex/contracts/"
	var gas uint64
	var value *big.Int

	switch inst.GetType() {
	case byzcoin.SpawnType:
		emptyData := []byte{}
		bvm = spawnEvm(emptyData)
		cs := NewContractStruct(inst.Spawn.Args)
		var csBuf []byte
		csBuf, err = protobuf.Encode(&cs)
		if err != nil {
			return
		}
		instID := inst.DeriveID("")
		scs = []byzcoin.StateChange{
			byzcoin.NewStateChange(byzcoin.Create, instID, ContractBvmID, csBuf, darcID),
		}
		return

	case byzcoin.InvokeType:
		//create db out of csbuf
		switch inst.Invoke.Command {
		case "createAccount":

		case "deployContract":
			fmt.Println("Deploying contract")
			gas = 10000000000
			value = big.NewInt(0)
			emptyData := []byte{}
			db, err := getDB(emptyData)
			if err != nil {
				fmt.Println(err)
			}
			publicKey := LoadAccount(db)
			accountRef := vm.AccountRef(publicKey)
			_, contractBinary := getSmartContract(contractsPath, "ModifiedToken")
			ret, addrContract, leftOverGas, err := bvm.Create(accountRef, common.Hex2Bytes(contractBinary), gas, value)
			if err != nil {
				fmt.Println("Contract deployment unsuccessful. ", leftOverGas, " left over gas.")
				fmt.Println("Return of contract creation", common.Bytes2Hex(ret))
				fmt.Println(err)
			} else {
				fmt.Println("- Successful contract deployment")
				fmt.Println("- New contract address", addrContract.Hex())
			}

		case "mintCoin":
			fmt.Println("Sending command")

		}

		scs = []byzcoin.StateChange{
			byzcoin.NewStateChange(byzcoin.Update, inst.InstanceID, ContractBvmID, csBuf, darcID),
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
