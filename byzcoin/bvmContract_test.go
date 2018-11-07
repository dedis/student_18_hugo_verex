package byzcoin

import (
	"fmt"
	"testing"
	"time"

	"github.com/dedis/protobuf"

	"github.com/stretchr/testify/require"

	"github.com/dedis/cothority"
	"github.com/dedis/cothority/byzcoin"
	"github.com/dedis/cothority/darc"
	"github.com/dedis/onet"
)

func TestEVMContract_Spawn(t *testing.T) {
	fmt.Println("Test of EVM creation")
	// Create a new ledger and prepare for proper closing
	bct := newBCTest(t)
	defer bct.Close()

	// Create a new empty instance
	args := byzcoin.Arguments{
		{
			Name:  "test",
			Value: []byte{13},
		},
		{
			Name:  "two",
			Value: []byte{2},
		},
	}
	// And send it to the ledger.
	instID := bct.createInstance(t, args)

	//Actual call to Spawn is made here
	// Wait for the proof to be available.
	pr, err := bct.cl.WaitProof(instID, bct.gMsg.BlockInterval, nil)
	require.Nil(t, err)
	fmt.Println(pr.KeyValue())
	// Make sure the proof is a matching proof and not a proof of absence.
	require.True(t, pr.InclusionProof.Match())

	// Get the raw values of the proof.
	values, err := pr.InclusionProof.RawValues()
	require.Nil(t, err)
	// And decode the buffer to a ContractStruct.
	cs := KeyValueData{}
	fmt.Println("The buffer contains : ", cs)
	err = protobuf.Decode(values[0], &cs)
	fmt.Println("The buffer now contains : ", cs)
	require.Nil(t, err)

	//do the actual testing here

}

func TestEVMContract_Invoke(t *testing.T) {
	fmt.Println("Test of EVM invocation")
	bct := newBCTest(t)
	defer bct.Close()
	args := byzcoin.Arguments{}
	instID := bct.createInstance(t, args)
	// Wait for the proof to be available.
	pr1, err := bct.cl.WaitProof(instID, bct.gMsg.BlockInterval, nil)
	require.Nil(t, err)
	fmt.Println("HERE")
	bct.updateInstance(t, instID, args)
	fmt.Println("HERE")
	_, values1, err := pr1.KeyValue()
	fmt.Println(values1)
	require.Nil(t, err)

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
		[]string{"spawn:keyValue", "spawn:darc", "spawn:bvm", "invoke:update", "invoke:deployContract", "invoke:createAccount"}, out.signer.Identity())
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
	_, err = bct.cl.AddTransaction(ctx)
	require.Nil(t, err)
	return ctx.Instructions[0].DeriveID("")
}

func (bct *bcTest) updateInstance(t *testing.T, instID byzcoin.InstanceID, args byzcoin.Arguments) {
	ctx := byzcoin.ClientTransaction{
		Instructions: []byzcoin.Instruction{{
			InstanceID: instID,
			Nonce:      byzcoin.Nonce{},
			Index:      0,
			Length:     1,
			Invoke: &byzcoin.Invoke{
				Command: "deployContract",
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
