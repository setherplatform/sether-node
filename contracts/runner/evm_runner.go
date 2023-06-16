package runner

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/setherplatform/sether-node/contracts"
	"math/big"
)

// EVMRunner provides a simplified API to run EVM calls
type EVMRunner interface {
	// Execute performs a potentially write operation over the runner's state
	// It can be seen as a message (input,value) from sender to recipient that returns `ret`
	Execute(recipient common.Address, input []byte, gas uint64, value *big.Int) (ret []byte, err error)

	// ExecuteFrom is like Execute, but lets you specify the sender to use for the EVM call.
	// It exists only for use in the Tobin tax calculation done as part of TobinTransfer, because that
	// originally used the transaction's sender instead of the zero address.
	ExecuteFrom(sender, recipient common.Address, input []byte, gas uint64, value *big.Int) (ret []byte, err error)

	// Query performs a read operation over the runner's state
	// It can be seen as a message (input,value) from sender to recipient that returns `ret`
	Query(recipient common.Address, input []byte, gas uint64) (ret []byte, err error)

	// StopGasMetering stop gas metering
	StopGasMetering()

	// StartGasMetering start gas metering
	StartGasMetering()
}

type SharedEVMRunner struct{ *vm.EVM }

func (runner *SharedEVMRunner) Execute(recipient common.Address, input []byte, gas uint64, value *big.Int) (ret []byte, err error) {
	ret, _, err = runner.Call(vm.AccountRef(contracts.ZeroAddress), recipient, input, gas, value)
	return ret, err
}

func (runner *SharedEVMRunner) ExecuteFrom(sender, recipient common.Address, input []byte, gas uint64, value *big.Int) (ret []byte, err error) {
	ret, _, err = runner.Call(vm.AccountRef(sender), recipient, input, gas, value)
	return ret, err
}

func (runner *SharedEVMRunner) Query(recipient common.Address, input []byte, gas uint64) (ret []byte, err error) {
	ret, _, err = runner.StaticCall(vm.AccountRef(contracts.ZeroAddress), recipient, input, gas)
	return ret, err
}
