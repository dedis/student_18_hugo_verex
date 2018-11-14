package byzcoin

import (
	"crypto/ecdsa"
	"io/ioutil"
	"math/big"

	"github.com/dedis/onet/log"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
	"github.com/pborman/uuid"
)

//Key creation from Ethereum library
type Key struct {
	Id uuid.UUID // Version 4 "random" for unique id not derived from key data
	// to simplify lookups we also store the address
	Address common.Address
	// we only store privkey as pubkey/address can be derived from it
	// privkey in this struct is always in plaintext
	PrivateKey *ecdsa.PrivateKey
}

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

func GenerateKeys() (address common.Address, privateKey *ecdsa.PrivateKey) {
	private, err := crypto.GenerateKey()
	if err != nil {
		log.LLvl1(err)
	}
	key := NewKeyFromECDSA(private)
	address = key.Address
	privateKey = key.PrivateKey
	return
}

//CreditAccount creates an account and load it with ether
func CreditAccount(db *state.StateDB, key common.Address, value int64) common.Address {
	db.SetBalance(key, big.NewInt(1000000000000000000*value))
	log.Lvl2("Loaded account", key.Hex(), "with ", value, " ether")
	return key
}

//NewKeyFromECDSA :
func NewKeyFromECDSA(privateKeyECDSA *ecdsa.PrivateKey) *Key {
	id := uuid.NewRandom()
	key := &Key{
		Id:         id,
		Address:    crypto.PubkeyToAddress(privateKeyECDSA.PublicKey),
		PrivateKey: privateKeyECDSA,
	}
	return key
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

func getDB(memDb *MemDatabase) (*state.StateDB, error) {
	db := state.NewDatabase(memDb)
	//Creates a new state DB
	sdb, err := state.New(common.HexToHash("0x0000000000000000000000000000000000000000"), db)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	return sdb, err
}
