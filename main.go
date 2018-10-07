package main

import (
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/core/vm/runtime"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/params"
)

func main() {

	//To call newEVM function the following parameters are needed :
	//ctx Context, statedb StateDB, chainconfig *params.ChainConfig, vmconfig Config

	//func NewTable(db Database, prefix string) Database
	//NewTable returns a Database object that prefixes all keys with a given string.

	//func NewDatabase(db ethdb.Database) Database
	/*
		NewDatabase creates a backing store for state. The returned database is safe for concurrent use and retains cached trie nodes in memory.
		The pool is an optional intermediate trie-node memory pool between the low level storage layer and the high level trie abstraction.
	*/
	db := state.NewDatabase(ethdb.NewMemDatabase())

	//StateDB creation
	//func New(root common.Hash, db Database) (*StateDB, error)
	//Create a new state from a given trie.
	sdb, err := state.New(common.HexToHash("0x0000000000000000000000000000000000000000"), db)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	test := func(vm.StateDB, common.Address, *big.Int) bool {
		return true
	}
	transfer := func(vm.StateDB, common.Address, common.Address, *big.Int) {
		log.Println("tried to transfer")
	}
	gethash := func(uint64) common.Hash {
		return common.HexToHash("0x0000000000000000000000000000000000000000")
	}
	//Context creation
	ctx := vm.Context{CanTransfer: test, Transfer: transfer, GetHash: gethash, Origin: common.HexToAddress("0x0000000000000000000000000000000000000000"), GasPrice: big.NewInt(1), Coinbase: common.HexToAddress("0x0000000000000000000000000000000000000000"), GasLimit: 10000000000, BlockNumber: big.NewInt(0), Time: big.NewInt(1), Difficulty: big.NewInt(1)}

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

	//vmConfig Config
	vmconfig := &vm.Config{
		// Debug enabled debugging Interpreter options
		Debug: true,
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

	flags := &runtime.Config{
		ChainConfig: chainconfig,
		Difficulty:  big.NewInt(1),
		Origin:      common.HexToAddress("0xE420b7546D387039dDaD2741a688CbEBD2578363"),
		Coinbase:    common.HexToAddress("0xE420b7546D387039dDaD2741a688CbEBD2578363"),
		BlockNumber: big.NewInt(1),
		Time:        big.NewInt(1),
		GasLimit:    1,
		GasPrice:    big.NewInt(1),
		Value:       big.NewInt(1),
		Debug:       true,
		EVMConfig:   *vmconfig,

		State: sdb,
		//GetHashFn: func(n uint64) common.Hash,
		//GetHashFn: nil,

	}

	bvm := vm.NewEVM(ctx, sdb, chainconfig, *vmconfig)

	//func NewEnv(cfg *Config) *vm.EVM

	bvm1 := runtime.NewEnv(flags)

	//func Call(address common.Address, input []byte, cfg *Config) ([]byte, uint64, error)
	//runtime.Call()
	log.Println("Hello")
	log.Println(bvm)
	log.Println("Hello")
	log.Println(bvm1)
	log.Println("Hello")

}
