package main

import (
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/core/vm/runtime"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/params"
)

func getKeys() (string, string) {
	private_key := "d07fa6ac3deb2a186b2a6381c9012d595d5c3d4fefb4dbb2856d00485e9ed1af"
	public_key := "0xE420b7546D387039dDaD2741a688CbEBD2578363"
	return public_key, private_key
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

func getDB() (*state.StateDB, error) {
	db := state.NewDatabase(ethdb.NewMemDatabase())
	//func New(root common.Hash, db Database) (*StateDB, error)
	//Create a new state from a given trie.
	sdb, err := state.New(common.HexToHash("0x0000000000000000000000000000000000000000"), db)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	fmt.Println("DB setup")
	return sdb, err
}

func getConfig() *runtime.Config {
	public_key, _ := getKeys()
	sdb, _ := getDB()
	config := &runtime.Config{
		ChainConfig: getChainConfig(),
		Difficulty:  big.NewInt(1),
		Origin:      common.HexToAddress(public_key),
		Coinbase:    common.HexToAddress(public_key),
		BlockNumber: big.NewInt(1),
		Time:        big.NewInt(1),
		GasLimit:    1,
		GasPrice:    big.NewInt(0),
		Value:       big.NewInt(1),
		Debug:       false,
		EVMConfig:   getVMConfig(),

		State: sdb,
		//GetHashFn: func(n uint64) common.Hash,
		//GetHashFn: core.GetHashFn,
	}
	return config
}

/*
	fmt.Println("===== Through runtime =====")
	fmt.Println("Creation of contract")
	create_ret, contract_addr, _, err := runtime.Create(common.Hex2Bytes(minimum_token), getConfig())
	if err != nil {
		fmt.Println("Contract deployment unsuccessful")
		fmt.Println(err)
	} else {
		fmt.Println("Successful contract deployment")
	}

	fmt.Println("Return of contract", create_ret)
	fmt.Println("Address of contract", contract_addr.Hex())

	fmt.Println("===== End runtime =====")
*/
