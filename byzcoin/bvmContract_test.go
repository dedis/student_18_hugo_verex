package byzcoin

import (
	"fmt"
	"testing"
	"time"

	"github.com/dedis/cothority"
	"github.com/dedis/cothority/byzcoin"
	"github.com/dedis/darc"
	"github.com/dedis/onet"
)

func TestbvmContract_Spawn(t *testing.T) {
	fmt.Println("Test of spawn")
	bct := newBCTest(t)
	defer bct.Close()
  args := {

  }

  instID := bct.createBvm(t, args)


  //Is proof needed here?

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



  //do the actual testing here
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
		[]string{"spawn:keyValue", "spawn:darc", "invoke:update"}, out.signer.Identity())
	require.Nil(t, err)
	out.gDarc = &out.gMsg.GenesisDarc

	// This BlockInterval is good for testing, but in real world applications this
	// should be more like 5 seconds.
	out.gMsg.BlockInterval = time.Second / 2

	out.cl, _, err = byzcoin.NewLedger(out.gMsg, false)
	require.Nil(t, err)
	return out
}

func (bct *bcTest) createBvm(t *testing.T, args byzcoin.Arguments) byzcoin.InstanceID {
  ctx := byzcoin.ClientTransaction{
    Instructions: []byzcoin.Instruction{{
      InstanceID: byzcoin.NewInstanceID(bct.gDarc.GetBaseID()),
      Nonce:      byzcoin.Nonce{},
      Index:      0,
      Length:     1,
      Spawn: &byzcoin.Spawn{
        ContractID: ContractKeyValueID,
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
