package db

import (
	_ "github.com/bmizerany/pq"
	"database/sql"
	"log"
	"math/big"
	"github.com/shopspring/decimal"
	"time"
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
	Addr string
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

type TokenContractEntity struct {
	Id int64
	BlockNum uint32
	BlockTime string
	Txid string
	ContractId string
	ContractType string
	OwnerPubkey string
	OwnerAddr string
	RegisterTime string
	InheritFrom string
	GasPrice uint64
	GasLimit uint64
	State *string
	TotalSupply *big.Int
	Precision *uint32
	TokenSymbol *string
	TokenName *string
	Logo *string
	Url *string
	Description *string
}

// token合约各用户的余额
type TokenBalanceEntity struct {
	Id int64
	ContractAddr string
	OwnerAddr string
	Amount decimal.Decimal
	CreatedAt time.Time
	UpdatedAt time.Time
}

// token合约的转账历史记录
type TokenContractTransferHistoryEntity struct {
	Id int64
	ContractAddr string
	FromAddr string
	ToAddr string
	Amount decimal.Decimal
	BlockNum uint32
	Txid string
	OpNum uint32
	EventName string
	TxTime time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}
