package byzcoin

import (
	"github.com/dedis/onet/log"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/dedis/cothority"
	"github.com/dedis/cothority/byzcoin"
	"github.com/dedis/cothority/darc"
	"github.com/dedis/onet"
	"github.com/stretchr/testify/require"
)

func TestEVMContract_Spawn(t *testing.T) {
	log.LLvl1("test: instance creation")
	// Create a new ledger and prepare for proper closing
	bct := newBCTest(t)

	bct.local.Check = onet.CheckNone
	defer bct.local.CloseAll()

	//defer bct.Close()

	// Create an empty argument
	args := byzcoin.Arguments{}

	// And send it to the ledger.
	instID := bct.createInstance(t, args)

	// Get the proof from byzcoin
	reply, err := bct.cl.GetProof(instID.Slice())
	require.Nil(t, err)
	// Make sure the proof is a matching proof and not a proof of absence.
	pr := reply.Proof
	require.True(t, pr.InclusionProof.Match(instID.Slice()))
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
}

func TestEVMContract_Invoke_Credit(t *testing.T) {
	log.LLvl1("test: crediting an account")
	bct := newBCTest(t)
	bct.local.Check = onet.CheckNone
	defer bct.Close()
	address := "0x2afd357E96a3aCbcd01615681C1D7e3398d5fb61"
	addressB := []byte(address)
	value := []byte("1")
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
	bct.displayAccountInstance(t, instID, args)

}

func TestEVMContract_Invoke_Transaction_Deploy(t *testing.T){
	log.LLvl1("test: contract deployment")
	bct := newBCTest(t)
	bct.local.Check = onet.CheckNone
	defer bct.Close()
	privateA := "a33fca62081a2665454fe844a8afbe8e2e02fb66af558e695a79d058f9042f0d"
	_ , bytecode := getSmartContract("ModifiedToken")
	gasLimit := uint64(1e18)
	value := big.NewInt(0)
	gasPrice := big.NewInt(0)
	deployTx := types.NewContractCreation(0, value, gasLimit,  gasPrice, []byte(bytecode))
	signedTxBuffer, err := signAndMarshalTx(privateA, deployTx)
	require.Nil(t, err)
	args := byzcoin.Arguments{
		{
			Name: "tx",
			Value : signedTxBuffer,
		},
	}
	instID := bct.createInstance(t, args)
	_, err = bct.cl.WaitProof(instID, bct.gMsg.BlockInterval, nil)
	require.Nil(t, err)
	bct.transactionInstance(t, instID, args)
}

func TestEVMContract_Invoke_Transaction_Mint(t *testing.T){
	log.LLvl1("test: sending mint transaction")
	bct := newBCTest(t)
	bct.local.Check = onet.CheckNone
	defer bct.Close()
	privateA := "a33fca62081a2665454fe844a8afbe8e2e02fb66af558e695a79d058f9042f0d"
	gasLimit := uint64(1e18)
	value := big.NewInt(0)
	gasPrice := big.NewInt(0)
	contractAddress := []byte("0x45663483f58d687c8aF17B85cCCDD9391b567498")
	addressA := common.HexToAddress("0x2afd357E96a3aCbcd01615681C1D7e3398d5fb61")
	totalSupply := big.NewInt(21000000)
	RawAbi, _ := getSmartContract("ModifiedToken")
	methodBuf, _ := abiMethodPack(RawAbi,"create", totalSupply, addressA)
	generalTx := types.NewTransaction(0, common.HexToAddress(string(contractAddress)), value ,gasLimit, gasPrice, methodBuf)
	signedTxBuffer, err := signAndMarshalTx(privateA, generalTx)
	require.Nil(t, err)
	args := byzcoin.Arguments{
		{
			Name:  "tx",
			Value: signedTxBuffer,
		},
	}
	instID := bct.createInstance(t, args)
	_, err = bct.cl.WaitProof(instID, bct.gMsg.BlockInterval, nil)
	require.Nil(t, err)
	bct.transactionInstance(t, instID, args)
}

func signAndMarshalTx(privateKey string, tx *types.Transaction) ([]byte, error ){
	private, err := crypto.HexToECDSA(privateKey)
	if err !=nil {
		return nil, err
	}
	var signer types.Signer = types.HomesteadSigner{}
	signedTx, err := types.SignTx(tx, signer, private)
	if err !=nil {
		return nil, err
	}
	signedBuffer, err := signedTx.MarshalJSON()
	if err !=nil {
		return nil, err
	}
	return signedBuffer, err
}

