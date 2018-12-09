package byzcoin

import (
	"github.com/ethereum/go-ethereum/ethdb"
	"io/ioutil"
	"math/big"

	"github.com/dedis/onet/log"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/params"
)

//returns abi and bytecode of solidity contract
func getSmartContract(path string, nameOfContract string) (string, string) {
	abi, err := ioutil.ReadFile(path + nameOfContract + "_sol_" + nameOfContract + ".abi")
	if err != nil {
		log.LLvl1("Problem generating contract ABI")
	} else {
		log.LLvl1("ABI generated")
	}
	bin, err := ioutil.ReadFile(path + nameOfContract + "_sol_" + nameOfContract + ".bin")
	if err != nil {
		log.LLvl1("Problem generating contract BIN")
	} else {
		log.LLvl1("BIN generated")
	}
	return string(abi), string(bin)
}

func getChainConfig() *params.ChainConfig {
	///ChainConfig (adapted from Rinkeby test net)
	chainconfig := &params.ChainConfig{
		ChainID:             big.NewInt(1),
		HomesteadBlock:      big.NewInt(0),
		DAOForkBlock:        nil,
		DAOForkSupport:      false,
		EIP150Block:         nil,
		EIP150Hash:          common.HexToHash("0x0000000000000000000000000000000000000000"),
		EIP155Block:         big.NewInt(0),
		EIP158Block:         big.NewInt(0),
		ByzantiumBlock:      big.NewInt(0),
		ConstantinopleBlock: nil,
		Clique: &params.CliqueConfig{
			Period: 15,
			Epoch:  30000,
		},
	}
	return chainconfig
}

func getVMConfig() vm.Config {
	//vmConfig Config
	vmconfig := &vm.Config{
		// Debug enabled debugging Interpreter options
		Debug: false,
		// Tracer is the op code logger
		Tracer: nil,
		// NoRecursion disabled Interpreter call, callcode,
		// delegate call and create.
		NoRecursion: false,
		// Enable recording of SHA3/keccak preimages
		EnablePreimageRecording: true,
		// JumpTable contains the EVM instruction table. This
		// may be left uninitialised and will be set to the default
		// table.
		//JumpTable [256]operation
		//JumpTable: ,
		// Type of the EWASM interpreter
		EWASMInterpreter: "",
		// Type of the EVM interpreter
		EVMInterpreter: "",
	}
	return *vmconfig
}

func returnCanTransfer() func(vm.StateDB, common.Address, *big.Int) bool {
	canTransfer := func(vm.StateDB, common.Address, *big.Int) bool {
		return true
	}
	return canTransfer
}

func returnTransfer() func(vm.StateDB, common.Address, common.Address, *big.Int) {
	transfer := func(vm.StateDB, common.Address, common.Address, *big.Int) {
	}
	return transfer
}

func returnGetHash() func(uint64) common.Hash {
	gethash := func(uint64) common.Hash {
		log.LLvl1("tried to get hash")
		return common.HexToHash("0x0000000000000000000000000000000000000000")
	}
	return gethash

}

func getContext() vm.Context {
	placeHolder := common.HexToAddress("0x0000000000000000000000000000000000000000")
	return vm.Context{
		CanTransfer: returnCanTransfer(),
		Transfer: returnTransfer(),
		GetHash: returnGetHash(),
		Origin: placeHolder,
		GasPrice: big.NewInt(0),
		Coinbase: placeHolder,
		GasLimit: 10000000000,
		BlockNumber: big.NewInt(0),
		Time: big.NewInt(1),
		Difficulty: big.NewInt(1),
	}

}

func getDB(memDb *MemDatabase) (*state.StateDB, error) {
	db := state.NewDatabase(memDb)
	//Creates a new state DB
	sdb, err := state.New(common.Hash{0x0}, db)
	if err != nil {
		return nil, err
	}
	return sdb, err
}


func getDB1(memDb *MemDatabase) (*state.StateDB, error) {
	db := ethdb.NewMemDatabase()
	sdb, err := state.New(common.Hash{}, state.NewDatabase(db))
	if err != nil {
		return nil, err
	}
	return sdb, nil
}


func spawnEvm(memDB *MemDatabase) (*vm.EVM, error) {
	sdb, err := getDB1(memDB)
	if err != nil {
		return nil, err
	}
	bvm := vm.NewEVM(getContext(), sdb, getChainConfig(), getVMConfig())
	return bvm, nil
}
