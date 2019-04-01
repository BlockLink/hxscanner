package db

import (
	_ "github.com/bmizerany/pq"
	"database/sql"
	"log"
)

var dbConn *sql.DB = nil

func OpenDb(connStr string) (err error) {
	dbConn, err = sql.Open("postgres", connStr)
	return
}

func CloseDb() {
	if dbConn != nil {
		err := dbConn.Close()
		if err != nil {
			log.Println("close db connection error " + err.Error())
		}
		dbConn = nil
	}
}

type BlockEntity struct {
	Id int64
	Number uint32
	Previous string
	Timestamp string
	Trxfee uint64
	Miner string
	TransactionMerkleRoot string
	NextSecretHash string
	BlockId string
	Reward uint64
	TxsCount int
}

type TransactionEntity struct {
	Id string
	RefBlockNum uint64
	RefBlockPrefix uint64
	Expiration string
	OperationsCount int
	IndexInBlock int
	FirstOperationType int
	Txid string
}

type BaseOperationEntity struct {
	Id string
	Trxid string
	BlockNum int
	TxIndexInBlock int
	OperationType int
	OperationTypeName string
	OperationJSON string
}

type ScanConfigEntity struct {
	Id int64
	ConfigKey string
	ConfigValue string
}

type PgTableSchemaColumn struct {
	ColumnName string
	ColumnType string
}

type PgTableSchema struct {
	Columns []*PgTableSchemaColumn
}

func (schema *PgTableSchema) HasColumn(name string) bool {
	for _, col := range schema.Columns {
		if col.ColumnName == name {
			return true
		}
	}
	return false
}