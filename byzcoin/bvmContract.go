package byzcoin

import (
	"errors"
	"math/big"
	"reflect"
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
			transaction, addrContract, err := generalArgsParser(inst)
			if err !=nil {
				return nil,nil, err
			}

			//Sending the transaction to the bvm
			addressOfContract := common.HexToAddress(string(addrContract))
			_, _, err = bvm.Call(accountRef, addressOfContract, transaction, 100000000, big.NewInt(0))
			if err != nil {
				return nil, nil, err
			}
			log.LLvl1("Successful method call at address :", addressOfContract.Hex())
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
func createArgumentParser(inst byzcoin.Instruction) (abiPack []byte, contractAddress []byte,  err error) {
	//return generalArgsParser(inst)
	//log.LLvl1("Parsing arguments for create method")
	arguments := inst.Invoke.Args
	if len(arguments)<3{
		log.LLvl1("Please provide at least a contract address, the contract abi and the method name.")
		return nil, nil, err
	}
	//Getting the general arguments needed to call an Ethereum SC method :
	//contract address, abi, name of the method
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
	abi, err := abi.JSON(strings.NewReader(string(abiBuf)))
	if err != nil {
		return  nil, nil, err
	}
	fromBuf := inst.Invoke.Args.Search("from")
	initialSupplyBuf := inst.Invoke.Args.Search("initialSupply")
	initialSupply, err := strconv.ParseUint(string(initialSupplyBuf), 10, 32)
	if err!=nil {
		return nil, nil , err
	}
	transaction, err := abi.Pack(string(methodBuf), initialSupply, common.BytesToAddress(fromBuf))
	return transaction, contractAddressBuf, nil
}



func generalArgsParser(inst byzcoin.Instruction) (abiPack []byte, contractAddress []byte,  err error){
	
	arguments := inst.Invoke.Args
	if len(arguments)<3{
		log.LLvl1("Please provide at least a contract address, the contract abi and the method name.")
		return nil, nil, err
	}
	//Getting the general arguments needed to call an Ethereum SC method :
	//contract address, abi, name of the method
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
	methodName := string(methodBuf)
	log.LLvl1("Parsing of ", methodName, " arguments")
	abi, err := abi.JSON(strings.NewReader(string(abiBuf)))
	if err != nil {
		return  nil, nil, err
	}
	switch len(abi.Methods[methodName].Inputs) {
		case 1:
			log.LLvl1("one argument")
			input0 := abi.Methods[methodName].Inputs[0]
			arg0 := inst.Invoke.Args.Search(input0.Name)
			argC0 := byteSliceToAllConverter(arg0, input0.Type)
			transaction, err := abi.Pack(methodName, argC0)
			if err != nil {
				return nil, nil, err
			}
			return transaction, contractAddressBuf, nil

		case 2:
			log.LLvl1("two arguments")
			input0 := abi.Methods[methodName].Inputs[0]
			arg0 := inst.Invoke.Args.Search(input0.Name)
			argC0 := byteSliceToAllConverter(arg0, input0.Type)

			input1 := abi.Methods[methodName].Inputs[1]
			arg1 := inst.Invoke.Args.Search(input1.Name)
			argC1 := byteSliceToAllConverter(arg1, input1.Type)

			log.LLvl1("the two parsed argument are", argC0, argC1)
			log.LLvl1(reflect.TypeOf(argC0))
			log.LLvl1(reflect.TypeOf(argC1))
			transaction, err := abi.Pack(methodName, argC0, argC1)
			if err != nil {
				return nil, nil, err
			}
			return transaction, contractAddressBuf, nil

		case 3:
			log.LLvl1("three arguments")
			input0 := abi.Methods[methodName].Inputs[0]
			arg0 := inst.Invoke.Args.Search(input0.Name)
			argC0 := byteSliceToAllConverter(arg0, input0.Type)

			input1 := abi.Methods[methodName].Inputs[1]
			arg1 := inst.Invoke.Args.Search(input1.Name)
			argC1 := byteSliceToAllConverter(arg1, input1.Type)

			input2 := abi.Methods[methodName].Inputs[2]
			arg2 := inst.Invoke.Args.Search(input2.Name)
			argC2 := byteSliceToAllConverter(arg2, input2.Type)

			log.LLvl1("the three parsed argument are", argC0, argC1, argC2)
			transaction, err := abi.Pack(methodName, argC0, argC1)
			if err != nil {
				return nil, nil, err
			}
			return transaction, contractAddressBuf, nil

		case 4:
			log.LLvl1("four arguments")
			input0 := abi.Methods[methodName].Inputs[0]
			arg0 := inst.Invoke.Args.Search(input0.Name)
			argC0 := byteSliceToAllConverter(arg0, input0.Type)

			input1 := abi.Methods[methodName].Inputs[1]
			arg1 := inst.Invoke.Args.Search(input1.Name)
			argC1 := byteSliceToAllConverter(arg1, input1.Type)

			input2 := abi.Methods[methodName].Inputs[2]
			arg2 := inst.Invoke.Args.Search(input2.Name)
			argC2 := byteSliceToAllConverter(arg2, input2.Type)

			input3 := abi.Methods[methodName].Inputs[3]
			arg3 := inst.Invoke.Args.Search(input3.Name)
			argC3 := byteSliceToAllConverter(arg3, input3.Type)

			log.LLvl1("the three parsed argument are", argC0, argC1, argC2, argC3)
			transaction, err := abi.Pack(methodName, argC0, argC1)
			if err != nil {
				return nil, nil, err
			}
			return transaction, contractAddressBuf, nil



	}





	return nil, contractAddressBuf, nil
}


func byteSliceToAllConverter(argument []byte, argumentType abi.Type) interface{}{
	log.LLvl1("We have an argument of type :", argumentType.String())
	switch argumentType.String() {
	case "string":
		log.LLvl1("converting a string")
		return string(argument)
	case "address":
		log.LLvl1("converting an address")
		return common.HexToAddress(string(argument))
	case "uint256":
		log.LLvl1("converting a number")
		number, _ := strconv.ParseUint(string(argument), 10, 32)
		log.LLvl1(reflect.TypeOf(number))
		return number
	default:
		log.LLvl1("type of argument not recognized")
		return nil
	}
	return nil
}
