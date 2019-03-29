package scanner

import (
	"github.com/zoowii/hxscanner/src/nodeservice"
	"log"
	"strconv"
	"time"
	"reflect"
	"encoding/json"
	"github.com/zoowii/hxscanner/src/db"
	"fmt"
)

var tableSchemaCache = make(map[string]*db.PgTableSchema)

func cachedGetTableSchema(tableName string) (result *db.PgTableSchema, err error) {
	var ok bool
	result, ok = tableSchemaCache[tableName]
	if ok {
		return result, nil
	}
	result, err = db.GetTableSchema(tableName)
	return
}

func ScanBlocksFrom(startBlockNum int, quit <-chan int) {
	scannedBlockNum := startBlockNum
	end := false
	for ;; {
		if end {
			break
		}
		select {
		case <- quit:
			end = true
			break
		default:

		}
		// TODO: fetch 100 blocks concurrency
		block, err := nodeservice.GetBlock(scannedBlockNum)
		if err != nil {
			log.Println("got block at #" + strconv.Itoa(scannedBlockNum) + " with error "+ err.Error())
			break
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
		for txIndex := 0;txIndex < len(block.Transactions);txIndex++ {
			txInfo := block.Transactions[txIndex]
			//log.Println("tx index " + strconv.Itoa(txIndex) + " trxid " + txInfo.Trxid)
			txHasContractOp := nodeservice.CheckTransactionHasContractOp(txInfo)
			var txReceipts *nodeservice.HxContractTxReceipt = nil
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