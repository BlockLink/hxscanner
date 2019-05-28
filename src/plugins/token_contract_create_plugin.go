package plugins

import (
	"github.com/blocklink/hxscanner/src/types"
	"log"
	"errors"
	"github.com/blocklink/hxscanner/src/db"
)

type TokenContractCreateScanPlugin struct {

}

type contractRegisterOperation struct {
	ContractCode map[string]interface{}
	Abi []string
	OfflineAbi []string
	OwnerPubKey string
	OwnerAddr string
	InheritFrom string
	RegisterTime string
	ContractId string
	GasPrice uint64
	GasLimit uint64
}

func decodeTokenRegisterOperation(opJSON map[string]interface{}) (result *contractRegisterOperation, err error) {
	result = new(contractRegisterOperation)
	contractCode, ok := opJSON["contract_code"] // json of {abi: Array[string], offline_abi: Array[string], storage_properties: Array[Array[string]], code_hash: "", events: Array[string]}
	if !ok {
		err = errors.New("contract_code not found")
		return
	}
	contractCodeMap, ok := contractCode.(map[string]interface{})
	if !ok {
		err = errors.New("invalid contract_code map")
		return
	}
	result.ContractCode = contractCodeMap
	abiObj, ok := contractCodeMap["abi"]
	if !ok {
		err = errors.New("abi not found")
		return
	}
	abi, ok := objToStringArray(abiObj)
	if !ok {
		err = errors.New("abi not array<string>")
		return
	}
	result.Abi = abi
	offlineAbiObj, ok := contractCodeMap["offline_abi"]
	if !ok {
		err = errors.New("invalid offline_abi")
		return
	}
	offlineAbi, ok := objToStringArray(offlineAbiObj)
	if !ok {
		err = errors.New("invalid offline_abi")
		return
	}
	result.OfflineAbi = offlineAbi
	ownerPubkey, ok := getStringPropFromJSONObj(opJSON, "owner_pubkey")
	if !ok {
		err = errors.New("owner_pubkey not found")
		return
	}
	result.OwnerPubKey = ownerPubkey
	registerTime, ok := getStringPropFromJSONObj(opJSON, "register_time")
	if !ok {
		err = errors.New("register_time not found")
		return
	}
	result.RegisterTime = registerTime
	ownerAddr, ok := getStringPropFromJSONObj(opJSON, "owner_addr")
	if !ok {
		err = errors.New("owner_addr not found")
		return
	}
	result.OwnerAddr = ownerAddr
	inheritFrom, ok := getStringPropFromJSONObj(opJSON, "inherit_from")
	if !ok {
		err = errors.New("inherit_from not found")
		return
	}
	result.InheritFrom = inheritFrom
	gasPrice, ok := getIntPropFromJSONObj(opJSON, "gas_price")
	if !ok {
		err = errors.New("gasPrice not found")
		return
	}
	result.GasPrice = uint64(gasPrice)
	gasLimit, ok := getIntPropFromJSONObj(opJSON, "init_cost")
	if !ok {
		err = errors.New("init_cost not found")
		return
	}
	result.GasLimit = uint64(gasLimit)
	contractId, ok := getStringPropFromJSONObj(opJSON, "contract_id")
	if !ok {
		err = errors.New("contract_id not found")
		return
	}
	result.ContractId = contractId
	return
}

func (plugin *TokenContractCreateScanPlugin) ApplyOperation(block *types.HxBlock, txid string, opNum int, opType int, opTypeName string,
	opJSON map[string]interface{}, receipt *types.HxContractOpReceipt) (err error) {
	if opTypeName == "contract_register_operation" {
		// contract register
		log.Println("found a new contract in tx " + txid)
		if receipt==nil || !receipt.ExecSucceed {
			log.Println("receive fail contract tx " + txid)
			return
		}
		var contractOp *contractRegisterOperation
		contractOp, err = decodeTokenRegisterOperation(opJSON)
		if err != nil {
			log.Println("decode token register operation error", err)
			return
		}
		if !isAllInArray([]string{"transfer", "transferFrom", "approve"}, contractOp.Abi) {
			log.Println("not token contract")
			return
		}
		if !isAllInArray([]string{"balanceOf", "totalSupply", "precision", "approvedBalanceFrom"}, contractOp.OfflineAbi) {
			log.Println("not token contract")
			return
		}

		log.Println("found a token contract")
		// TODO: 调用合约查询token的基本属性
		// save to db
		var dbTokenContract *db.TokenContractEntity
		dbTokenContract, err = db.FindTokenContractByContractId(contractOp.ContractId)
		if err != nil {
			return
		}
		if dbTokenContract == nil {
			dbTokenContract = &db.TokenContractEntity{
				BlockNum:     uint32(block.BlockNumber),
				BlockTime:    block.Timestamp,
				Txid:         txid,
				ContractId:   contractOp.ContractId,
				ContractType: "common",
				OwnerPubkey:  contractOp.OwnerPubKey,
				OwnerAddr:    contractOp.OwnerAddr,
				RegisterTime: block.Timestamp,
				InheritFrom:  contractOp.InheritFrom,
				GasPrice:     uint64(contractOp.GasPrice),
				GasLimit:     uint64(contractOp.GasLimit),
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