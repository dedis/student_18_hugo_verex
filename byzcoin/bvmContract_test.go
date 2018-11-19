package byzcoin

import (
	"testing"
	"time"

	"github.com/dedis/onet/log"

	"github.com/ethereum/go-ethereum/common"

	"github.com/dedis/protobuf"

	"github.com/stretchr/testify/require"

	"github.com/dedis/cothority"
	"github.com/dedis/cothority/byzcoin"
	"github.com/dedis/cothority/darc"
	"github.com/dedis/onet"
)

func TestEVMContract_Spawn(t *testing.T) {
	log.LLvl1("test: instance creation")
	// Create a new ledger and prepare for proper closing
	bct := newBCTest(t)
	defer bct.Close()
	// Create a new empty instance
	args := byzcoin.Arguments{}
	// And send it to the ledger.
	instID := bct.createInstance(t, args)
	// Wait for the proof to be available.
	pr, err := bct.cl.WaitProof(instID, bct.gMsg.BlockInterval, nil)
	require.Nil(t, err)
	// Make sure the proof is a matching proof and not a proof of absence.
	require.True(t, pr.InclusionProof.Match())
	// Get the raw values of the proof.
	values, err := pr.InclusionProof.RawValues()
	require.Nil(t, err)
	// And decode the buffer to a ContractStruct.
	cs := KeyValueData{}
	err = protobuf.Decode(values[0], &cs)
	require.Nil(t, err)
}

func TestEVMContract_Invoke_Deploy(t *testing.T) {
	log.LLvl1("test: deploy a contract")
	bct := newBCTest(t)
	defer bct.Close()
	bytecode := common.Hex2Bytes("608060405234801561001057600080fd5b506106b7806100206000396000f30060806040526004361061006d576000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff1680630779afe61461007257806370a08231146100f7578063beabacc81461014e578063f01fe692146101d3578063f8b2cb4f14610220575b600080fd5b34801561007e57600080fd5b506100dd600480360381019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190803573ffffffffffffffffffffffffffffffffffffffff16906020019092919080359060200190929190505050610277565b604051808215151515815260200191505060405180910390f35b34801561010357600080fd5b50610138600480360381019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190505050610465565b6040518082815260200191505060405180910390f35b34801561015a57600080fd5b506101b9600480360381019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190803573ffffffffffffffffffffffffffffffffffffffff1690602001909291908035906020019092919050505061047d565b604051808215151515815260200191505060405180910390f35b3480156101df57600080fd5b5061021e60048036038101908080359060200190929190803573ffffffffffffffffffffffffffffffffffffffff1690602001909291905050506105fc565b005b34801561022c57600080fd5b50610261600480360381019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190505050610643565b6040518082815260200191505060405180910390f35b6000816000808673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020541015801561034557506000808473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002054826000808673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020540110155b1561045957816000808573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002054016000808573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002081905550816000808673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002054036000808673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020819055506001905061045e565b600090505b9392505050565b60006020528060005260406000206000915090505481565b6000816000808673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002054101515156104cc57600080fd5b6000808473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002054826000808673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002054011015151561055957600080fd5b816000808673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060008282540392505081905550816000808573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060008282540192505081905550600190509392505050565b816000808373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020819055505050565b60008060008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000205490509190505600a165627a7a72305820ddbfb05f7beb9052ec4080e56a86a2e3c87aa191ce50d9aefa285e723291889d0029")
	publicKey := []byte("0x2afd357E96a3aCbcd01615681C1D7e3398d5fb61")
	args := byzcoin.Arguments{
		{
			Name:  "publicKey",
			Value: publicKey,
		},
		{
			Name:  "bytecode",
			Value: bytecode,
		},
	}

	instID := bct.createInstance(t, args)
	// Wait for the proof to be available.
	_, err := bct.cl.WaitProof(instID, bct.gMsg.BlockInterval, nil)
	require.Nil(t, err)
	bct.deployContractInstance(t, instID, args)

}


