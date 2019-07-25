package scanner

import (
	"github.com/blocklink/hxscanner/src/types"
)

type OpScannerPlugin interface {
	PluginName() string
	ApplyOperation(block *types.HxBlock, txid string, opNum int, opType int, opTypeName string, opJSON map[string]interface{}, receipt *types.HxContractOpReceipt) (err error)
}
