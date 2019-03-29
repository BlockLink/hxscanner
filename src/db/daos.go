package db

import (
	"github.com/zoowii/hxscanner/src/nodeservice"
	"encoding/json"
	"github.com/pkg/errors"
	"fmt"
	"strings"
	"github.com/zoowii/hxscanner/src/config"
	"strconv"
	"log"
)

func GetScanConfigOr(configKey string, elseValue string) (string, error) {
	config, err := FindScanConfig(configKey)
	if err != nil {
		return "", err
	}
	if config == nil {
		return elseValue, nil
	}
	return config.ConfigValue, nil
}

func GetLastScannedBlockNumber() (uint32, error) {
	configStr, err := GetScanConfigOr(config.LastScannedBlockNumberConfigKey, "0")
	if err != nil {
		return 0, err
	}
	configInt, err := strconv.Atoi(configStr)
	if err != nil {
		return 0, err
	}
	return uint32(configInt), nil
}

func UpdateLastScannedBlockNumber(newVal int) error {
	newValStr := strconv.Itoa(newVal)
	configEntity, err := FindScanConfig(config.LastScannedBlockNumberConfigKey)
	if err != nil {
		return err
	}
	if configEntity == nil {
		return SaveConfig(config.LastScannedBlockNumberConfigKey, newValStr)
	}
	if configEntity.ConfigValue == newValStr {
		return nil
	}
	configEntity.ConfigValue = newValStr
	return UpdateConfig(configEntity)
}

func FindScanConfig(configKey string) (result *ScanConfigEntity, err error) {
	rows, err := dbConn.Query("SELECT id, config_key, config_value FROM public.scan_configs where config_key=$1", configKey)
	if err != nil {
		return
	}
	defer rows.Close()
	if rows.Next() {
		result = new(ScanConfigEntity)
		err = rows.Scan(&result.Id, &result.ConfigKey, &result.ConfigValue)
		if err != nil {
			return
		}
		return
	} else {
		result = nil
		err = nil
		return
	}
	err = rows.Err()
	if err != nil {
		return
	}
	return
}

func FindContractOpReceipt(trxid string, opNum int) (result *nodeservice.HxContractOpReceipt, err error) {
	rows, err := dbConn.Query("SELECT id, trxid, block_num, op_num, api_result, events, exec_succeed, actual_fee, invoker, contract_registered, contract_withdraw_info, contract_balance_changes, deposit_to_address_changes, deposit_to_contract_changes, transfer_fees FROM public.contract_operation_receipt where trxid=$1 and op_num=$2", trxid, opNum)
	if err != nil {
		return
	}
	defer rows.Close()
	if rows.Next() {
		result = nodeservice.NewHxContractOpReceipt()
		var eventsStr, contractWithdrawInfoStr, contractBalancesChangesStr, depositToAddressChangesStr, depositToContractChangesStr, transferFeesStr string
		err = rows.Scan(&result.Id, &result.Trxid, &result.BlockNum, &result.OpNum, &result.ApiResult, &eventsStr, &result.ExecSucceed, &result.ActualFee, &result.Invoker, &result.ContractRegistered, &contractWithdrawInfoStr, &contractBalancesChangesStr, &depositToAddressChangesStr, &depositToContractChangesStr, &transferFeesStr)
		if err != nil {
			return
		}
		if len(eventsStr) > 0 {
			err = json.Unmarshal([]byte(eventsStr), &result.Events)
			if err != nil {
				return
			}
		}
		if len(contractWithdrawInfoStr) > 0 {
			err = json.Unmarshal([]byte(contractWithdrawInfoStr), &result.ContractWithdrawInfo)
			if err != nil {
				return
			}
		}
		if len(contractBalancesChangesStr) > 0 {
			err = json.Unmarshal([]byte(contractBalancesChangesStr), &result.ContractBalanceChanges)
			if err != nil {
				return
			}
		}
		if len(depositToAddressChangesStr) > 0 {
			err = json.Unmarshal([]byte(depositToAddressChangesStr), &result.DepositToAddressChanges)
			if err != nil {
				return
			}
		}
		if len(depositToContractChangesStr) > 0 {
			err = json.Unmarshal([]byte(depositToContractChangesStr), &result.DepositToContractChanges)
			if err != nil {
				return
			}
		}
		if len(transferFeesStr) > 0 {
			err = json.Unmarshal([]byte(transferFeesStr), &result.TransferFees)
			if err != nil {
				return
			}
		}
		return
	} else {
		result = nil
		err = nil
		return
	}
	err = rows.Err()
	if err != nil {
		return
	}
	return
}

