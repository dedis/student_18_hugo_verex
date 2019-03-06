package byzcoin

import (
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"go.dedis.ch/cothority/v3"
	"go.dedis.ch/onet/v3/log"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.dedis.ch/cothority/v3/byzcoin"
	"go.dedis.ch/cothority/v3/darc"
	"go.dedis.ch/onet/v3"
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
	args := byzcoin.Arguments{
		{
			Name:  "address",
			Value: addressB,
		},
	}
	instID := bct.createInstance(t, args)
	_, err := bct.cl.WaitProof(instID, bct.gMsg.BlockInterval, nil)
	require.Nil(t, err)

	bct.creditAccountInstance(t, instID, args)
	reply, err := bct.cl.GetProof(instID.Slice())
	require.Nil(t, err)
	pr := reply.Proof
	require.True(t, pr.InclusionProof.Match(instID.Slice()))

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
		args := byzcoin.Arguments{
			{
				Name:  "address",
				Value: addressBuf,
			},
		}
		bct.creditAccountInstance(t, instID, args)
		// Get the proof from byzcoin
		reply, err := bct.cl.GetProof(instID.Slice())
		require.Nil(t, err)

		//Make sure the proof is a matching proof and not a proof of absence.
		pr := reply.Proof
		require.True(t, pr.InclusionProof.Match(instID.Slice()))

		_, err = bct.cl.WaitProof(instID, bct.gMsg.BlockInterval, nil)
		require.Nil(t, err)
		bct.displayAccountInstance(t, instID, args)
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
	//Crediting two accounts to have enough ether for tests
	addressA := "0x2afd357E96a3aCbcd01615681C1D7e3398d5fb61"
	addressABuffer := []byte(addressA)
	argsA := byzcoin.Arguments{
		{
			Name: "address",
			Value: addressABuffer,
		},
	}
	//Send credit instructions to Byzcoin and incrementing nonce counter
	bct.creditAccountInstance(t, instID, argsA)
	// Get the proof from byzcoin
	reply, err = bct.cl.GetProof(instID.Slice())
	require.Nil(t, err)
	// Make sure the proof is a matching proof and not a proof of absence.
	pr = reply.Proof
	require.True(t, pr.InclusionProof.Match(instID.Slice()))



	//Crediting second account to have enough ether to deploy
	addressB := "0x2887A24130cACFD8f71C479d9f9Da5b9C6425CE8"
	addressBBuffer := []byte(addressB)
	argsB := byzcoin.Arguments{
		{
			Name: "address",
			Value: addressBBuffer,
		},
	}
	//Send credit instructions to Byzcoin and incrementing nonce counter
	bct.creditAccountInstance(t, instID, argsB)
	// Get the proof from byzcoin
	reply, err = bct.cl.GetProof(instID.Slice())
	require.Nil(t, err)
	// Make sure the proof is a matching proof and not a proof of absence.
	pr = reply.Proof
	require.True(t, pr.InclusionProof.Match(instID.Slice()))


	//DISPLAY
	//Verifying both accounts credit
	bct.displayAccountInstance(t, instID, argsA)
	// Get the proof from byzcoin
	reply, err = bct.cl.GetProof(instID.Slice())
	require.Nil(t, err)
	// Make sure the proof is a matching proof and not a proof of absence.
	pr = reply.Proof
	require.True(t, pr.InclusionProof.Match(instID.Slice()))

	//Verifying second account credit
	bct.displayAccountInstance(t, instID, argsB)
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

	//use the abiPack method with no function name provided to abi encode constructor arguments
	//Constructor arguments encoded using abi specification :  (addressA, 100)
	encodedConstructor, err := abiMethodPack(RawAbi,"", common.HexToAddress(addressA), big.NewInt(100))
	require.Nil(t, err)
	s := []string{}
	s = append(s, bytecode)
	encodedArgs := common.Bytes2Hex(encodedConstructor)
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
	//Sends deploy transaction to Byzcoin
	bct.transactionInstance(t, instID, args)
	// Get the proof from byzcoin
	reply, err = bct.cl.GetProof(instID.Slice())
	require.Nil(t, err)
	// Make sure the proof is a matching proof and not a proof of absence.
	pr = reply.Proof
	require.True(t, pr.InclusionProof.Match(instID.Slice()))

	log.LLvl1("send two tokens from address a to b")
	//TRANSACT A to B
	//Now send a token transaction from A to B
	contractAddress := crypto.CreateAddress(common.HexToAddress(addressA), deployTx.Nonce())

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
	// Get the proof from byzcoin
	reply, err = bct.cl.GetProof(instID.Slice())
	require.Nil(t, err)
	// Make sure the proof is a matching proof and not a proof of absence.
	pr = reply.Proof
	require.True(t, pr.InclusionProof.Match(instID.Slice()))

	log.LLvl1("send one token back from address b to a")
	//TRANSACT B to A
	//Now send back the token
	privateB := "a3e6a98125c8f88fdcb45f13ad65e762b8662865c214ff85e1b1f3efcdffbcc1"
	methodBuf, err = abiMethodPack(RawAbi, "transferFrom", toAddress, fromAddress, big.NewInt(1))
	require.Nil(t, err)
	//create transaction
	sendTxBA := types.NewTransaction(0, contractAddress, big.NewInt(0), gasLimit, gasPrice, methodBuf)
	txBufferBA, err := signAndMarshalTx(privateB, sendTxBA)
	require.Nil(t, err)
	args = byzcoin.Arguments{
		{
			Name:  "tx",
			Value: txBufferBA,

		},
	}
	bct.transactionInstance(t,instID, args)
	// Get the proof from byzcoin
	reply, err = bct.cl.GetProof(instID.Slice())
	require.Nil(t, err)
	// Make sure the proof is a matching proof and not a proof of absence.
	pr = reply.Proof
	require.True(t, pr.InclusionProof.Match(instID.Slice()))

	//DISPLAY
	//Verifying both accounts credit were deducted from the gas fees
	bct.displayAccountInstance(t, instID, argsA)
	// Get the proof from byzcoin
	reply, err = bct.cl.GetProof(instID.Slice())
	require.Nil(t, err)
	// Make sure the proof is a matching proof and not a proof of absence.
	pr = reply.Proof
	require.True(t, pr.InclusionProof.Match(instID.Slice()))

	//Verifying second account credit
	bct.displayAccountInstance(t, instID, argsB)
	// Get the proof from byzcoin
	reply, err = bct.cl.GetProof(instID.Slice())
	require.Nil(t, err)
	// Make sure the proof is a matching proof and not a proof of absence.
	pr = reply.Proof
	require.True(t, pr.InclusionProof.Match(instID.Slice()))
}

