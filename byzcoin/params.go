package byzcoin

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"os/exec"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/core/vm/runtime"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/params"
)

func compileContract() {
	fmt.Println("compiling contract")
	cmd := exec.Command("solcjs", "simple.sol", "--abi")
	if err := cmd.Start(); err != nil {
		fmt.Println("problem in compiling")
		log.Fatal(err)
	} else {
		fmt.Println("Compilation successful")
	}
}

//returns abi and bytecode
func getSC(path string, name_of_contract string) (string, string) {
	abi, err := ioutil.ReadFile(path + name_of_contract + "_sol_" + name_of_contract + ".abi")
	if err != nil {
		fmt.Println("Problem generating contract ABI")
	} else {
		fmt.Println("ABI generated")
	}
	bin, err := ioutil.ReadFile(path + name_of_contract + "_sol_" + name_of_contract + ".bin")
	if err != nil {
		fmt.Println("Problem generating contract BIN")
	} else {
		fmt.Println("BIN generated")
	}
	return string(abi), string(bin)
}

func getKeys() (string, string) {
	private_key := "d07fa6ac3deb2a186b2a6381c9012d595d5c3d4fefb4dbb2856d00485e9ed1af"
	public_key := "0xE420b7546D387039dDaD2741a688CbEBD2578363"
	return public_key, private_key
}
func getKeys1() (string, string) {
	private_key := "2d456877faf65f60ec24d5a55a9a4c4aa6580ea7313c6733cd3afe83888bef6a"
	public_key := "0xe745E7ceA88A02a1Fabd4aE591371eF50BFDc099"
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
	//pass byzcoin evm db instead
	db := state.NewDatabase(ethdb.NewMemDatabase())
	//func New(root common.Hash, db Database) (*StateDB, error)
	//Create a new state from a given trie.
	sdb, err := state.New(common.HexToHash("0x0000000000000000000000000000000000000000"), db)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
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
