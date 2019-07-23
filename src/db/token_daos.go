package db

import (
	"time"
	"github.com/shopspring/decimal"
)

func SaveTokenBalance(tokenBalance *TokenBalanceEntity) error {
	now := time.Now()
	stmt, err := dbConn.Prepare("INSERT INTO public.token_balance (contract_addr, owner_addr," +
		" amount, created_at, updated_at)" +
		" VALUES (($1),($2),($3),($4),($5))")
	if err != nil {
		return err
	}
	defer stmt.Close()
	res, err := stmt.Exec(tokenBalance.ContractAddr, tokenBalance.OwnerAddr, tokenBalance.Amount.String(), now.Unix(), now.Unix())
	if err != nil {
		return err
	}
	_ = res
	//lastId, err := res.LastInsertId()
	//if err != nil {
	//	return err
	//}
	//logger.Println("last tx record id " + strconv.Itoa(int(lastId)))
	return nil
}

func SaveTokenContractTransferHistory(record *TokenContractTransferHistoryEntity) error {
	now := time.Now()
	stmt, err := dbConn.Prepare("INSERT INTO public.token_contract_transfer_history (contract_addr, from_addr," +
		" to_addr, amount, block_num, txid, op_num, event_name, tx_time, created_at, updated_at)" +
		" VALUES (($1),($2),($3),($4),($5), $6, $7, $8, $9, $10, $11)")
	if err != nil {
		return err
	}
	defer stmt.Close()
	res, err := stmt.Exec(record.ContractAddr, record.FromAddr, record.ToAddr, record.Amount.String(), record.BlockNum,
		record.Txid, record.OpNum, record.EventName, record.TxTime.Unix(), now.Unix(), now.Unix())
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

func UpdateTokenBalance(tokenBalance *TokenBalanceEntity) error {
	stmt, err := dbConn.Prepare("UPDATE public.token_balance SET contract_addr = $1, owner_addr = $2, amount = $3," +
		"created_at = $4, updated_at = $5 WHERE id=$6")
	if err != nil {
		return err
	}
	defer stmt.Close()
	res, err := stmt.Exec(tokenBalance.ContractAddr, tokenBalance.OwnerAddr, tokenBalance.Amount.String(), tokenBalance.CreatedAt.Unix(), tokenBalance.UpdatedAt.Unix(), tokenBalance.Id)
	if err != nil {
		return err
	}
	_ = res
	return nil
}

func UpdateTokenContractTransferHistory(record *TokenContractTransferHistoryEntity) error {
	stmt, err := dbConn.Prepare("UPDATE public.token_contract_transfer_history SET contract_addr = $1, from_addr = $2, to_addr = $3, amount = $4," +
		"block_num = $5, txid = $6, op_num = $7, event_name = $8, tx_time = $9," +
		"created_at = $10, updated_at = $11 WHERE id=$12")
	if err != nil {
		return err
	}
	defer stmt.Close()
	res, err := stmt.Exec(record.ContractAddr, record.FromAddr, record.ToAddr, record.Amount.String(), record.BlockNum,
		record.Txid, record.OpNum, record.EventName, record.TxTime.Unix(), record.CreatedAt.Unix(), record.UpdatedAt.Unix(), record.Id)
	if err != nil {
		return err
	}
	_ = res
	return nil
}

func FindTokenBalanceByContractAddrAndOwnerAddr(contractAddr string, ownerAddr string) (result *TokenBalanceEntity, err error) {
	rows, err := dbConn.Query("SELECT id, contract_addr, owner_addr, amount, created_at, updated_at FROM public.token_balance where contract_addr=$1 and owner_addr=$2", contractAddr, ownerAddr)
	if err != nil {
		return
	}
	defer rows.Close()
	if rows.Next() {
		result = new(TokenBalanceEntity)
		var amountStr string
		var createdAtUnix, updatedAtUnix int64
		err = rows.Scan(&result.Id, &result.ContractAddr, &result.OwnerAddr, &amountStr,
			&createdAtUnix, &updatedAtUnix)
		if err != nil {
			return
		}
		result.Amount, err = decimal.NewFromString(amountStr)
		if err != nil {
			return
		}
		result.CreatedAt = time.Unix(createdAtUnix, 0)
		result.UpdatedAt = time.Unix(updatedAtUnix, 0)
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

func FindTokenContractTransferHistoryItemByTxIdAndOpNum(txid string, opNum int) (result *TokenContractTransferHistoryEntity, err error) {
	rows, err := dbConn.Query("SELECT id, contract_addr, from_addr," +
		" to_addr, amount, block_num, txid, op_num, event_name, tx_time, created_at, updated_at" +
		" FROM public.token_contract_transfer_history where txid=$1 and op_num=$2", txid, opNum)
	if err != nil {
		return
	}
	defer rows.Close()
	if rows.Next() {
		result = new(TokenContractTransferHistoryEntity)
		var amountStr string
		var txTimeUnix, createdAtUnix, updatedAtUnix int64
		err = rows.Scan(&result.Id, &result.ContractAddr, &result.FromAddr, &result.ToAddr, &amountStr, &result.BlockNum,
			&result.Txid, &result.OpNum, &result.EventName, &txTimeUnix,
			&createdAtUnix, &updatedAtUnix)
		if err != nil {
			return
		}
		result.Amount, err = decimal.NewFromString(amountStr)
		if err != nil {
			return
		}
		result.CreatedAt = time.Unix(createdAtUnix, 0)
		result.UpdatedAt = time.Unix(updatedAtUnix, 0)
		result.TxTime = time.Unix(txTimeUnix, 0)
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