func TestEVMContract_Invoke_Call(t *testing.T) {
	log.LLvl1("test: call a contract")
	bct := newBCTest(t)
	defer bct.Close()
	args := byzcoin.Arguments{}
	instID := bct.createInstance(t, args)
	// Wait for the proof to be available.
	pr1, err := bct.cl.WaitProof(instID, bct.gMsg.BlockInterval, nil)
	_, _, err = pr1.KeyValue()
	require.Nil(t, err)
	args = getArgsForCreate()
	bct.methodCallInstance(t, instID, args)
	require.Nil(t, err)
	/*
	args = getArgsForTransfer()
	bct.methodCallInstance(t, instID, args)
	require.Nil(t, err)*/

}
/*

func TestEVMContract_Invoke_General(t *testing.T) {
	log.LLvl1("test: call a contract")
	bct := newBCTest(t)
	defer bct.Close()
	args := byzcoin.Arguments{}
	instID := bct.createInstance(t, args)
	// Wait for the proof to be available.
	pr1, err := bct.cl.WaitProof(instID, bct.gMsg.BlockInterval, nil)
	_, _, err = pr1.KeyValue()
	require.Nil(t, err)
	args = getArgsGeneral()
	bct.methodCallInstance(t, instID, args)
	require.Nil(t, err)


}*/

func getArgsForCreate() byzcoin.Arguments {
	abi := []byte(`[{"constant":false,"inputs":[{"name":"_from","type":"address"},{"name":"_to","type":"address"},{"name":"_amount","type":"uint256"}],"name":"send","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[{"name":"","type":"address"}],"name":"balanceOf","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"_from","type":"address"},{"name":"_to","type":"address"},{"name":"_value","type":"uint256"}],"name":"transfer","outputs":[{"name":"success","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"name":"initialSupply","type":"uint256"},{"name":"toGiveTo","type":"address"}],"name":"create","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[{"name":"account","type":"address"}],"name":"getBalance","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"}]`)
	methodName := []byte("create")
	contractAddress := []byte("0xBd770416a3345F91E4B34576cb804a576fa48EB1")
	publicKey := []byte("0x1111111111111111111111111111111111111111")
	initialSupply := []byte("21000000")

	args := byzcoin.Arguments{
		{
			Name:  "contractAddress",
			Value: contractAddress,
		},
		{
			Name:  "abi",
			Value: abi,
		},
		{
			Name:  "method",
			Value: methodName,
		},
		{
			Name: "initialSupply",
			Value: initialSupply,
		},
		{
			Name: "from",
			Value: publicKey,
		},

	}
	return args
}

func getArgsGeneral() byzcoin.Arguments {
	abi := []byte(`[{"constant":false,"inputs":[{"name":"_from","type":"address"},{"name":"_to","type":"address"},{"name":"_amount","type":"uint256"}],"name":"send","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[{"name":"","type":"address"}],"name":"balanceOf","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"_from","type":"address"},{"name":"_to","type":"address"},{"name":"_value","type":"uint256"}],"name":"transfer","outputs":[{"name":"success","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"name":"initialSupply","type":"uint256"},{"name":"toGiveTo","type":"address"}],"name":"create","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[{"name":"account","type":"address"}],"name":"getBalance","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"}]`)
	methodName := []byte("create")
	contractAddress := []byte("0xBd770416a3345F91E4B34576cb804a576fa48EB1")
	publicKey := []byte("0x1111111111111111111111111111111111111111")
	initialSupply := []byte("21000000")

	args := byzcoin.Arguments{
		{
			Name:  "contractAddress",
			Value: contractAddress,
		},
		{
			Name:  "abi",
			Value: abi,
		},
		{
			Name:  "method",
			Value: methodName,
		},
		{
			Name: "initialSupply:uint32",
			Value: initialSupply,
		},
		{
			Name: "from:common.Address",
			Value: publicKey,
		},

	}
	return args
}

func getArgsForTransfer() byzcoin.Arguments{
	abi := []byte(`[{"constant":false,"inputs":[{"name":"_from","type":"address"},{"name":"_to","type":"address"},{"name":"_amount","type":"uint256"}],"name":"send","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[{"name":"","type":"address"}],"name":"balanceOf","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"_from","type":"address"},{"name":"_to","type":"address"},{"name":"_value","type":"uint256"}],"name":"transfer","outputs":[{"name":"success","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"name":"initialSupply","type":"uint256"},{"name":"toGiveTo","type":"address"}],"name":"create","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[{"name":"account","type":"address"}],"name":"getBalance","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"}]`)
	methodName := []byte("transfer")
	contractAddress := []byte("0xBd770416a3345F91E4B34576cb804a576fa48EB1")
	aPublicKey := []byte("0x1111111111111111111111111111111111111111")
	bPublicKey := []byte("0x2222222222222222222222222222222222222222")
	args := byzcoin.Arguments{
		{
			Name:  "abi",
			Value: abi,
		},
		{
			Name:  "method",
			Value: methodName,
		},
		{
			Name:  "contractAddress",
			Value: contractAddress,
		},
		{
			Name: "from",
			Value: aPublicKey,
		},
		{
			Name: "to",
			Value: bPublicKey,
		},

	}
	return args

}


