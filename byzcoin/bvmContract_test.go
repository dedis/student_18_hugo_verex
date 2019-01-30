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

func Test_Spawn(t *testing.T) {
	log.LLvl1("test: instantiating evm")
	// Create a new ledger and prepare for proper closing
	bct := newBCTest(t)
	bct.local.Check = onet.CheckNone
	defer bct.local.CloseAll()


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

func TestInvoke_Credit(t *testing.T) {
	log.LLvl1("test: crediting and displaying an account balance")
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
	_, err := bct.cl.WaitProof(instID, bct.gMsg.BlockInterval, nil)
	require.Nil(t, err)
	bct.creditAccountInstance(t, instID, args)
	bct.ct = bct.ct + 1
	bct.displayAccountInstance(t, instID, args)
}

func TestInvoke_Credit_Accounts(t *testing.T){
	log.LLvl1("test: crediting and checking accounts balances")
	// Create a new ledger and prepare for proper closing
	bct := newBCTest(t)
	bct.local.Check = onet.CheckNone
	defer bct.local.CloseAll()


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

	addresses := [3]string{"0x627306090abab3a6e1400e9345bc60c78a8bef57", "0xf17f52151ebef6c7334fad080c5704d77216b732", "0xc5fdf4076b8f3a5357c5e395ab970b5b54098fef"}
	for i:=0;i<3;i++{
		addressBuf := []byte(addresses[i])
		value := []byte("1")
		args := byzcoin.Arguments{
			{
				Name:  "address",
				Value: addressBuf,
			},
			{
				Name:  "value",
				Value: value,
			},
		}
		bct.creditAccountInstance(t, instID, args)
		bct.ct = bct.ct + 1
		bct.displayAccountInstance(t, instID, args)
		bct.ct = bct.ct + 1
	}
}


func TestInvoke_DeployToken(t *testing.T) {
	log.LLvl1("Deploying Token Contract")
	//Preparing ledger
	bct := newBCTest(t)
	bct.local.Check = onet.CheckNone
	defer bct.Close()

	//Instantiating evm
	args := byzcoin.Arguments{}
	instID := bct.createInstance(t, args)

	// Get the proof from byzcoin
	reply, err := bct.cl.GetProof(instID.Slice())
	require.Nil(t, err)
	// Make sure the proof is a matching proof and not a proof of absence.
	pr := reply.Proof
	require.True(t, pr.InclusionProof.Match(instID.Slice()))
	_, err = bct.cl.WaitProof(instID, bct.gMsg.BlockInterval, nil)
	require.Nil(t, err)

	//Preparing parameters to credit account to have enough ether to deploy
	addressA := "0x2afd357E96a3aCbcd01615681C1D7e3398d5fb61"
	addressABuffer := []byte(addressA)
	args = byzcoin.Arguments{
		{
			Name: "address",
			Value: addressABuffer,
		},
	}

	//Send credit instructions and incrementing counter
	bct.creditAccountInstance(t, instID, args)
	bct.ct = bct.ct +1

	//Verifying account credit
	bct.displayAccountInstance(t, instID, args)
	bct.ct = bct.ct +1


	//Getting smartcontract
	RawAbi, bytecode := getSmartContract("MinimumToken")

	//Getting transaction parameters
	gasLimit, gasPrice := transactionGasParameters()

	//Creating deploying transaction
	deployTx := types.NewContractCreation(0,  big.NewInt(0), gasLimit, gasPrice, common.Hex2Bytes(bytecode))


	//Signing transaction with private key corresponding to addressA
	privateA := "a33fca62081a2665454fe844a8afbe8e2e02fb66af558e695a79d058f9042f0d"
	signedTxBuffer, err := signAndMarshalTx(privateA, deployTx)
	require.Nil(t, err)
	args = byzcoin.Arguments{
		{
			Name:  "tx",
			Value: signedTxBuffer,
		},
	}


	bct.transactionInstance(t, instID, args)
	bct.ct = bct.ct +1


	//Calling constructor method to mint 100 coins
	methodBuf, err := abiMethodPack(RawAbi, "constructor", addressA, big.NewInt(100))
	require.Nil(t, err)


	// Old hardcoded contract address : contractAddress := common.HexToAddress("0x45663483f58d687c8aF17B85cCCDD9391b567498")

	//The contract address is derived from the sending address and its nonce
	contractAddress := crypto.CreateAddress(common.HexToAddress(addressA), deployTx.Nonce())


	constructorTx := types.NewTransaction(0, contractAddress, big.NewInt(0), gasLimit, gasPrice, methodBuf)
	txBuffer, err := signAndMarshalTx(privateA, constructorTx)
	require.Nil(t, err)
	args = byzcoin.Arguments{
		{
			Name: "tx",
			Value: txBuffer,

		},
	}
	bct.transactionInstance(t,instID, args)
	//bct.ct = bct.ct + 1

}


func TestContractBvm_Invoke_MintToken(t *testing.T) {
	log.LLvl1("Minting tokens")

	//PREPARE
	//Prepare the ledger
	bct := newBCTest(t)
	defer bct.Close()
	args := byzcoin.Arguments{}
	instID := bct.createInstance(t, args)
	// Get the proof from byzcoin
	reply, err := bct.cl.GetProof(instID.Slice())
	require.Nil(t, err)
	// Make sure the proof is a matching proof and not a proof of absence.
	pr := reply.Proof
	require.True(t, pr.InclusionProof.Match(instID.Slice()))

	//CREDIT
	//Preparing parameters to credit account to have enough ether to deploy
	addressA := "0x2afd357E96a3aCbcd01615681C1D7e3398d5fb61"
	addressABuffer :=[]byte(addressA)
	args = byzcoin.Arguments{
		{
			Name: "address",
			Value: addressABuffer,
		},
	}

	//Send credit instructions and incrementing counter
	bct.creditAccountInstance(t, instID, args)
	bct.ct = bct.ct +1

	//Verifying account credit
	bct.displayAccountInstance(t, instID, args)
	bct.ct = bct.ct +1

	//CONSTRUCTOR
	//Preparing the parameters to send the constructor transaction
	RawAbi, _ := getSmartContract("MinimumToken")
	fromAddress := common.HexToAddress(addressA)

	//Select the function using abi and give parameters. Will mint 100 coins
	methodBuf, err := abiMethodPack(RawAbi, "constructor", fromAddress, big.NewInt(100))
	require.Nil(t, err)

	//Transaction parameters
	contractAddress := common.HexToAddress("0x45663483f58d687c8aF17B85cCCDD9391b567498")
	amount := big.NewInt(0)
	gasLimit, gasPrice := transactionGasParameters()

	//Create transaction
	constructorTx := types.NewTransaction(0, contractAddress, amount, gasLimit, gasPrice, methodBuf)

	privateA := "a33fca62081a2665454fe844a8afbe8e2e02fb66af558e695a79d058f9042f0d"

	//sign with private key corresponding to address credited
	txBuffer, err := signAndMarshalTx(privateA, constructorTx)
	require.Nil(t, err)
	args = byzcoin.Arguments{
		{
			Name: "tx",
			Value: txBuffer,

		},
	}
	bct.transactionInstance(t,instID, args)
	bct.ct = bct.ct + 1



	//TRANSACT
	//Now send a token transaction from A to B
	addressB := "0x2887A24130cACFD8f71C479d9f9Da5b9C6425CE8"
	toAddress := common.HexToAddress(addressB)


	methodBuf, err = abiMethodPack(RawAbi, "transferFrom",fromAddress, toAddress, big.NewInt(100))
	require.Nil(t, err)

	//create transaction
	sendTxAB := types.NewTransaction(0, contractAddress, amount, gasLimit, gasPrice, methodBuf)
	txBufferAB, err := signAndMarshalTx(privateA, sendTxAB)
	require.Nil(t, err)
	args = byzcoin.Arguments{
		{
			Name:  "tx",
			Value: txBufferAB,

		},
	}
	bct.transactionInstance(t,instID, args)
	bct.ct = bct.ct +1


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

func transactionGasParameters()(gasLimit uint64, gasPrice *big.Int){
	gasLimit = uint64(1e7)
	gasPrice = big.NewInt(1)
	return
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



/*
func TestContractBvm_Invoke_SendTokenAB(t *testing.T) {
	log.LLvl1("Sending token from A to B")
	//prepare the ledger
	bct := newBCTest(t)
	defer bct.Close()
	args := byzcoin.Arguments{}

	instID := bct.createInstance(t, args)

	//get abi to select sc function
	RawAbi, _ := getSmartContract("MinimumToken")

	//select the function and give parameters
	privateA := "a33fca62081a2665454fe844a8afbe8e2e02fb66af558e695a79d058f9042f0d"
	privateB := "a3e6a98125c8f88fdcb45f13ad65e762b8662865c214ff85e1b1f3efcdffbcc1"
	fromAddress :=common.HexToAddress("0x2afd357E96a3aCbcd01615681C1D7e3398d5fb61")
	toAddress := common.HexToAddress("0x2887A24130cACFD8f71C479d9f9Da5b9C6425CE8")
	methodBuf, err := abiMethodPack(RawAbi, "transferFrom",fromAddress, toAddress, big.NewInt(100))
	require.Nil(t, err)

	//transaction parameters
	//placeholder waiting for deploying transaction to work
	contractAddress := common.HexToAddress("0x0000000000000000000000000000000000000000")
	amount := big.NewInt(0)
	gasLimit, gasPrice := transactionGasParameters()

	//create transaction
	sendTxAB := types.NewTransaction(0, contractAddress, amount, gasLimit, gasPrice, methodBuf)
	txBufferAB, err := signAndMarshalTx(privateA, sendTxAB)
	require.Nil(t, err)
	args = byzcoin.Arguments{
		{
			Name:  "tx",
			Value: txBufferAB,

		},
	}
	bct.transactionInstance(t,instID, args)
	bct.ct = bct.ct +1
	methodBuf, err = abiMethodPack(RawAbi, "transferFrom", toAddress, fromAddress, big.NewInt(100))
	require.Nil(t, err)
	sendTxBA := types.NewTransaction(0, contractAddress, amount, gasLimit, gasPrice, methodBuf)
	txBufferBA, err := signAndMarshalTx(privateB, sendTxBA)
	require.Nil(t, err)
	args = byzcoin.Arguments{
		{
			Name:  "tx",
			Value: txBufferBA,

		},
	}
	bct.transactionInstance(t, instID, args)



}*/


