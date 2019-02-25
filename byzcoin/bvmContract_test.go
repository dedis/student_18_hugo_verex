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

//Spawn a bvm
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

//Credits and displays an account balance
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

//Credits and displays three accounts balances
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
		// Get the proof from byzcoin
		reply, err := bct.cl.GetProof(instID.Slice())
		require.Nil(t, err)

		//Make sure the proof is a matching proof and not a proof of absence.
		pr := reply.Proof
		require.True(t, pr.InclusionProof.Match(instID.Slice()))

		_, err = bct.cl.WaitProof(instID, bct.gMsg.BlockInterval, nil)
		require.Nil(t, err)
		bct.displayAccountInstance(t, instID, args)
		bct.ct = bct.ct + 1
		// Get the proof from byzcoin
		reply, err = bct.cl.GetProof(instID.Slice())
		require.Nil(t, err)

		//Make sure the proof is a matching proof and not a proof of absence.
		pr = reply.Proof
		require.True(t, pr.InclusionProof.Match(instID.Slice()))

		_, err = bct.cl.WaitProof(instID, bct.gMsg.BlockInterval, nil)
		require.Nil(t, err)
	}
}

//Test the MinimumToken from the contract folders. Deploying, setting the constructor and then transferring between accounts a to b
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

	//Make sure the proof is a matching proof and not a proof of absence.
	pr := reply.Proof
	require.True(t, pr.InclusionProof.Match(instID.Slice()))

	_, err = bct.cl.WaitProof(instID, bct.gMsg.BlockInterval, nil)
	require.Nil(t, err)


	//CREDIT
	//Crediting account to have enough ether to deploy
	addressA := "0x2afd357E96a3aCbcd01615681C1D7e3398d5fb61"
	addressABuffer := []byte(addressA)
	args = byzcoin.Arguments{
		{
			Name: "address",
			Value: addressABuffer,
		},
	}

	log.LLvl1("credit")
	//Send credit instructions to Byzcoin and incrementing nonce counter
	bct.creditAccountInstance(t, instID, args)
	bct.ct = bct.ct +1
	// Get the proof from byzcoin
	reply, err = bct.cl.GetProof(instID.Slice())
	require.Nil(t, err)
	// Make sure the proof is a matching proof and not a proof of absence.
	pr = reply.Proof
	require.True(t, pr.InclusionProof.Match(instID.Slice()))

	log.LLvl1("display")
	//Verifying account credit
	bct.displayAccountInstance(t, instID, args)
	bct.ct = bct.ct +1
	// Get the proof from byzcoin
	reply, err = bct.cl.GetProof(instID.Slice())
	require.Nil(t, err)
	// Make sure the proof is a matching proof and not a proof of absence.
	pr = reply.Proof
	require.True(t, pr.InclusionProof.Match(instID.Slice()))

	//Getting smart-contract abi and bytecode
	RawAbi, bytecode := getSmartContract("MinimumToken")

	//Getting transaction parameters
	gasLimit, gasPrice := transactionGasParameters()

	s := []string{}
	s = append(s, bytecode)

	//TODO: create encodeConstructorArgument function
	//use the abiPack method with no function name provided
	//Constructor arguments encoded using abi specification :  (addressA, 100)
	encodedArgs := "0000000000000000000000002afd357e96a3acbcd01615681c1d7e3398d5fb610000000000000000000000000000000000000000000000000000000000000064"
	s = append(s, encodedArgs)
	data := strings.Join(s, "")


	//Ethereum transaction for deploying new contract
	deployTx := types.NewContractCreation(0,  big.NewInt(0), gasLimit, gasPrice, common.Hex2Bytes(data))

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
	log.LLvl1("deploy")
	//Sends deploy transaction to Byzcoin
	bct.transactionInstance(t, instID, args)
	//bct.ct = bct.ct +1
	// Get the proof from byzcoin
	reply, err = bct.cl.GetProof(instID.Slice())
	require.Nil(t, err)
	// Make sure the proof is a matching proof and not a proof of absence.
	pr = reply.Proof
	require.True(t, pr.InclusionProof.Match(instID.Slice()))


	args = byzcoin.Arguments{
		{
			Name: "address",
			Value: addressABuffer,
		},
	}
	time.Sleep(5)
	log.LLvl1("display")
	//Verifying account credit
	bct.displayAccountInstance(t, instID, args)
	bct.ct = bct.ct +1
	// Get the proof from byzcoin
	reply, err = bct.cl.GetProof(instID.Slice())
	require.Nil(t, err)
	// Make sure the proof is a matching proof and not a proof of absence.
	pr = reply.Proof
	require.True(t, pr.InclusionProof.Match(instID.Slice()))

	contractAddress := crypto.CreateAddress(common.HexToAddress(addressA), deployTx.Nonce())

	//TRANSACT A to B
	//Now send a token transaction from A to B
	addressB := "0x2887A24130cACFD8f71C479d9f9Da5b9C6425CE8"
	fromAddress := common.HexToAddress(addressA)
	toAddress := common.HexToAddress(addressB)
	methodBuf, err := abiMethodPack(RawAbi, "transferFrom", fromAddress, toAddress, big.NewInt(2))
	require.Nil(t, err)

	//create transaction
	sendTxAB := types.NewTransaction(1, contractAddress, big.NewInt(0), gasLimit, gasPrice, methodBuf)
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
	log.LLvl1("all transactions realised")
}