func FindBlock(blockNumber int) (result *BlockEntity, err error) {
	rows, err := dbConn.Query("SELECT id, number, previous, timestamp, trxfee, miner, transaction_merkle_root, next_secret_hash, block_id, reward, txs_count FROM public.blocks where number=$1", blockNumber)
	if err != nil {
		return
	}
	defer rows.Close()
	if rows.Next() {
		result = new(BlockEntity)
		err = rows.Scan(&result.Id, &result.Number, &result.Previous, &result.Timestamp, &result.Trxfee, &result.Miner, &result.TransactionMerkleRoot, &result.NextSecretHash, &result.BlockId, &result.Reward, &result.TxsCount)
		if err != nil {
			return
		}
		return
	} else {
		return
	}
	err = rows.Err()
	if err != nil {
		return
	}
	return
}

func CheckOperationExist(tableName string, trxid string, indexInTx int) (result bool, err error) {
	rows, err := dbConn.Query("SELECT * FROM public." + tableName + " where trxid=$1 and index_in_tx=$2", trxid, indexInTx)
	if err != nil {
		return
	}
	defer rows.Close()
	if rows.Next() {
		result = true
		return
	} else {
		return
	}
	err = rows.Err()
	if err != nil {
		return
	}
	return
}

func GetTableSchema(tableName string) (result *PgTableSchema, err error) {
	rows, err := dbConn.Query("select column_name, data_type from INFORMATION_SCHEMA.COLUMNS where table_name =$1", tableName)
	if err != nil {
		return
	}
	defer rows.Close()
	result = new (PgTableSchema)
	result.Columns = make([]*PgTableSchemaColumn, 0)
	for rows.Next() {
		col := new(PgTableSchemaColumn)
		err = rows.Scan(&col.ColumnName, &col.ColumnType)
		if err != nil {
			return
		}
		result.Columns = append(result.Columns, col)
	}
	err = rows.Err()
	if err != nil {
		return
	}
	return
}

func FindTransaction(txid string) (result *TransactionEntity, err error) {
	rows, err := dbConn.Query("SELECT id, ref_block_num, ref_block_prefix, expiration, operations_count, index_in_block, first_operation_type, txid FROM public.transactions where txid=$1", txid)
	if err != nil {
		return
	}
	defer rows.Close()
	if rows.Next() {
		result = new(TransactionEntity)
		err = rows.Scan(&result.Id, &result.RefBlockNum, &result.RefBlockPrefix, &result.Expiration, &result.OperationsCount, &result.IndexInBlock, &result.FirstOperationType, &result.Txid)
		if err != nil {
			return
		}
		return
	} else {
		return
	}
	err = rows.Err()
	if err != nil {
		return
	}
	return
}

func SaveBlock(block *nodeservice.HxBlock) error {
	stmt, err := dbConn.Prepare("INSERT INTO public.blocks (id, number, previous, timestamp, trxfee, miner, transaction_merkle_root, next_secret_hash, block_id, reward, txs_count) VALUES (($1),($2),($3),($4),($5),($6),($7),($8),($9),($10),($11) )")
	if err != nil {
		return err
	}
	defer stmt.Close()
	res, err := stmt.Exec(block.BlockNumber, block.BlockNumber, block.Previous, block.Timestamp, block.Trxfee, block.Miner, block.TransactionMerkleRoot, block.NextSecretHash, "", 0, len(block.Transactions))
	if err != nil {
		return err
	}
	_ = res
	//lastId, err := res.LastInsertId()
	//if err != nil {
	//	return err
	//}
	//log.Println("last block record id " + strconv.Itoa(int(lastId)))
	return nil
}

