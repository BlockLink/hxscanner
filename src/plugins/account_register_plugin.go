package plugins

import (
	"github.com/blocklink/hxscanner/src/types"
	"github.com/blocklink/hxscanner/src/db"
	"time"
)

type AccountRegisterPlugin struct {}

func (plugin *AccountRegisterPlugin) PluginName() string {
	return "AccountRegisterPlugin"
}

func (plugin *AccountRegisterPlugin) ApplyOperation(block *types.HxBlock, txid string, opNum int, opType int, opTypeName string,
	opJSON map[string]interface{}, receipt *types.HxContractOpReceipt) (err error) {
	if opTypeName != "account_create_operation" {
		return
	}
	accountName, ok := mapGetString(opJSON, "name")
	if !ok {
		return
	}
	payerAddr, ok := mapGetString(opJSON, "payer")
	if !ok {
		return
	}
	accountEntity, err := db.FindAccountByOwnerAddr(payerAddr)
	if err != nil {
		return
	}
	now := time.Now()
	if accountEntity == nil {
		accountEntity = &db.AccountEntity{OwnerAddr:payerAddr, AccountName:accountName, CreatedAt: now, UpdatedAt:now}
		err = db.SaveAccount(accountEntity)
		if err != nil {
			return
		}
	}
	return
}

// TODO: 账户名的转移, 账户的角色变更