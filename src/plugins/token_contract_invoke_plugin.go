package plugins

import (
	"github.com/blocklink/hxscanner/src/types"
	"github.com/blocklink/hxscanner/src/db"
	"encoding/json"
	"time"
	"github.com/shopspring/decimal"
	"fmt"
	"github.com/blocklink/hxscanner/src/nodeservice"
	"github.com/blocklink/hxscanner/src/config"
	"math/big"
)

// 扫描token合约的,init_token后触发事件导致state变化也要扫描. transfer记录，得到合约转账记录历史信息等

type TokenContractInvokeScanPlugin struct {
}

func (plugin *TokenContractInvokeScanPlugin) PluginName() string {
	return "TokenContractInvokeScanPlugin"
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
		logger.Println("parse block timestamp error", err)
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
					logger.Println("find token contract error", err)
					continue
				}
				if tokenContract != nil {
					state := "common"
					tokenContract.State = &state

					// query tokenName, tokenSymbol, totalSupply, precision
					precision := new(uint32)
					tokenName := new(string)
					tokenSymbol := new(string)
					*tokenName, *tokenSymbol, *precision, err = queryTokenContractBaseInfo(contractId)
					if err == nil {
						tokenContract.TokenName = tokenName
						tokenContract.TokenSymbol = tokenSymbol
						tokenContract.Precision = precision
					}
					totalSupply := new(int64)
					*totalSupply, err = queryTokenContractTotalSupply(contractId)
					if err == nil {
						totalSupplyBig := big.NewInt(*totalSupply)
						tokenContract.TotalSupply = totalSupplyBig
					}

					err = db.UpdateTokenContract(tokenContract)
					if err != nil {
						logger.Println("update token contract error", err)
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
					logger.Println("invalid transfer token event arg " + eventArg)
					continue
				}
				var historyItem *db.TokenContractTransferHistoryEntity
				historyItem, err = db.FindTokenContractTransferHistoryItemByTxIdAndOpNum(txid, opNum)
				if err != nil {
					logger.Println("find token transfer history error", err)
					continue
				}
				var transferAmountDecimal decimal.Decimal
				transferAmountDecimal, err = decimal.NewFromString(fmt.Sprintf("%d", transferArg.Amount))
				if err != nil {
					logger.Println("decimal from int error", err)
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
						logger.Println("save token transfer history error", err)
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
						logger.Println("query token balance of " + userAddr + " in contract " + contractId + " error")
						continue
					}
					var tokenBalanceItem *db.TokenBalanceEntity
					tokenBalanceItem, err = db.FindTokenBalanceByContractAddrAndOwnerAddr(contractId, userAddr)
					if err != nil {
						logger.Println("FindTokenBalanceByContractAddrAndOwnerAddr error", err)
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
							logger.Println("SaveTokenBalance error", err)
							return
						}
					} else {
						tokenBalanceItem.UpdatedAt = now
						tokenBalanceItem.Amount = userBalanceDecimal
						err = db.UpdateTokenBalance(tokenBalanceItem)
						if err != nil {
							logger.Println("UpdateTokenBalance error", err)
							return
						}
					}
				}
				// if fromAddr or toAddr is empty, query totalSupply from node
				if len(transferArg.From) < 1 || len(transferArg.To) < 1 {
					var tokenContract *db.TokenContractEntity
					tokenContract, err = db.FindTokenContractByContractId(contractId)
					if err == nil {
						totalSupply := new(int64)
						*totalSupply, err = queryTokenContractTotalSupply(contractId)
						if err == nil {
							totalSupplyBig := big.NewInt(*totalSupply)
							tokenContract.TotalSupply = totalSupplyBig
							err = db.UpdateTokenContract(tokenContract)
							if err != nil {
								logger.Println("update token contract totalSupply error", err)
								return
							}
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
	// process contract withdraws
	for _, change := range receipt.DepositToAddressChanges {
		var ok bool
		changeItem, ok := change.([]interface{})
		if !ok {
			continue
		}
		if len(changeItem) < 2 {
			continue
		}
		addressAssetPair, ok := changeItem[0].([]interface{})
		if !ok {
			continue
		}
		if len(addressAssetPair) < 2 {
			continue
		}
		addr, ok := addressAssetPair[0].(string)
		if !ok {
			continue
		}
		assetId, ok := addressAssetPair[1].(string)
		if !ok {
			continue
		}
		//amount, ok := changeItem[1].(json.Number)
		//if !ok {
		//	amountT := reflect.TypeOf(changeItem[1])
		//	logger.Println("invalid deposit_address type " + amountT.String())
		//	continue
		//}
		err = updateAddressBalance(addr, assetId)
		if err != nil {
			return
		}
	}
	return
}
