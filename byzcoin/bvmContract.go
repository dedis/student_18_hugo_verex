package byzcoin

import (
	"errors"

	"github.com/dedis/cothority/byzcoin"
	"github.com/dedis/darc"
	"github.com/dedis/protobuf"
)

/*
spawn: create a new EVM and initialize the database structure
invoke:createAccount
invoke:sendCommand
invoke:mintCoins - send that many coins directly to the account, out of nowhere
*/

var contractBvmID = "bvm"

func contractBvm(cdb byzcoin.CollectionView, inst byzcoin.Instruction, cIn []byzcoin.Coin) (scs []byzcoin.StateChange, cOut []byzcoin.Coin, err error) {
	cOut = cIn

	err = inst.VerifyDarcSignature(cdb)
	if err != nil {
		return
	}

	var value []byte
	var darcID darc.ID
	value, _, darcID, err = cdb.GetValues(inst.InstanceID.Slice())
	if err != nil {
		return
	}
	var coin_instance byzcoin.Coin
	if inst.Spawn == nil {
		// Only if its NOT a spawn instruction is there data in the instance
		if value != nil {
			err = protobuf.Decode(value, &coin_instance)
			if err != nil {
				return nil, nil, errors.New("couldn't unmarshal instance data: " + err.Error())
			}
		}
	}

	switch inst.GetType() {
	case byzcoin.SpawnType:
		bvm := spawnEvm()

	case byzcoin.InvokeType:

		switch inst.Invoke.Command {
		case "createAccount":
		case "sendCommand":
		case "mintCoins":
		}
	}

	return nil
}