// bcTest is used here to provide some simple test structure for different
// tests.
type bcTest struct {
	local   *onet.LocalTest
	signer  darc.Signer
	servers []*onet.Server
	roster  *onet.Roster
	cl      *byzcoin.Client
	gMsg    *byzcoin.CreateGenesisBlock
	gDarc   *darc.Darc
}

func newBCTest(t *testing.T) (out *bcTest) {
	out = &bcTest{}
	// First create a local test environment with three nodes.
	out.local = onet.NewTCPTest(cothority.Suite)

	out.signer = darc.NewSignerEd25519(nil, nil)
	out.servers, out.roster, _ = out.local.GenTree(3, true)

	// Then create a new ledger with the genesis darc having the right
	// to create and update keyValue contracts.
	var err error
	out.gMsg, err = byzcoin.DefaultGenesisMsg(byzcoin.CurrentVersion, out.roster,
		[]string{"spawn:keyValue", "spawn:darc", "spawn:bvm", "invoke:deploy", "invoke:call"}, out.signer.Identity())
	require.Nil(t, err)
	out.gDarc = &out.gMsg.GenesisDarc

	// This BlockInterval is good for testing, but in real world applications this
	// should be more like 5 seconds.
	out.gMsg.BlockInterval = time.Second / 2

	out.cl, _, err = byzcoin.NewLedger(out.gMsg, false)
	require.Nil(t, err)
	return out
}

func (bct *bcTest) Close() {
	bct.local.CloseAll()
}

func (bct *bcTest) createInstance(t *testing.T, args byzcoin.Arguments) byzcoin.InstanceID {
	ctx := byzcoin.ClientTransaction{
		Instructions: []byzcoin.Instruction{{
			InstanceID: byzcoin.NewInstanceID(bct.gDarc.GetBaseID()),
			Nonce:      byzcoin.Nonce{},
			Index:      0,
			Length:     1,
			Spawn: &byzcoin.Spawn{
				ContractID: ContractBvmID,
				Args:       args,
			},
		}},
	}
	// And we need to sign the instruction with the signer that has his
	// public key stored in the darc.
	require.Nil(t, ctx.Instructions[0].SignBy(bct.gDarc.GetBaseID(), bct.signer))

	// Sending this transaction to ByzCoin does not directly include it in the
	// global state - first we must wait for the new block to be created.
	var err error
	_, err = bct.cl.AddTransactionAndWait(ctx, 20)
	require.Nil(t, err)
	return ctx.Instructions[0].DeriveID("")
}


func (bct *bcTest) deployContractInstance(t *testing.T, instID byzcoin.InstanceID, args byzcoin.Arguments) {
	ctx := byzcoin.ClientTransaction{
		Instructions: []byzcoin.Instruction{{
			InstanceID: instID,
			Nonce:      byzcoin.Nonce{},
			Index:      0,
			Length:     1,
			Invoke: &byzcoin.Invoke{
				Command: "deploy",
				Args:    args,
			},
		}},
	}
	// And we need to sign the instruction with the signer that has his
	// public key stored in the darc.
	require.Nil(t, ctx.Instructions[0].SignBy(bct.gDarc.GetBaseID(), bct.signer))

	// Sending this transaction to ByzCoin does not directly include it in the
	// global state - first we must wait for the new block to be created.
	var err error
	_, err = bct.cl.AddTransactionAndWait(ctx, 10)
	require.Nil(t, err)
}

func (bct *bcTest) methodCallInstance(t *testing.T, instID byzcoin.InstanceID, args byzcoin.Arguments) {

	ctx := byzcoin.ClientTransaction{
		Instructions: []byzcoin.Instruction{{
			InstanceID: instID,
			Nonce:      byzcoin.Nonce{},
			Index:      0,
			Length:     1,
			Invoke: &byzcoin.Invoke{
				Command: "call",
				Args:    args,
			},
		}},
	}

	// And we need to sign the instruction with the signer that has his
	// public key stored in the darc.
	require.Nil(t, ctx.Instructions[0].SignBy(bct.gDarc.GetBaseID(), bct.signer))

	// Sending this transaction to ByzCoin does not directly include it in the
	// global state - first we must wait for the new block to be created.
	var err error
	_, err = bct.cl.AddTransactionAndWait(ctx, 20)
	require.Nil(t, err)
}
