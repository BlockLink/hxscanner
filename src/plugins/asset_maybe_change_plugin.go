package plugins

import (
	"github.com/blocklink/hxscanner/src/types"
)

type AssetMaybeChangePlugin struct {}

func (plugin *AssetMaybeChangePlugin) PluginName() string {
	return "AssetMaybeChangePlugin"
}

func mapGetString(m map[string]interface{}, key string) (val string, ok bool) {
	valObj, ok := m[key]
	if ok {
		val, ok = valObj.(string)
	}
	return
}

// 有amount的字段就可能触发调用者的余额更新
func (plugin *AssetMaybeChangePlugin) ApplyOperation(block *types.HxBlock, txid string, opNum int, opType int, opTypeName string,
	opJSON map[string]interface{}, receipt *types.HxContractOpReceipt) (err error) {
	tryAddrProps := []string{"addr", "caller_addr", "lock_balance_addr", "foreclose_addr"}
	tryAssetIdProps := []string {"asset_id", "lock_asset_id", "foreclose_asset_id"}
	var addr string
	var ok bool
	for _, prop := range tryAddrProps {
		addr, ok = mapGetString(opJSON, prop)
		if ok {
			break
		}
	}
	if !ok || len(addr)<1 {
		return
	}
	var assetId string
	for _, prop := range tryAssetIdProps {
		assetId, ok = mapGetString(opJSON, prop)
		if ok {
			break
		}
	}
	if !ok || len(assetId)<1 {
		amountMapObj, ok := opJSON["amount"]
		if !ok {
			return
		}
		amountMap, ok := amountMapObj.(map[string]interface{})
		if !ok {
			return
		}
		assetId, ok = mapGetString(amountMap, "asset_id")
		if !ok {
			return
		}
	}
	err = updateAddressBalance(addr, assetId)
	if err != nil {
		return
	}
	return
}