func SaveConfig(configKey string, configValue string) error {
	stmt, err := dbConn.Prepare("INSERT INTO public.scan_configs (config_key, config_value) VALUES (($1),($2) )")
	if err != nil {
		return err
	}
	defer stmt.Close()
	res, err := stmt.Exec(configKey, configValue)
	if err != nil {
		return err
	}
	_ = res
	//lastId, err := res.LastInsertId()
	//if err != nil {
	//	return err
	//}
	//log.Println("last config record id " + strconv.Itoa(int(lastId)))
	return nil
}

func SaveContractOpReceipt(contractOpReceipt *nodeservice.HxContractOpReceipt) error {
	stmt, err := dbConn.Prepare("INSERT INTO public.contract_operation_receipt (trxid, block_num, op_num, api_result,  exec_succeed, actual_fee, invoker, contract_registered, events, contract_withdraw_info, contract_balance_changes, deposit_to_address_changes, deposit_to_contract_changes, transfer_fees) VALUES (($1),($2),($3),($4),($5),($6),($7),($8),($9),($10),($11),($12),($13),($14) )")
	if err != nil {
		return err
	}
	defer stmt.Close()
	eventsBytes, err := json.Marshal(contractOpReceipt.Events)
	if err != nil {
		return err
	}
	contractWithdrawInfoBytes, err := json.Marshal(contractOpReceipt.ContractWithdrawInfo)
	if err != nil {
		return err
	}
	contractBalanceChangesBytes, err := json.Marshal(contractOpReceipt.ContractBalanceChanges)
	if err != nil {
		return err
	}
	depositToAddressChangesBytes, err := json.Marshal(contractOpReceipt.DepositToAddressChanges)
	if err != nil {
		return err
	}
	depositToContractChangesBytes, err := json.Marshal(contractOpReceipt.DepositToContractChanges)
	if err != nil {
		return err
	}
	transferFeesBytes, err := json.Marshal(contractOpReceipt.TransferFees)
	if err != nil {
		return err
	}
	_, err = stmt.Exec(contractOpReceipt.Trxid, contractOpReceipt.BlockNum, contractOpReceipt.OpNum, contractOpReceipt.ApiResult, contractOpReceipt.ExecSucceed, contractOpReceipt.ActualFee, contractOpReceipt.Invoker, contractOpReceipt.ContractRegistered, string(eventsBytes), string(contractWithdrawInfoBytes), string(contractBalanceChangesBytes), string(depositToAddressChangesBytes), string(depositToContractChangesBytes), string(transferFeesBytes))
	if err != nil {
		return err
	}
	// save events to single table
	if contractOpReceipt.Events != nil {
		for _, event := range contractOpReceipt.Events {
			stmt2, err := dbConn.Prepare("INSERT INTO public.contract_operation_receipt_event (trxid, block_num, op_num, caller_addr, contract_address, event_arg, event_name) VALUES (($1),($2),($3),($4),($5),($6),($7) )")
			if err != nil {
				return err
			}
			defer stmt2.Close()
			_, err = stmt2.Exec(contractOpReceipt.Trxid, event.BlockNum, event.OpNum, event.CallerAddr, event.ContractAddress, event.EventArg, event.EventName)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func UpdateConfig(configEntity *ScanConfigEntity) error {
	stmt, err := dbConn.Prepare("UPDATE public.scan_configs SET config_value = $1 WHERE config_key=$2")
	if err != nil {
		return err
	}
	defer stmt.Close()
	res, err := stmt.Exec(configEntity.ConfigValue, configEntity.ConfigKey)
	if err != nil {
		return err
	}
	_ = res
	return nil
}


func SaveTransaction(tx *nodeservice.HxTransaction) error {
	stmt, err := dbConn.Prepare("INSERT INTO public.transactions (id, ref_block_num, ref_block_prefix, expiration, operations_count, index_in_block, first_operation_type, txid) VALUES (($1),($2),($3),($4),($5),($6),($7),($8) )")
	if err != nil {
		return err
	}
	defer stmt.Close()
	var firstOpType int = -1
	if len(tx.Operations) > 0 {
		var ok bool
		firstOpPair := tx.Operations[0]
		firstOpTypeObj := firstOpPair[0]
		var firstOpTypeNum json.Number
		if firstOpTypeNum, ok = firstOpTypeObj.(json.Number); !ok {
			return errors.New("invalid operation type")
		}
		firstOpTypeInt64, err := firstOpTypeNum.Int64()
		if err != nil {
			return err
		}
		firstOpType = int(firstOpTypeInt64)
	}
	res, err := stmt.Exec(tx.Trxid, tx.RefBlockNum, tx.RefBlockPrefix, tx.Expiration, len(tx.Operations), tx.IndexInBlock, firstOpType, tx.Trxid)
	if err != nil {
		return err
	}
	_ = res
	//lastId, err := res.LastInsertId()
	//if err != nil {
	//	return err
	//}
	//log.Println("last tx record id " + strconv.Itoa(int(lastId)))
	return nil
}

func CheckTableExist(tableName string) (bool, error) {
	rows, err := dbConn.Query("SELECT table_name FROM information_schema.tables WHERE table_type = 'BASE TABLE' and table_name=$1 AND table_schema NOT IN ('pg_catalog', 'information_schema')", tableName)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	if rows.Next() {
		return true, nil
	} else {
		return false, nil
	}
	err = rows.Err()
	if err != nil {
		return false, err
	}
	return false, nil
}

func ExecSql(sql string) error {
	stmt, err := dbConn.Prepare(sql)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec()
	if err != nil {
		return err
	}
	return nil
}

func CreateTable(tableName string, columnDefinitions []string, extraSql string) error {
	columnDefsSql := strings.Join(columnDefinitions, ", ")
	var extendSql string
	if len(extraSql) > 0 {
		extendSql = ", " + extraSql
	} else {
		extendSql = ""
	}
	sql := fmt.Sprintf("CREATE TABLE \"%s\" (%s %s)", tableName, columnDefsSql, extendSql)
	return ExecSql(sql)
}

func InsertDynamicOperation(tableName string, tableSchema *PgTableSchema, opJson map[string]interface{}) error {
	opTableColumnNameSqls := make([]string, 0)
	prepareValueSqls := make([]string, 0)
	opValuesForSql := make([]interface{}, 0)
	for opKey, opColVal := range opJson {
		if !tableSchema.HasColumn(opKey) {
			continue
		}
		var opValueForSql interface{}
		switch opColVal.(type) {
		case int, uint32, int64, uint64:
			opValueForSql = opColVal
		case string, json.Number:
			opValueForSql = opColVal
		default:
			valBytes, err := json.Marshal(opColVal)
			if err != nil {
				return err
			}
			opValueForSql = string(valBytes)
		}
		colNameSql := fmt.Sprintf("\"%s\"", opKey)
		opTableColumnNameSqls = append(opTableColumnNameSqls, colNameSql)
		prepareValueSqls = append(prepareValueSqls, fmt.Sprintf(" $%d ", len(prepareValueSqls) + 1))
		opValuesForSql = append(opValuesForSql, opValueForSql)
	}
	columnsSql := strings.Join(opTableColumnNameSqls, ",")
	sql := fmt.Sprintf("INSERT INTO public.%s (%s) VALUES (%s)", tableName, columnsSql, strings.Join(prepareValueSqls, ","))
	stmt, err := dbConn.Prepare(sql)
	if err != nil {
		log.Println("insert sql " + sql)
		return err
	}
	defer stmt.Close()
	res, err := stmt.Exec(opValuesForSql...)
	if err != nil {
		log.Println("insert sql " + sql)
		return err
	}
	_ = res
	return nil
}