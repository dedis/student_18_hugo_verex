package byzcoin

import (
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
)

func returnCanTransfer() func(vm.StateDB, common.Address, *big.Int) bool {
	canTransfer := func(vm.StateDB, common.Address, *big.Int) bool {
		//log.Println("Verified transfer")
		return true
	}
	return canTransfer
}
func returnTransfer() func(vm.StateDB, common.Address, common.Address, *big.Int) {
	transfer := func(vm.StateDB, common.Address, common.Address, *big.Int) {
		//log.Println("tried to transfer")
	}
	return transfer
}

func returnGetHash() func(uint64) common.Hash {
	gethash := func(uint64) common.Hash {
		log.Println("tried to get hash")
		return common.HexToHash("0x0000000000000000000000000000000000000000")
	}
	return gethash

}

func spawnEvm() *vm.EVM {
	sdb, err := getDB()
	if err != nil {
		fmt.Println(err)
	}
	canTransfer := returnCanTransfer()
	transfer := returnTransfer()
	gethash := returnGetHash()
	pk, _ := getKeys()
	ctx := vm.Context{CanTransfer: canTransfer, Transfer: transfer, GetHash: gethash, Origin: common.HexToAddress(pk), GasPrice: big.NewInt(1), Coinbase: common.HexToAddress(pk), GasLimit: 10000000000, BlockNumber: big.NewInt(0), Time: big.NewInt(1), Difficulty: big.NewInt(1)}
	bvm := vm.NewEVM(ctx, sdb, getChainConfig(), getVMConfig())
	return bvm
}
