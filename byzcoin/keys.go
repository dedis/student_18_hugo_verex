package byzcoin

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"github.com/dedis/onet/log"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pborman/uuid"
	"math/big"

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

func GenerateKeys() (address common.Address, privateKey *ecdsa.PrivateKey) {
	private, err := crypto.GenerateKey()
	if err != nil {
		log.Fatal(err)

	}
	//key := NewKeyFromECDSA(private)
	address = crypto.PubkeyToAddress(private.PublicKey)
	//address = key.Address
	//privateKey = key.PrivateKey

	log.Lvlf2("Public key : %x ",  elliptic.Marshal(crypto.S256(), private.PublicKey.X, private.PublicKey.Y))
	log.Lvl2("Address generated : ", address.Hex())
	privateKey = private
	return
}

//CreditAccount creates an account and load it with ether
func CreditAccount(db *state.StateDB, key common.Address, value int64) common.Address {
	db.SetBalance(key, big.NewInt(1e9*value))
	log.LLvl1("the balance of account ",key.Hex() ," is ", db.GetBalance(key))
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