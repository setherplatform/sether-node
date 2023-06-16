package runner

import (
	"bytes"
	"errors"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/log"
	"math/big"
)

var (
	errorSig     = []byte{0x08, 0xc3, 0x79, 0xa0} // Keccak256("Error(string)")[:4]
	abiString, _ = abi.NewType("string", "", nil)
)

// Method represents a contract's method
type Method struct {
	abi    *abi.ABI
	method string
	maxGas uint64
}

// NewMethod creates a new Method
func NewMethod(abi *abi.ABI, method string, maxGas uint64) Method {
	return Method{
		abi:    abi,
		method: method,
		maxGas: maxGas,
	}
}

// Bind returns a BoundMethod instance which can be used to call the contract method represented by am
// and residing at contracAddress.
func (am Method) Bind(contractAddress common.Address) *BoundMethod {
	return &BoundMethod{
		Method:  am,
		address: contractAddress,
	}
}

// encodeCall will encodes the msg into []byte format for EVM consumption
func (am Method) encodeCall(args ...interface{}) ([]byte, error) {
	return am.abi.Pack(am.method, args...)
}

// decodeResult will decode the result of msg execution into the result parameter
func (am Method) decodeResult(result interface{}, output []byte) error {
	if result == nil {
		return nil
	}
	return am.abi.UnpackIntoInterface(result, am.method, output)
}

// BoundMethod represents a Method that is bounded to an address
// In particular, instead of address we use an address resolver to cope the fact
// that addresses need to be obtained from the Registry before making a call
type BoundMethod struct {
	Method
	address common.Address
}

// NewBoundMethod constructs a new bound method instance bound to the given address.
func NewBoundMethod(contractAddress common.Address, abi *abi.ABI, methodName string, maxGas uint64) *BoundMethod {
	return NewMethod(abi, methodName, maxGas).Bind(contractAddress)
}

// Query executes the method with the given EVMRunner as a read only action, the returned
// value is unpacked into result.
func (bm *BoundMethod) Query(vmRunner EVMRunner, result interface{}, args ...interface{}) error {
	return bm.run(vmRunner, result, true, nil, args...)
}

// Execute executes the method with the given EVMRunner and unpacks the return value into result.
// If the method does not return a value then result should be nil.
func (bm *BoundMethod) Execute(vmRunner EVMRunner, result interface{}, value *big.Int, args ...interface{}) error {
	return bm.run(vmRunner, result, false, value, args...)
}

func (bm *BoundMethod) run(vmRunner EVMRunner, result interface{}, readOnly bool, value *big.Int, args ...interface{}) error {
	contractAddress := bm.address

	logger := log.New("to", contractAddress, "method", bm.method)

	input, err := bm.encodeCall(args...)
	if err != nil {
		logger.Error("Error invoking evm function: can't encode method arguments", "args", args, "err", err)
		return err
	}

	var output []byte
	if readOnly {
		output, err = vmRunner.Query(contractAddress, input, bm.maxGas)
	} else {
		output, err = vmRunner.Execute(contractAddress, input, bm.maxGas, value)
	}

	if err != nil {
		message, _ := unpackError(output)
		logger.Error("Error invoking evm function: EVM call failure", "input", hexutil.Encode(input), "maxgas", bm.maxGas, "err", err, "message", message)
		return err
	}

	if err := bm.decodeResult(result, output); err != nil {
		logger.Error("Error invoking evm function: can't unpack result", "err", err, "maxgas", bm.maxGas)
		return err
	}

	logger.Trace("EVM call successful", "input", hexutil.Encode(input), "output", hexutil.Encode(output))
	return nil
}

func unpackError(result []byte) (string, error) {
	if len(result) < 4 || !bytes.Equal(result[:4], errorSig) {
		return "<tx result not Error(string)>", errors.New("TX result not of type Error(string)")
	}
	vs, err := abi.Arguments{{Type: abiString}}.UnpackValues(result[4:])
	if err != nil {
		return "<invalid tx result>", err
	}
	return vs[0].(string), nil
}
