package main

import (
	"fmt"
	"log"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/core/vm/runtime"
)

func main() {

	simple_bin := "608060405234801561001057600080fd5b506102f4806100206000396000f30060806040526004361061004c576000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff1680633450bd6a1461005157806368b64d1f1461007c575b600080fd5b34801561005d57600080fd5b5061006661015e565b6040518082815260200191505060405180910390f35b34801561008857600080fd5b506100e3600480360381019080803590602001908201803590602001908080601f0160208091040260200160405190810160405280939291908181526020018383808284378201915050505050509192919290505050610168565b6040518080602001828103825283818151815260200191508051906020019080838360005b83811015610123578082015181840152602081019050610108565b50505050905090810190601f1680156101505780820380516001836020036101000a031916815260200191505b509250505060405180910390f35b6000611002905090565b60608160009080519060200190610180929190610223565b5060008054600181600116156101000203166002900480601f0160208091040260200160405190810160405280929190818152602001828054600181600116156101000203166002900480156102175780601f106101ec57610100808354040283529160200191610217565b820191906000526020600020905b8154815290600101906020018083116101fa57829003601f168201915b50505050509050919050565b828054600181600116156101000203166002900490600052602060002090601f016020900481019282601f1061026457805160ff1916838001178555610292565b82800160010185558215610292579182015b82811115610291578251825591602001919060010190610276565b5b50905061029f91906102a3565b5090565b6102c591905b808211156102c15760008160009055506001016102a9565b5090565b905600a165627a7a72305820ab52658a6b1cc887cf50fa6a1119c5695892aeacdbb0b8e6b4a19e85fea53bc50029"

	simple_abi := `[{"constant":true,"inputs":[],"name":"returnNumber","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"pure","type":"function"},{"constant":true,"inputs":[{"name":"_param","type":"string"}],"name":"returnAddress","outputs":[{"name":"","type":"string"}],"payable":false,"stateMutability":"view","type":"function"}]`

	public_key, _ := getKeys()

	//accountRef := &vm.AccountRef{}
	accountRef := vm.AccountRef(common.HexToAddress(public_key))

	//accountRef := vm.NewContract(vm.AccountRef(common.HexToAddress(public_key)), nil, new(big.Int), 100000)

	//contract := NewContract(AccountRef(common.HexToAddress("1337")), nil, new(big.Int), reqGas)

	//contract_pk := common.HexToAddress("0xBd770416a3345F91E4B34576cb804a576fa48EB1")

	abi, err := abi.JSON(strings.NewReader(simple_abi))
	if err != nil {
		fmt.Println(err)
	}
	returnAddress, err := abi.Pack("returnAddress")
	if err != nil {
		fmt.Println(err)
	}
	returnNumber, err := abi.Pack("returnNumber")
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("DB setup")

	sdb, _ := getDB()

	canTransfer := func(vm.StateDB, common.Address, *big.Int) bool {
		log.Println("Verified transfer")
		return true
	}

	transfer := func(vm.StateDB, common.Address, common.Address, *big.Int) {
		log.Println("tried to transfer")
	}

	gethash := func(uint64) common.Hash {
		log.Println("tried to get hash")
		return common.HexToHash("0x0000000000000000000000000000000000000000")
	}

	fmt.Println("Setting up balances")

	sdb.SetBalance(common.HexToAddress(public_key), big.NewInt(1000000000000))

	fmt.Println("Sending", sdb.GetBalance(common.HexToAddress(public_key)), "eth to", common.HexToAddress(public_key).Hex())

	fmt.Println("Setting up context")
	ctx := vm.Context{CanTransfer: canTransfer, Transfer: transfer, GetHash: gethash, Origin: common.HexToAddress(public_key), GasPrice: big.NewInt(1), Coinbase: common.HexToAddress(public_key), GasLimit: 10000000000, BlockNumber: big.NewInt(0), Time: big.NewInt(1), Difficulty: big.NewInt(1)}

	fmt.Println("Setting up & checking VMs")
	bvm := vm.NewEVM(ctx, sdb, getChainConfig(), getVMConfig())
	bvm1 := runtime.NewEnv(getConfig())
	bvmInterpreter := vm.NewEVMInterpreter(bvm, getVMConfig())
	bvm1Interpreter := vm.NewEVMInterpreter(bvm1, getVMConfig())
	a := bvmInterpreter.CanRun(common.Hex2Bytes(simple_bin))
	b := bvm1Interpreter.CanRun(common.Hex2Bytes(simple_bin))
	if !a || !b {
		fmt.Println("Problem setting up vms")
	}

	fmt.Println("### Contract creation")
	ret, addrContract, leftOverGas, err := bvm.Create(accountRef, common.Hex2Bytes(simple_bin), 100000000, big.NewInt(0))
	if err != nil {
		fmt.Println("Contract deployment unsuccessful")
		fmt.Println(err)
	} else {
		fmt.Println("Successful contract deployment")
	}
	fmt.Println("Return of contract", common.Bytes2Hex(ret))
	fmt.Println("Left over gas : ", leftOverGas)
	fmt.Println("Contract address", addrContract.Hex())
	fmt.Println("### Contract call")
	ret_call, leftOverGas, err := bvm.Call(accountRef, addrContract, returnAddress, leftOverGas, big.NewInt(0))
	if err != nil {
		fmt.Println("Contract call unsuccessful")
		fmt.Println(err)
	} else {
		fmt.Println("Successful contract call")
	}
	fmt.Println("Return of call", common.Bytes2Hex(ret_call))
	fmt.Println("Left over gas : ", leftOverGas)
	fmt.Println("Nonce contract", sdb.GetNonce(addrContract))
	ret_call1, leftOverGas1, err := bvm.Call(accountRef, addrContract, returnNumber, leftOverGas, big.NewInt(0))
	if err != nil {
		fmt.Println("Contract call unsuccessful")
		fmt.Println(err)
	} else {
		fmt.Println("Successful contract call")
	}
	fmt.Println("Return of call", common.Bytes2Hex(ret_call1))
	fmt.Println("Left over gas : ", leftOverGas1)
	fmt.Println("Nonce contract", sdb.GetNonce(addrContract))

	fmt.Println("===== End vm =====")

	fmt.Println("===== Through runtime =====")
	fmt.Println("Creation of contract")
	create_ret, contract_addr, _, err := runtime.Create(common.Hex2Bytes(simple_bin), getConfig())
	if err != nil {
		fmt.Println("Contract deployment unsuccessful")
		fmt.Println(err)
	} else {
		fmt.Println("Successful contract deployment")
		fmt.Println("Return of contract", create_ret)
		fmt.Println("Address of contract", contract_addr.Hex())
	}

	call_ret, _, err := runtime.Execute([]byte(simple_bin), returnNumber, getConfig())
	if err != nil {
		fmt.Println("Contract deployment unsuccessful")
		fmt.Println(err)
	} else {
		fmt.Println("Contract deployment successful ")
		fmt.Println("Contract returned : ", call_ret)
	}

	fmt.Println("===== End runtime =====")

	//func (in *EVMInterpreter) Run(contract *Contract, input []byte, readOnly bool) (ret []byte, err error)

	/*contract := &vm.Contract{
		CallerAddress: common.HexToAddress("0xE420b7546D387039dDaD2741a688CbEBD2578363"),
		Code:          common.Hex2Bytes(minimum_token),
		CodeHash:      common.HexToHash("0x0000000000000000000000000000000000000000"),
		//CodeAddr:      &(common.HexToAddress("0xE420b7546D387039dDaD2741a688CbEBD2578363")),
		Input: nil,
		Gas:   1000000,
	}*/

	//fmt.Println(ret)
	//fmt.Println(addrContract)

	//new_contract := vm.NewContract(accountRef, contract, big.NewInt(10000), 1)

}