func Test_Keys(t *testing.T){
	address, privateKey := GenerateKeys()
	log.LLvl1(address, privateKey)

	test, err := crypto.HexToECDSA(privateKey)
	require.Nil(t, err)
	txOpts := bind.NewKeyedTransactor(test)
	log.LLvl1("is this, THE nonce?", txOpts.Nonce)

}


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
	//adressA := "0x2afd357E96a3aCbcd01615681C1D7e3398d5fb61"
	//privateA := "a33fca62081a2665454fe844a8afbe8e2e02fb66af558e695a79d058f9042f0d"
	addressA, privateA := GenerateKeys()
	nonceA := uint64(0)
	addressABuffer := []byte(addressA)
	args = byzcoin.Arguments{
		{
			Name: "address",
			Value: addressABuffer,
		},
	}

	//Send credit instructions to Byzcoin and incrementing nonce counter
	bct.creditAccountInstance(t, instID, args)
	// Get the proof from byzcoin
	reply, err = bct.cl.GetProof(instID.Slice())
	require.Nil(t, err)
	// Make sure the proof is a matching proof and not a proof of absence.
	pr = reply.Proof
	require.True(t, pr.InclusionProof.Match(instID.Slice()))

	_, err = bct.cl.WaitProof(instID, bct.gMsg.BlockInterval, nil)
	require.Nil(t, err)

	//Verifying account credit
	bct.displayAccountInstance(t, instID, args)
	// Get the proof from byzcoin
	reply, err = bct.cl.GetProof(instID.Slice())
	require.Nil(t, err)
	// Make sure the proof is a matching proof and not a proof of absence.
	pr = reply.Proof
	require.True(t, pr.InclusionProof.Match(instID.Slice()))
	_, err = bct.cl.WaitProof(instID, bct.gMsg.BlockInterval, nil)
	require.Nil(t, err)

	//DEPLOY
	//Getting smartcontract abi and bytecode
	rawAbi, bytecode := getSmartContract("ERC20Token")

	//Getting transaction parameters
	gasLimit, gasPrice := transactionGasParameters()

	deployTx := types.NewContractCreation(nonceA,  big.NewInt(0), gasLimit, gasPrice, common.Hex2Bytes(bytecode))
	signedTxBuffer, err := signAndMarshalTx(privateA, deployTx)
	require.Nil(t, err)
	nonceA++
	args = byzcoin.Arguments{
		{
			Name:  "tx",
			Value: signedTxBuffer,
		},
	}
	bct.transactionInstance(t, instID, args)
	// Get the proof from byzcoin
	reply, err = bct.cl.GetProof(instID.Slice())
	require.Nil(t, err)
	// Make sure the proof is a matching proof and not a proof of absence.
	pr = reply.Proof
	require.True(t, pr.InclusionProof.Match(instID.Slice()))

	_, err = bct.cl.WaitProof(instID, bct.gMsg.BlockInterval, nil)
	require.Nil(t, err)
	erc20Address := crypto.CreateAddress(common.HexToAddress(addressA), deployTx.Nonce()).Hex()
	log.LLvl1("ERC20 deployed @", erc20Address)



	//Now lets gets the LoanContract abi & bytecode
	lcABI, lcBIN := getSmartContract("LoanContract")
	//constructor (uint256 _wantedAmount, uint256 _interest, uint256 _tokenAmount, string _tokenName, ERC20Token _tokenContractAddress, uint256 _length) public {
	//CONSTRUCTOR
	constructorData, err := abiMethodPack(lcABI, "", big.NewInt(3), big.NewInt(1), big.NewInt(10000), "", common.HexToAddress(erc20Address), big.NewInt(10))
	require.Nil(t, err)
	s := []string{}
	s = append(s, lcBIN)
	encodedArgs := common.Bytes2Hex(constructorData)
	s = append(s, encodedArgs)
	data := strings.Join(s, "")

	//Creating deploying transaction
	deployTx = types.NewContractCreation(nonceA,  big.NewInt(0), gasLimit, gasPrice, common.Hex2Bytes(data))

	//Signing transaction with private key corresponding to addressA
	signedTxBuffer, err = signAndMarshalTx(privateA, deployTx)
	nonceA++
	require.Nil(t, err)
	args = byzcoin.Arguments{
		{
			Name:  "tx",
			Value: signedTxBuffer,
		},
	}
	bct.transactionInstance(t, instID, args)
	contractAddress := crypto.CreateAddress(common.HexToAddress(addressA), deployTx.Nonce())
	log.LLvl1("contract loan deployed")

	//CHECK TOKENS
	checkTokenData, err := abiMethodPack(lcABI, "checkTokens")
	require.Nil(t,err)
	checkTokenTx := types.NewTransaction(nonceA, contractAddress, big.NewInt(0), gasLimit, gasPrice, checkTokenData)
	txBuffer, err := signAndMarshalTx(privateA, checkTokenTx)
	require.Nil(t, err)
	args = byzcoin.Arguments{
		{
			Name: "tx",
			Value: txBuffer,

		},
	}
	log.LLvl1("check tokens good")
	bct.transactionInstance(t,instID, args)
	log.LLvl1("check tokens good")

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
	/*
	keyB, err := crypto.HexToECDSA(addressB)
	require.Nil(t, err)
	transactorB := bind.NewKeyedTransactor(keyB)
	log.LLvl1("is this the correct nonce?", transactorB.Nonce)
	*/




	lendData, err := abiMethodPack(rawAbi, "lend")
	require.Nil(t, err)
	lendTx := types.NewTransaction(1, contractAddress, big.NewInt(3*1e18),gasLimit, gasPrice, lendData)
	signedTxBuffer, err = signAndMarshalTx(privateB, lendTx)
	require.Nil(t, err)
	args = byzcoin.Arguments{
		{
			Name: "tx",
			Value: txBuffer,

		},
	}
	bct.transactionInstance(t,instID, args)
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
		[]string{"spawn:bvm", "invoke:bvm.display", "invoke:bvm.credit", "invoke:bvm.transaction"}, out.signer.Identity())
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
	require.NoError(t, ctx.FillSignersAndSignWith(bct.signer))

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
	bct.ct++
	ctx.Instructions[0].Invoke.ContractID = "bvm"
	// And we need to sign the instruction with the signer that has his
	// public key stored in the darc.
	require.NoError(t, ctx.FillSignersAndSignWith(bct.signer))
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
	bct.ct++
	ctx.Instructions[0].Invoke.ContractID = "bvm"
	// And we need to sign the instruction with the signer that has his
	// public key stored in the darc.
	require.NoError(t, ctx.FillSignersAndSignWith(bct.signer))

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
	ctx.Instructions[0].Invoke.ContractID = "bvm"
	// And we need to sign the instruction with the signer that has his
	// public key stored in the darc.
	require.NoError(t, ctx.FillSignersAndSignWith(bct.signer))

	// Sending this transaction to ByzCoin does not directly include it in the
	// global state - first we must wait for the new block to be created.
	var err error
	_, err = bct.cl.AddTransactionAndWait(ctx, 30)
	require.Nil(t, err)
}


