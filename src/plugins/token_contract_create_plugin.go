package plugins

import (
	"github.com/blocklink/hxscanner/src/types"
	"log"
	"encoding/json"
	"github.com/blocklink/hxscanner/src/db"
)

type TokenContractCreateScanPlugin struct {

}

func isStringInArray(item string, arr []string) bool {
	for _, val := range arr {
		if val == item {
			return true
		}
	}
	return false
}

func isAllInArray(items []string, arr []string) bool {
	for _, item := range items {
		if !isStringInArray(item, arr) {
			return false
		}
	}
	return true
}

func objArrayToStringArray(src []interface{}) (result []string, ok bool) {
	for _, item := range src {
		itemStr, isStr := item.(string)
		if !isStr {
			ok = false
			return
		}
		result = append(result, itemStr)
	}
	ok = true
	return
}

func objToStringArray(src interface{}) (result []string, ok bool) {
	objArray, ok := src.([]interface{})
	if !ok {
		return
	}
	result, ok = objArrayToStringArray(objArray)
	return
}

func getStringPropFromJSONObj(jsonObj map[string]interface{}, prop string) (result string, ok bool) {
	item, ok := jsonObj[prop]
	if !ok {
		return
	}
	result, ok = item.(string)
	return
}

func getIntPropFromJSONObj(jsonObj map[string]interface{}, prop string) (result int64, isInt bool) {
	itemObj, ok := jsonObj[prop]
	if !ok {
		return
	}
	if item, ok := itemObj.(int64); ok {
		result = item
		isInt = true
		return
	}
	if item, ok := itemObj.(int32); ok {
		result = int64(item)
		isInt = true
		return
	}
	if item, ok := itemObj.(json.Number); ok {
		itemInt, err := item.Int64()
		if err != nil {
			isInt = false
			return
		}
		result = itemInt
		isInt = true
		return
	}
	isInt = false
	return
}

func (plugin *TokenContractCreateScanPlugin) ApplyOperation(block *types.HxBlock, txid string, opType int, opTypeName string,
	opJSON map[string]interface{}, receipt *types.HxContractOpReceipt) (err error) {
	if opTypeName == "contract_register_operation" {
		// contract register
		log.Println("found a new contract in tx " + txid)
		if receipt==nil || !receipt.ExecSucceed {
			log.Println("receive fail contract tx " + txid)
			return
		}
		contractCode, ok := opJSON["contract_code"] // json of {abi: Array[string], offline_abi: Array[string], storage_properties: Array[Array[string]], code_hash: "", events: Array[string]}
		if !ok {
			log.Println("contract_code not found")
			return
		}
		contractCodeMap, ok := contractCode.(map[string]interface{})
		if !ok {
			log.Println("invalid contract_code map")
			return
		}
		abiObj, ok := contractCodeMap["abi"]
		if !ok {
			log.Println("abi not found")
			return
		}
		abi, ok := objToStringArray(abiObj)
		if !ok {
			log.Println("abi not array<string>")
			return
		}
		if !isAllInArray([]string{"transfer", "transferFrom", "approve"}, abi) {
			log.Println("not token contract")
			return
		}
		offlineAbiObj, ok := contractCodeMap["offline_abi"]
		if !ok {
			return
		}
		offlineAbi, ok := objToStringArray(offlineAbiObj)
		if !ok {
			return
		}
		if !isAllInArray([]string{"balanceOf", "totalSupply", "precision", "approvedBalanceFrom"}, offlineAbi) {
			log.Println("not token contract")
			return
		}

		log.Println("found a token contract")

		ownerPubkey, ok := getStringPropFromJSONObj(opJSON, "owner_pubkey")
		if !ok {
			log.Println("owner_pubkey not found")
			return
		}
		registerTime, ok := getStringPropFromJSONObj(opJSON, "register_time")
		if !ok {
			log.Println("register_time not found")
			return
		}
		_ = registerTime
		ownerAddr, ok := getStringPropFromJSONObj(opJSON, "owner_addr")
		if !ok {
			log.Println("owner_addr not found")
			return
		}
		inheritFrom, ok := getStringPropFromJSONObj(opJSON, "inherit_from")
		if !ok {
			log.Println("inherit_from not found")
			return
		}
		gasPrice, ok := getIntPropFromJSONObj(opJSON, "gas_price")
		if !ok {
			log.Println("gasPrice not found")
			return
		}
		gasLimit, ok := getIntPropFromJSONObj(opJSON, "init_cost")
		if !ok {
			log.Println("init_cost not found")
			return
		}
		contractId, ok := getStringPropFromJSONObj(opJSON, "contract_id")
		if !ok {
			log.Println("contract_id not found")
			return
		}
		// TODO: 调用合约查询token的基本属性
		// save to db
		var dbTokenContract *db.TokenContractEntity
		dbTokenContract, err = db.FindTokenContractByContractId(contractId)
		if err != nil {
			return
		}
		if dbTokenContract == nil {
			dbTokenContract = &db.TokenContractEntity{
				BlockNum:     uint32(block.BlockNumber),
				BlockTime:    block.Timestamp,
				Txid:         txid,
				ContractId:   contractId,
				ContractType: "common",
				OwnerPubkey:  ownerPubkey,
				OwnerAddr:    ownerAddr,
				RegisterTime: block.Timestamp,
				InheritFrom:  inheritFrom,
				GasPrice:     uint64(gasPrice),
				GasLimit:     uint64(gasLimit),
				State:        nil,
				TotalSupply:  nil,
				Precision:    nil,
				TokenSymbol:  nil,
				TokenName:    nil,
				Logo:         nil,
				Url:          nil,
				Description:  nil}
			err = db.SaveTokenContract(dbTokenContract)
			if err != nil {
				return
			}
		}
	} else if opTypeName == "native_contract_register_operation" {
		// native contract register
		log.Println("found a new native contract in tx " + txid)
		if receipt==nil || !receipt.ExecSucceed {
			log.Println("receive fail contract tx " + txid)
			return
		}
		// TODO
	} else {
		return
	}
	return
}