//Signs the transaction with a private key and returns the transaction in byte format, ready to be included into the Byzcoin transaction
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

//Creates the data to interact with an existing contract, with a variadic number of arguments
func abiMethodPack(contractABI string, methodCall string,  args ...interface{}) (data []byte, err error){
	ABI, err := abi.JSON(strings.NewReader(contractABI))
	if err != nil {
		return nil, err
	}
	abiCall, err := ABI.Pack(methodCall, args...)
	if err != nil {
		log.LLvl1("error in packing args", err)
		return nil, err
	}
	return abiCall, nil
}

//Return gas parameters for easy modification
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
	out.gMsg.BlockInterval = time.Second 

	out.cl, _, err = byzcoin.NewLedger(out.gMsg, false)
	require.Nil(t, err)
	out.ct = 1

	return out
}

func (bct *bcTest) Close() {
	bct.local.CloseAll()
}

//The following functions are Byzcoin transactions (instances) that will cary either the Ethereum transactions or
// a credit and display command

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
	_, err = bct.cl.AddTransactionAndWait(ctx, 30)
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
	_, err = bct.cl.AddTransactionAndWait(ctx, 30)
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
	_, err = bct.cl.AddTransactionAndWait(ctx, 30)
	require.Nil(t, err)
}

/*
func TestInvoke_LoanContract(t *testing.T){
	log.LLvl1("Deploying Loan Contract")

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


	//CREDIT
	//Preparing parameters to credit account to have enough ether to deploy
	addressA := "0x2afd357E96a3aCbcd01615681C1D7e3398d5fb61"
	addressABuffer := []byte(addressA)
	args = byzcoin.Arguments{
		{
			Name: "address",
			Value: addressABuffer,
		},
	}

	//Send credit instructions to Byzcoin and incrementing nonce counter
	bct.creditAccountInstance(t, instID, args)
	bct.ct = bct.ct +1

	//Verifying account credit
	bct.displayAccountInstance(t, instID, args)
	bct.ct = bct.ct +1

	//DEPLOY
	//Getting smartcontract abi and bytecode
	rawAbi, bytecode := getSmartContract("LoanContract")

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


	//CONSTRUCTOR
	//Once deployed we will now send the constructor arguments
	//constructor (uint256 _wantedAmount, uint256 _interest, uint256 _tokenAmount, string _tokenName, ERC20Token _tokenContractAddress, uint256 _length) public {
	tokenContractAddress := "0xdac17f958d2ee523a2206206994597c13d831ec7"
	constructorData, err := abiMethodPack(rawAbi, "constructor", big.NewInt(3), big.NewInt(1), big.NewInt(10000), "USDT", tokenContractAddress, big.NewInt(10))
	require.Nil(t, err)
	contractAddress := crypto.CreateAddress(common.HexToAddress(addressA), deployTx.Nonce())
	constructorTx := types.NewTransaction(1, contractAddress, big.NewInt(0), gasLimit, gasPrice, constructorData)
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

	//CHECK TOKENS

	checkTokenData, err := abiMethodPack(rawAbi, "checkTokens")
	require.Nil(t,err)
	checkTokenTx := types.NewTransaction(2, contractAddress, big.NewInt(0), gasLimit, gasPrice, checkTokenData)
	txBuffer, err = signAndMarshalTx(privateA, checkTokenTx)
	require.Nil(t, err)
	args = byzcoin.Arguments{
		{
			Name: "tx",
			Value: txBuffer,

		},
	}
	bct.transactionInstance(t,instID, args)
	bct.ct = bct.ct + 1

	//LEND

	addressB := "0x2887A24130cACFD8f71C479d9f9Da5b9C6425CE8"
	privateB := "a3e6a98125c8f88fdcb45f13ad65e762b8662865c214ff85e1b1f3efcdffbcc1"

	//CREDIT account B for lending

	addressBBuffer := []byte(addressB)
	args = byzcoin.Arguments{
		{
			Name: "address",
			Value: addressBBuffer,
		},
	}

	//Send credit instructions to Byzcoin and incrementing nonce counter
	bct.creditAccountInstance(t, instID, args)
	bct.ct = bct.ct +1

	lendData, err := abiMethodPack(rawAbi, "lend")
	require.Nil(t, err)
	lendTx := types.NewTransaction(0, contractAddress, big.NewInt(3*1e18),gasLimit, gasPrice, lendData)
	signedTxBuffer, err = signAndMarshalTx(privateB, lendTx)
	require.Nil(t, err)
	args = byzcoin.Arguments{
		{
			Name: "tx",
			Value: txBuffer,

		},
	}
	bct.transactionInstance(t,instID, args)
	bct.ct = bct.ct + 1
}

*/

