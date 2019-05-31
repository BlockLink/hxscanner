package scanner

import (
	"github.com/blocklink/hxscanner/src/nodeservice"
	"log"
	"strconv"
	"time"
	"reflect"
	"encoding/json"
	"github.com/blocklink/hxscanner/src/db"
	"fmt"
	"context"
	"github.com/blocklink/hxscanner/src/types"
)

var tableSchemaCache = make(map[string]*db.PgTableSchema)

func cachedGetTableSchema(tableName string) (result *db.PgTableSchema, err error) {
	var ok bool
	result, ok = tableSchemaCache[tableName]
	if ok {
		return result, nil
	}
	result, err = db.GetTableSchema(tableName)
	if err == nil && result != nil {
		tableSchemaCache[tableName] = result
	}
	return
}

var scanPlugins = make([]OpScannerPlugin, 0)

func AddScanPlugin(plugin OpScannerPlugin) {
	scanPlugins = append(scanPlugins, plugin)
}

func ApplyPluginsToOperation(block *types.HxBlock, txid string, opIndex int, opType int, opTypeName string, opJSON map[string]interface{}, receipt *types.HxContractOpReceipt) (err error) {
	for _, plugin := range scanPlugins {
		err = plugin.ApplyOperation(block, txid, opIndex, opType, opTypeName, opJSON, receipt)
		if err != nil {
			return
		}
	}
	return
}

