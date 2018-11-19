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
			transaction, addrContract, err := createArgumentParser(inst)
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
	log.LLvl1("ABI DISSECTING")
	log.LLvl1(abi.Methods["create"].Inputs)
	log.LLvl1(abi.Methods["balanceOf"].Inputs)
	log.LLvl1("unicorn", abi.Methods["unicorn"])

	log.LLvl1("size of args",len(abi.Methods["create"].Inputs))

	log.LLvl1(reflect.TypeOf(abi.Methods["create"]))
	log.LLvl1(abi.Constructor.Inputs)

	transaction, err := abi.Pack(string(methodBuf), initialSupply, common.BytesToAddress(fromBuf))
	return transaction, contractAddressBuf, nil
}

//WARNING : incomplete
//argumentParser parses an arbitrary number of arguments and creates a transaction for an arbitrary method call
func argumentParser(inst byzcoin.Instruction) (abiPack []byte, contractAddress []byte,  err error) {

	log.LLvl1("Parsing arguments and creating transaction")
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
	//parsing the arguments in order
	var leftLength = len(arguments) - 3
	leftArgs := make([][]byte, leftLength)

	if leftLength == 0 {
		transaction, err := abi.Pack(string(methodBuf), big.NewInt(45), common.BytesToAddress(contractAddressBuf))
		if err != nil {
			return  nil, nil, err
		}
		log.LLvl1("Only three arguments were provided, creating transaction")
		return transaction, contractAddressBuf, nil
	}

	for i:=0; i<leftLength; i++ {
		leftArgs[i] = inst.Invoke.Args[3+i].Value
	}

	return  nil, contractAddressBuf,nil
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
	abi, err := abi.JSON(strings.NewReader(string(abiBuf)))
	if err != nil {
		return  nil, nil, err
	}


	var leftLength = len(arguments) - 3
	log.LLvl1(leftLength)
	leftArgsBuf := make([][]byte, leftLength)
	log.LLvl1(reflect.TypeOf(leftArgsBuf))
	leftArgsType := make([]string, leftLength)

	for i:=0; i<leftLength; i++ {
		log.LLvl1(i)
		leftArgsBuf[i] = arguments[3+i].Value
		log.LLvl1("argument buffer", leftArgsBuf)
		typeOfArg := arguments[3+i].Name
		log.LLvl1(typeOfArg)
		s := strings.Split(typeOfArg, ":")
		log.LLvl1(s)
		_, argType := s[0], s[1]
		leftArgsType[i] = argType
		log.LLvl1("We have the value", leftArgsBuf[i], " which is ", argType)

	}
	fromBuf := []byte{}
	initialSupply := 0

	common.BytesToAddress(fromBuf)
	transaction, err := abi.Pack(string(methodBuf), initialSupply)
	log.LLvl1(transaction)


	return transaction, contractAddressBuf, nil
}
