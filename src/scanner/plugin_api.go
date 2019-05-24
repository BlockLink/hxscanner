package scanner

import (
	"github.com/blocklink/hxscanner/src/types"
)

type OpScannerPlugin interface {
	ApplyOperation(block *types.HxBlock, txid string, opType int, opTypeName string, opJSON map[string]interface{}, receipt *types.HxContractOpReceipt) (err error)
}
