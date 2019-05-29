package plugins

import (
	"github.com/blocklink/hxscanner/src/types"
	"github.com/blocklink/hxscanner/src/db"
	"encoding/json"
	"log"
	"time"
	"github.com/shopspring/decimal"
	"fmt"
)

// 扫描token合约的,init_token后触发事件导致state变化也要扫描. transfer记录，得到合约转账记录历史信息等

type TokenContractInvokeScanPlugin struct {
}

func (plugin *TokenContractInvokeScanPlugin) ApplyOperation(block *types.HxBlock, txid string, opNum int, opType int, opTypeName string,
	opJSON map[string]interface{}, receipt *types.HxContractOpReceipt) (err error) {
	if receipt == nil || !receipt.ExecSucceed {
		return
	}
	contractIsTokenCache := make(map[string]bool)
	isTokenContract := func(addr string) (bool, error) {
		isToken, ok := contractIsTokenCache[addr]
		if ok {
			return isToken, nil
		}
		dbTokenContract, err := db.FindTokenContractByContractId(addr)
		if err != nil {
			return false, err
		}
		return dbTokenContract != nil, nil
	}
	fmtLayout := "2006-01-02T15:04:05"
	txTime, err := time.Parse(fmtLayout, block.Timestamp)
	if err != nil {
		log.Println("parse block timestamp error", err)
		return
	}
	for _, event := range receipt.Events {
		contractId := event.ContractAddress
		var isToken bool
		isToken, err = isTokenContract(contractId)
		if err != nil {
			return
		}
		if !isToken {
			continue
		}
		eventName := event.EventName
		eventArg := event.EventArg
		eventArgBytes := []byte(eventArg)
		switch eventName {
		case "Inited":
			{
				tokenContract, err := db.FindTokenContractByContractId(contractId)
				if err != nil {
					log.Println("find token contract error", err)
					continue
				}
				if tokenContract != nil {
					state := "common"
					tokenContract.State = &state
					err = db.UpdateTokenContract(tokenContract)
					if err != nil {
						log.Println("update token contract error", err)
						continue
					}
				}
			}
		case "Transfer":
			{
				// event arg: {from: ..., to: ..., amount}
				type transferEventArgType struct {
					From   string `json:"from"`
					To     string `json:"to"`
					Amount uint64 `json:"amount"`
				}
				transferArg := new(transferEventArgType)
				err = json.Unmarshal(eventArgBytes, transferArg)
				if err != nil {
					log.Println("invalid transfer token event arg " + eventArg)
					continue
				}
				historyItem, err := db.FindTokenContractTransferHistoryItemByTxIdAndOpNum(txid, opNum)
				if err != nil {
					log.Println("find token transfer history error", err)
					continue
				}
				transferAmountDecimal, err := decimal.NewFromString(fmt.Sprintf("%d", transferArg.Amount))
				if err != nil {
					log.Println("decimal from int error", err)
					continue
				}
				if historyItem == nil {
					// save transfer tx to token history table
					now := time.Now()
					historyItem = &db.TokenContractTransferHistoryEntity{
						ContractAddr: contractId,
						FromAddr: transferArg.From,
						ToAddr: transferArg.To,
						Amount: transferAmountDecimal,
						BlockNum: uint32(block.BlockNumber),
						Txid: txid,
						OpNum: uint32(opNum),
						EventName: eventName,
						TxTime: txTime,
						CreatedAt: now,
						UpdatedAt: now}
					err = db.SaveTokenContractTransferHistory(historyItem)
					if err != nil {
						log.Println("save token transfer history error", err)
						continue
					}
				}
				// TODO: query and save from/to users(maybe same or empty) new token balance
			}
		case "Approved":
			{
				// TODO
			}
		default:
			continue
		}
	}
	return
}
