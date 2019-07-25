package plugins

import (
	"github.com/blocklink/hxscanner/src/types"
	"errors"
)

type TransferPlugin struct {

}

func (plugin *TransferPlugin) PluginName() string {
	return "TransferPlugin"
}

func (plugin *TransferPlugin) ApplyOperation(block *types.HxBlock, txid string, opNum int, opType int, opTypeName string,
	opJSON map[string]interface{}, receipt *types.HxContractOpReceipt) (err error) {
	if opTypeName != "transfer_operation" {
		return
	}
	toAddr, ok := mapGetString(opJSON, "to_addr")
	if !ok {
		err = errors.New("invalid to_addr in transfer op")
		return
	}
	fromAddr, ok := mapGetString(opJSON, "from_addr")
	if !ok {
		err = errors.New("invalid from_addr in transfer op")
		return
	}
	amountObj, ok := opJSON["amount"]
	if !ok {
		err = errors.New("invalid amount type in transfer op")
		return
	}
	amountMap, ok := amountObj.(map[string]interface{})
	if !ok {
		err = errors.New("invalid amount type in transfer op")
		return
	}
	assetId, ok := mapGetString(amountMap, "asset_id")
	if !ok {
		return
	}
	err = updateAddressBalance(fromAddr, assetId)
	if err != nil {
		return
	}
	err = updateAddressBalance(toAddr, assetId)
	if err != nil {
		return
	}
	return
}