func abiMethodPack(contractABI string, methodCall string,  args ...interface{}) (data []byte, err error){
	abiBuf := []byte(contractABI)
	ABI, err := abi.JSON(strings.NewReader(string(abiBuf)))
	if err != nil {
		return nil, err
	}
	abiCall, err := ABI.Pack(methodCall, args)
	return abiCall, nil
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
	ct      uint64
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
		[]string{"spawn:bvm", "invoke:transaction", "invoke:display", "invoke:credit"}, out.signer.Identity())
	require.Nil(t, err)
	out.gDarc = &out.gMsg.GenesisDarc

	// This BlockInterval is good for testing, but in real world applications this
	// should be more like 5 seconds.
	out.gMsg.BlockInterval = time.Second / 2

	out.cl, _, err = byzcoin.NewLedger(out.gMsg, false)
	require.Nil(t, err)
	out.ct = 1

	return out
}

func (bct *bcTest) Close() {
	bct.local.CloseAll()
}

//The following functions are Byzcoin transactions (instances) that will cary either the Ethereum transactions or
// a credit or display command

func (bct *bcTest) createInstance(t *testing.T, args byzcoin.Arguments) byzcoin.InstanceID {
	ctx := byzcoin.ClientTransaction{
		Instructions: []byzcoin.Instruction{{
			InstanceID:    byzcoin.NewInstanceID(bct.gDarc.GetBaseID()),
			SignerCounter: []uint64{bct.ct},
			Spawn: &byzcoin.Spawn{
				ContractID: ContractBvmID,
				Args:       args,
			},
		}},
	}
	bct.ct++
	// And we need to sign the instruction with the signer that has his
	// public key stored in the darc.
	require.NoError(t, ctx.SignWith(bct.signer))

	// Sending this transaction to ByzCoin does not directly include it in the
	// global state - first we must wait for the new block to be created.
	var err error
	_, err = bct.cl.AddTransactionAndWait(ctx, 20)
	require.Nil(t, err)
	return ctx.Instructions[0].DeriveID("")
}

func (bct *bcTest) displayAccountInstance(t *testing.T, instID byzcoin.InstanceID, args byzcoin.Arguments){
	ctx := byzcoin.ClientTransaction{
		Instructions: []byzcoin.Instruction{{
			InstanceID: instID,
			SignerCounter: []uint64{bct.ct},
			Invoke: &byzcoin.Invoke{
				Command: "display",
				Args:    args,
			},
		}},
	}
	// And we need to sign the instruction with the signer that has his
	// public key stored in the darc.
	require.NoError(t, ctx.SignWith(bct.signer))

	// Sending this transaction to ByzCoin does not directly include it in the
	// global state - first we must wait for the new block to be created.
	var err error
	_, err = bct.cl.AddTransactionAndWait(ctx, 20)
	require.Nil(t,err)
}



func (bct *bcTest) creditAccountInstance(t *testing.T, instID byzcoin.InstanceID, args byzcoin.Arguments){
	ctx := byzcoin.ClientTransaction{
		Instructions: []byzcoin.Instruction{{
			InstanceID: instID,
			SignerCounter: []uint64{bct.ct},
			Invoke: &byzcoin.Invoke{
				Command: "credit",
				Args:    args,
			},
		}},
	}
	// And we need to sign the instruction with the signer that has his
	// public key stored in the darc.
	require.NoError(t, ctx.SignWith(bct.signer))

	// Sending this transaction to ByzCoin does not directly include it in the
	// global state - first we must wait for the new block to be created.
	var err error
	_, err = bct.cl.AddTransactionAndWait(ctx, 20)
	require.Nil(t,err)
}


func (bct *bcTest) transactionInstance(t *testing.T, instID byzcoin.InstanceID, args byzcoin.Arguments) {
	ctx := byzcoin.ClientTransaction{
		Instructions: []byzcoin.Instruction{{
			InstanceID:    instID,
			SignerCounter: []uint64{bct.ct},
			Invoke: &byzcoin.Invoke{
				Command: "transaction",
				Args:    args,
			},
		}},
	}
	bct.ct++
	// And we need to sign the instruction with the signer that has his
	// public key stored in the darc.
	require.NoError(t, ctx.SignWith(bct.signer))

	// Sending this transaction to ByzCoin does not directly include it in the
	// global state - first we must wait for the new block to be created.
	var err error
	_, err = bct.cl.AddTransactionAndWait(ctx, 20)
	require.Nil(t, err)
}