func ScanBlocksFrom(ctx context.Context, startBlockNum int) {
	scannedBlockNum := startBlockNum
	end := false
	for ;; {
		if end {
			break
		}
		select {
		case <- ctx.Done():
			end = true
			break
		default:

		}
		// TODO: fetch 100 blocks concurrency
		// TODO: use channel to produce fetched-blocks goroutine and store goroutine
		block, err := nodeservice.GetBlock(scannedBlockNum)
		if err != nil {
			log.Println("got block at #" + strconv.Itoa(scannedBlockNum) + " with error "+ err.Error())
			break
		}
		if block == nil {
			log.Println("scanned block #" + strconv.Itoa(scannedBlockNum-1))
			err = db.UpdateLastScannedBlockNumber(scannedBlockNum-1)
			if err != nil {
				log.Fatal("UpdateLastScannedBlockNumber error " + err.Error())
				break
			}
			time.Sleep(5 * time.Second)
			continue
		}
		// find or replace same records
		// save block
		oldBlock, err := db.FindBlock(block.BlockNumber)
		if err != nil {
			log.Println("find block at #" + strconv.Itoa(scannedBlockNum) + " with error "+ err.Error())
			break
		}
		if oldBlock == nil {
			err = db.SaveBlock(block)
			if err != nil {
				log.Println("save block to db error " + err.Error())
				break
			}
		}
		// 取到block后，修改它上一个块的block_hash
		if block.BlockNumber > 1 {
			prevBlock, err := db.FindBlock(block.BlockNumber-1)
			if err != nil {
				log.Println("find block at #" + strconv.Itoa(block.BlockNumber - 1) + " with error " + err.Error())
				break
			}
			if prevBlock != nil && (prevBlock.BlockId == "" || prevBlock.BlockId=="TODO") {
				prevBlock.BlockId = block.Previous
				err = db.UpdateBlockHash(int(prevBlock.Number), prevBlock.BlockId)
				if err != nil {
					log.Println("UpdateBlock #" + strconv.Itoa(int(prevBlock.Number)) + " error")
					break
				}
			}
		}

		for txIndex := 0;txIndex < len(block.Transactions);txIndex++ {
			txInfo := block.Transactions[txIndex]
			//log.Println("tx index " + strconv.Itoa(txIndex) + " trxid " + txInfo.Trxid)
			txHasContractOp := nodeservice.CheckTransactionHasContractOp(txInfo)
			var txReceipts *types.HxContractTxReceipt = nil
			if txHasContractOp {
				txReceipts, err = nodeservice.GetTxReceipts(txInfo)
				if err != nil {
					log.Println("get tx receipts when txid " + txInfo.Trxid + " error " +err.Error())
					break
				}
			}

			// find or save txs
			oldTx, err := db.FindTransaction(txInfo.Trxid)
			if err != nil {
				log.Println("find tx at #" + strconv.Itoa(scannedBlockNum) + " with error "+ err.Error())
				break
			}
			if oldTx == nil {
				err = db.SaveTransaction(txInfo)
				if err != nil {
					log.Println("save tx to db error " + err.Error())
					break
				}
			}

			for opIndex := 0;opIndex < len(txInfo.Operations);opIndex++ {
				opPair := txInfo.Operations[opIndex]
				if len(opPair) != 2 {
					log.Println("invalid operation pair size(require 2 and got " + strconv.Itoa(len(opPair)) + ")")
					break
				}
				var ok bool
				var opTypeNumber json.Number
				if opTypeNumber, ok = opPair[0].(json.Number); !ok {
					t := reflect.TypeOf(opPair[0])
					log.Println("invalid operation type type " + t.Name())
					break
				}
				opTypeInt, err := strconv.Atoi(opTypeNumber.String())
				if err != nil {
					log.Println("parse operation type error " + err.Error())
					break
				}
				var opJson map[string]interface{}
				if opJson, ok = opPair[1].(map[string]interface{}); !ok {
					log.Println("invalid operation json type")
					break
				}
				opJson["block_num"] = block.BlockNumber
				opJson["trxid"] = txInfo.Trxid
				opJson["index_in_tx"] = opIndex
				for _, extraKey := range []string{"memo", "guarantee_id"} {
					if _, ok := opJson[extraKey]; !ok {
						opJson[extraKey] = ""
					}
				}
				opTypeName, err := nodeservice.GetOperationNameByOperationType(opTypeInt)
				if err != nil {
					log.Println("get operation name error " + err.Error())
					break
				}
				//operationKeys := nodeservice.GetKeysOfJson(opJson)
				//log.Println("operation " + opTypeName + " has " + strconv.Itoa(len(operationKeys)) + " keys")
				operationTableName := nodeservice.GetOperationTableNameByOperationName(opTypeName)
				// create operation table and find-or-save operation columns and save operations
				opTableExist, err := db.CheckTableExist(operationTableName)
				if err != nil {
					log.Println("CheckTableExist error " + err.Error())
					break
				}
				if !opTableExist {
					opTableColumnSqls := make([]string, 0)
					for opKey, opColVal := range opJson {
						var opColumnTypeSql string
						switch opColVal.(type) {
						case int, uint32, int64, uint64:
							opColumnTypeSql = "bigint"
						default:
							opColumnTypeSql = "text"
						}
						colSql := fmt.Sprintf("\"%s\" %s NULL", opKey, opColumnTypeSql)
						opTableColumnSqls = append(opTableColumnSqls, colSql)
					}
					err = db.CreateTable(operationTableName, opTableColumnSqls, "")
					if err != nil {
						log.Fatal("create operation table " + operationTableName + " error " + err.Error())
						break
					}
					// create op index
					indexName := fmt.Sprintf("%s_idx", operationTableName)
					createIndexSql := fmt.Sprintf("CREATE INDEX %s ON %s(trxid, index_in_tx)", indexName, operationTableName)
					err = db.ExecSql(createIndexSql)
					if err != nil {
						log.Fatal("create operation table index error " + err.Error())
						break
					}
				}
				opExistInDb, err := db.CheckOperationExist(operationTableName, txInfo.Trxid, opIndex)
				if err != nil {
					log.Fatal("CheckOperationExist when txid=" + txInfo.Trxid + " op #" + strconv.Itoa(opIndex) + " error " + err.Error())
					break
				}
				if !opExistInDb {
					// save operation
					opTableSchema, err := cachedGetTableSchema(operationTableName)
					if err != nil {
						log.Fatal("get operation table schema error " + err.Error())
						break
					}
					err = db.InsertDynamicOperation(operationTableName, opTableSchema, opJson)
					if err != nil {
						log.Fatal("InsertDynamicOperation to table " + operationTableName + " error " + err.Error())
						break
					}
				}
				// insert into base operations table
				baseOperation := new(db.BaseOperationEntity)
				baseOperation.OperationType = opTypeInt
				baseOperation.OperationTypeName = opTypeName
				baseOperation.TxIndexInBlock = txIndex
				baseOperation.BlockNum = block.BlockNumber
				baseOperation.Trxid = txInfo.Trxid
				opJSONBytes, err := json.Marshal(opJson)
				if err != nil {
					log.Fatal("json marshal operation error "+ err.Error())
					break
				}
				baseOperation.OperationJSON = string(opJSONBytes)
				var addr string
				var addrObj interface{}
				var addrFound = false
				maybeAddrProps := []string {"addr", "from_addr", "caller_addr", "owner_addr", "miner_address",
				"payer", "fee_paying_account", "lock_balance_addr", "pay_back_owner", "bonus_owner",
				"fee_pay_address", "publisher_addr", "addr_from_claim", "issuer_addr"}
				for _, prop := range maybeAddrProps {
					if !addrFound {
						addrObj, addrFound = opJson[prop]
					}
				}
				if !addrFound {
					addrObj = ""
				}
				addr, addrFound = addrObj.(string)
				if !addrFound {
					addr = ""
				}
				baseOperation.Addr = addr
				baseOperation.Id = db.GetBaseOperationId(baseOperation.BlockNum, baseOperation.Trxid, opIndex)
				oldBaseOpInDb, err := db.FindBaseOperation(baseOperation.Id)
				if err != nil {
					log.Println("find base operation at " +baseOperation.Id+ " with error "+ err.Error())
					break
				}
				if oldBaseOpInDb == nil {
					err = db.SaveBaseOperation(baseOperation)
					if err != nil {
						log.Fatal("SaveBaseOperation error " + err.Error())
						break
					}
				}
				var receipt *types.HxContractOpReceipt = nil
				if txReceipts != nil && len(txReceipts.OpReceipts) > opIndex {
					receipt = txReceipts.OpReceipts[opIndex]
				}
				err = ApplyPluginsToOperation(block, txInfo.Trxid, opIndex, opTypeInt, opTypeName, opJson, receipt)
				if err != nil {
					log.Fatal("apply plugin to op error", err)
					break
				}
			}
			if txHasContractOp && txReceipts != nil {
				for _, opReceipt := range txReceipts.OpReceipts {
					oldDbOpReceipt, err := db.FindContractOpReceipt(opReceipt.Trxid, opReceipt.OpNum)
					if err != nil {
						log.Fatal("FindContractOpReceipt error " + err.Error())
						break
					}
					if oldDbOpReceipt == nil {
						err = db.SaveContractOpReceipt(opReceipt)
						if err != nil {
							log.Fatal("SaveContractOpReceipt error " + err.Error())
							break
						}
					}
				}
			}
		}
		if scannedBlockNum % 100 == 0 {
			log.Println("scanned block #" + strconv.Itoa(scannedBlockNum))
			err = db.UpdateLastScannedBlockNumber(scannedBlockNum)
			if err != nil {
				log.Fatal("UpdateLastScannedBlockNumber error " + err.Error())
				break
			}
		}
		scannedBlockNum++

		time.Sleep(time.Millisecond * 1)
	}
}
