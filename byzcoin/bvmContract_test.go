package byzcoin

import (
	"github.com/ethereum/go-ethereum/accounts/abi"
	"math/big"
	"strings"
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
	bct.local.Check = onet.CheckNone
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

func TestEVMContract_Invoke_Credit(t *testing.T) {
	log.LLvl1("test: crediting an account")
	bct := newBCTest(t)
	bct.local.Check = onet.CheckNone
	defer bct.Close()
	address := "0x2afd357E96a3aCbcd01615681C1D7e3398d5fb61"
	addressB := []byte(address)
	value := []byte("1000")
	args := byzcoin.Arguments{
		{
			Name:  "address",
			Value: addressB,
		},
		{
			Name:  "value",
			Value: value,
		},
	}
	instID := bct.createInstance(t, args)
	// Wait for the proof to be available.
	_, err := bct.cl.WaitProof(instID, bct.gMsg.BlockInterval, nil)
	require.Nil(t, err)
	bct.creditAccountInstance(t, instID, args)
	require.Nil(t, err)

}


func TestEVMContract_Invoke_Display(t *testing.T){
	log.LLvl1("test: displaying account")
	bct := newBCTest(t)
	bct.local.Check = onet.CheckNone
	defer bct.Close()
	address := "0x2afd357E96a3aCbcd01615681C1D7e3398d5fb61"
	addressB := []byte(address)
	args := byzcoin.Arguments{
		{
			Name:  "address",
			Value: addressB,
		},
	}
	instID := bct.createInstance(t, args)
	_, err := bct.cl.WaitProof(instID, bct.gMsg.BlockInterval, nil)
	require.Nil(t, err)
	bct.displayAccountInstance(t, instID, args)
	require.Nil(t, err)


}


func TestEVMContract_Invoke_Call(t *testing.T) {
	log.LLvl1("test: call a contract")
	bct := newBCTest(t)
	//defer bct.Close()
	args := byzcoin.Arguments{}
	instID := bct.createInstance(t, args)
	// Wait for the proof to be available.
	pr1, err := bct.cl.WaitProof(instID, bct.gMsg.BlockInterval, nil)
	_, _, err = pr1.KeyValue()
	require.Nil(t, err)
	args, err = getAbiCallForCreate()
	require.Nil(t, err)
	bct.methodCallInstance(t, instID, args)
	require.Nil(t, err)
	/*
	args = getArgsForTransfer()
	bct.methodCallInstance(t, instID, args)
	require.Nil(t, err)*/

}


func TestEVMContract_Apply_Transaction(t *testing.T){
	log.LLvl1("Testing applying tx")
	sendTransactionHelper(&common.Address{0x01})

}

func getAbiCallForCreate() (ba byzcoin.Arguments , err error){
	abiBuf := `[{"constant":false,"inputs":[{"name":"_from","type":"address"},{"name":"_to","type":"address"},{"name":"_amount","type":"uint256"}],"name":"send","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[{"name":"","type":"address"}],"name":"balanceOf","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"_from","type":"address"},{"name":"_to","type":"address"},{"name":"_value","type":"uint256"}],"name":"transfer","outputs":[{"name":"success","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"name":"initialSupply","type":"uint256"},{"name":"toGiveTo","type":"address"}],"name":"create","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[{"name":"account","type":"address"}],"name":"getBalance","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"}]`
	methodName := "create"
	contractAddress := []byte("0xBd770416a3345F91E4B34576cb804a576fa48EB1")
	publicKey := common.HexToAddress("0x1111111111111111111111111111111111111111")
	initialSupply := int64(100)


	ABI, err := abi.JSON(strings.NewReader(string(abiBuf)))
	if err != nil {
		return nil, err

	}

	abiCall, err := ABI.Pack(methodName, big.NewInt(initialSupply), publicKey)
	if err != nil {
		return nil, err
	}

	ba = byzcoin.Arguments{
		{
			Name: "contractAddress",
			Value: contractAddress,

		},
		{
			Name: "abiCall",
			Value: abiCall,

		},
	}
	return
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
		[]string{"spawn:keyValue", "spawn:darc", "spawn:bvm", "invoke:credit", "invoke:call", "invoke:display"}, out.signer.Identity())
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


func (bct *bcTest) creditAccountInstance(t *testing.T, instID byzcoin.InstanceID, args byzcoin.Arguments) {
	ctx := byzcoin.ClientTransaction{
		Instructions: []byzcoin.Instruction{{
			InstanceID: instID,
			Nonce:      byzcoin.Nonce{},
			Index:      0,
			Length:     1,
			Invoke: &byzcoin.Invoke{
				Command: "credit",
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


func (bct *bcTest) displayAccountInstance(t *testing.T, instID byzcoin.InstanceID, args byzcoin.Arguments){
	ctx := byzcoin.ClientTransaction{
		Instructions: []byzcoin.Instruction{{
			InstanceID: instID,
			Nonce:      byzcoin.Nonce{},
			Index:      0,
			Length:     1,
			Invoke: &byzcoin.Invoke{
				Command: "display",
				Args:    args,
			},
		}},
	}
	require.Nil(t, ctx.Instructions[0].SignBy(bct.gDarc.GetBaseID(), bct.signer))
	var err error
	_, err = bct.cl.AddTransactionAndWait(ctx, 20)
	require.Nil(t,err)


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
