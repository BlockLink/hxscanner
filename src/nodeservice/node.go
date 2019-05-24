package nodeservice

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	netrpc "net/rpc"
	"time"

	"github.com/blocklink/hxscanner/wsjsonrpc/jsonrpc"
	"golang.org/x/net/websocket"
	"github.com/blocklink/hxscanner/src/types"
)

var _ws *websocket.Conn = nil
var _client *netrpc.Client = nil

func ConnectHxNode(ctx context.Context, apiUrl string) error {
	origin := apiUrl
	url := apiUrl
	var err error
	_ws, err = websocket.Dial(url, "", origin)
	if err != nil {
		return err
	}
	_client = jsonrpc.NewClient(_ws)
	go func() {
		for {
			select {
			case <-time.After(time.Second * 30):
				if _ws != nil {
					_ws.Write([]byte("ping"))
				} else {
					break
				}
			}
			if _ws == nil {
				break
			}
		}
	}()
	go func() {
		select {
		case <-ctx.Done():
			log.Fatal(CloseHxNodeConn())
		}
	}()
	return nil
}

func CloseHxNodeConn() (err error) {
	if _ws != nil {
		err = _ws.Close()
		_ws = nil
		_client = nil
	}
	return
}

func IsHxNodeConnected() bool {
	return _ws != nil
}

func GetKeysOfJson(val map[string]interface{}) []string {
	result := make([]string, 0)
	for k, _ := range val {
		result = append(result, k)
	}
	return result
}

func GetBlock(blockNum int) (block *types.HxBlock, err error) {
	if !IsHxNodeConnected() {
		log.Println("ws to hx_node disconnected")
		return
	}
	var wrapperReply interface{}
	var reply = new(types.HxBlock)
	c := _client
	err = c.Call("get_block", blockNum, &wrapperReply)
	if err != nil {
		if err.Error() == "error <nil>" {
			return nil, nil
		}
		log.Println("get_block error " + err.Error())
		return
	}
	replyJSONBytes, err := json.Marshal(&wrapperReply)
	replyJSONBytesDecoder := json.NewDecoder(bytes.NewReader(replyJSONBytes))
	replyJSONBytesDecoder.UseNumber()
	err = replyJSONBytesDecoder.Decode(&reply)
	if err != nil {
		log.Println("decode reply error: " + err.Error())
		return
	}
	block = reply
	block.BlockNumber = blockNum
	for i, tx := range block.Transactions {
		tx.IndexInBlock = i
	}
	//if len(block.Transactions) > 0 {
	//	log.Println(string(replyJSONBytes), err)
	//}
	// fetch transaction ids
	var fullTxsReply = make([]types.HxFullTransactionExtraInfo, 0)
	err = c.Call("fetch_block_transactions", blockNum, &fullTxsReply)
	if err != nil {
		log.Println("fetch_block_transactions error", err)
		return
	}
	block.TransactionIds = make([]string, 0)
	for i, info := range fullTxsReply {
		block.TransactionIds = append(block.TransactionIds, info.Trxid)
		if i < len(block.Transactions) {
			block.Transactions[i].Trxid = info.Trxid
			block.Transactions[i].BlockNum = info.BlockNum
			block.Transactions[i].ContractId = info.ContractId
		}
	}
	return
}

func IsContractOpType(operationType int) bool {
	return operationType >= 76 && operationType <= 81
}

func CheckTransactionHasContractOp(txInfo *types.HxTransaction) bool {
	for _, operationItemArray := range txInfo.Operations {
		if len(operationItemArray) >= 2 {
			if operationTypeNum, ok := operationItemArray[0].(json.Number); ok {
				operationType, err := operationTypeNum.Int64()
				if err == nil {
					if IsContractOpType(int(operationType)) {
						return true
					}
				}
			}
		}
	}
	return false
}


func GetTxReceipts(txInfo *types.HxTransaction) (txReceipts *types.HxContractTxReceipt, err error) {
	if !IsHxNodeConnected() {
		log.Println("ws to hx_node disconnected")
		return
	}
	txReceipts = new(types.HxContractTxReceipt)
	txReceipts.OpReceipts = make([]*types.HxContractOpReceipt, 0)
	c := _client
	err = c.Call("get_contract_invoke_object", txInfo.Trxid, &(txReceipts.OpReceipts))
	if err != nil {
		log.Println("get_contract_invoke_object error: " + err.Error())
		return
	}
	replyJSONBytes, err := json.Marshal(&txReceipts.OpReceipts)
	_ = replyJSONBytes
	// log.Println(string(replyJSONBytes), err)
	var hasFailedContractOperation = false
	for i := 0; i < len(txReceipts.OpReceipts); i++ {
		operationResult := txReceipts.OpReceipts[i]
		if operationResult != nil {
			if !operationResult.ExecSucceed {
				hasFailedContractOperation = true
				break
			}
		}
	}
	txReceipts.HasFailedContractOperation = hasFailedContractOperation
	return
}

func FindHxTransactionByTxid(txid string) (state string) {
	state = "TxStateNotFound"
	if !IsHxNodeConnected() {
		log.Println("ws to hx_node disconnected")
		return
	}
	//args := TxidArgs{txid}
	var reply = make(map[string]interface{})
	c := _client
	err := c.Call("get_transaction_by_id", txid, &reply)
	if err != nil {
		log.Println("get_transaction_by_id error", err)
		return
	}
	state = "TxStateSuccess"

	replyJSONBytes, err := json.Marshal(&reply)
	//log.Println(string(replyJSONBytes), err)
	replyJSONBytesDecoder := json.NewDecoder(bytes.NewReader(replyJSONBytes))
	replyJSONBytesDecoder.UseNumber()
	err = replyJSONBytesDecoder.Decode(&reply)
	if err != nil {
		log.Println("decode reply error: " + err.Error())
		return
	}
	// if it's contract tx, get_contract_invoke_object txid to query exec_succeed(whether fail)
	var hasContractOp = false
	if operationsJSON, ok := reply["operations"]; ok {
		if operations, ok := operationsJSON.([]interface{}); ok {
			for i := 0; i < len(operations); i++ {
				operationJSON := operations[i]
				if operationItemArray, ok := operationJSON.([]interface{}); ok {
					if len(operationItemArray) >= 2 {
						if operationTypeNum, ok := operationItemArray[0].(json.Number); ok {
							operationType, err := operationTypeNum.Int64()
							if err == nil {
								if operationType >= 76 && operationType <= 81 {
									hasContractOp = true
									break
								}
							}
						}
					}
				}
			}
		}
	}
	if hasContractOp {
		var contractInvokeObjectReply = []map[string]interface{}{}
		err := c.Call("get_contract_invoke_object", txid, &contractInvokeObjectReply)
		if err != nil {
			log.Println("get_contract_invoke_object error: " + err.Error())
			return
		}
		replyJSONBytes, err := json.Marshal(&contractInvokeObjectReply)
		_ = replyJSONBytes
		// log.Println(string(replyJSONBytes), err)
		var hasFailedContractOperation = false
		for i := 0; i < len(contractInvokeObjectReply); i++ {
			operationResult := contractInvokeObjectReply[i]
			if operationResult != nil {
				if execSucceedObj, ok := operationResult["exec_succeed"]; ok {
					if execSuccess, ok := execSucceedObj.(bool); ok {
						if !execSuccess {
							hasFailedContractOperation = true
							break
						}
					}
				}
			}
		}
		if hasFailedContractOperation {
			state = "TxStateFail"
		}
	}
	return
}
