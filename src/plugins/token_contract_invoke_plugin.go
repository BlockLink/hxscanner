package plugins

import (
	"github.com/blocklink/hxscanner/src/types"
	"github.com/blocklink/hxscanner/src/db"
	"encoding/json"
	"log"
	"time"
	"github.com/shopspring/decimal"
	"fmt"
	"github.com/blocklink/hxscanner/src/nodeservice"
	"github.com/blocklink/hxscanner/src/config"
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
				var historyItem *db.TokenContractTransferHistoryEntity
				historyItem, err = db.FindTokenContractTransferHistoryItemByTxIdAndOpNum(txid, opNum)
				if err != nil {
					log.Println("find token transfer history error", err)
					continue
				}
				var transferAmountDecimal decimal.Decimal
				transferAmountDecimal, err = decimal.NewFromString(fmt.Sprintf("%d", transferArg.Amount))
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
				// query and save from/to users(maybe same or empty) new token balance
				usersToUpdate := make([]string, 0)
				if len(transferArg.From) > 0 {
					usersToUpdate = append(usersToUpdate, transferArg.From)
				}
				if len(transferArg.To) > 0 && transferArg.To != transferArg.From {
					usersToUpdate = append(usersToUpdate, transferArg.To)
				}
				now := time.Now()
				for _, userAddr := range usersToUpdate {
					var userBalance int64
					userBalance, err = nodeservice.InvokeContractOfflineWithIntResult(config.SystemConfig.CallerPubKeyString, contractId, "balanceOf", userAddr)
					if err != nil {
						log.Println("query token balance of " + userAddr + " in contract " + contractId + " error")
						continue
					}
					var tokenBalanceItem *db.TokenBalanceEntity
					tokenBalanceItem, err = db.FindTokenBalanceByContractAddrAndOwnerAddr(contractId, userAddr)
					if err != nil {
						log.Println("FindTokenBalanceByContractAddrAndOwnerAddr error", err)
						return
					}
					var userBalanceDecimal decimal.Decimal
					userBalanceDecimal, err = decimal.NewFromString(fmt.Sprintf("%d", userBalance))
					if err != nil {
						return
					}
					if tokenBalanceItem == nil {
						tokenBalanceItem = &db.TokenBalanceEntity{
							ContractAddr: contractId,
							OwnerAddr: userAddr,
							Amount: userBalanceDecimal,
							CreatedAt: now,
							UpdatedAt: now}
						err = db.SaveTokenBalance(tokenBalanceItem)
						if err != nil {
							log.Println("SaveTokenBalance error", err)
							return
						}
					} else {
						tokenBalanceItem.UpdatedAt = now
						tokenBalanceItem.Amount = userBalanceDecimal
						err = db.UpdateTokenBalance(tokenBalanceItem)
						if err != nil {
							log.Println("UpdateTokenBalance error", err)
							return
						}
					}

				}
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
