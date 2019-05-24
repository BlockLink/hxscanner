package types


type HxContractOpReceiptEvent struct {
	BlockNum        uint32 `json:"block_num"`
	CallerAddr      string `json:"caller_addr"`
	ContractAddress string `json:"contract_address"`
	EventArg        string `json:"event_arg"`
	EventName       string `json:"event_name"`
	OpNum           int    `json:"op_num"`
	Trxid           string `json:"trx_id"`
}

type HxContractOpReceipt struct {
	Id                       string                      `json:"id"`
	Trxid                    string                      `json:"trx_id"`
	BlockNum                 uint32                      `json:"block_num"`
	OpNum                    int                         `json:"op_num"`
	ApiResult                string                      `json:"api_result"`
	Events                   []*HxContractOpReceiptEvent `json:"events"`
	ExecSucceed              bool                        `json:"exec_succeed"`
	ActualFee                uint64                      `json:"acctual_fee"`
	Invoker                  string                      `json:"invoker"`
	ContractRegistered       string                      `json:"contract_registed"`
	ContractWithdrawInfo     []interface{}               `json:"contract_withdraw"`
	ContractBalanceChanges   []interface{}               `json:"contract_balances"`
	DepositToAddressChanges  []interface{}               `json:"deposit_to_address"`
	DepositToContractChanges []interface{}               `json:"deposit_contract"`
	TransferFees             []interface{}               `json:"transfer_fees"`
}

func NewHxContractOpReceipt() *HxContractOpReceipt {
	result := new(HxContractOpReceipt)
	result.Events = make([]*HxContractOpReceiptEvent, 0)
	result.ContractWithdrawInfo = make([]interface{}, 0)
	result.ContractBalanceChanges = make([]interface{}, 0)
	result.DepositToAddressChanges = make([]interface{}, 0)
	result.DepositToContractChanges = make([]interface{}, 0)
	result.TransferFees = make([]interface{}, 0)
	return result
}

type HxContractTxReceipt struct {
	OpReceipts                 []*HxContractOpReceipt
	HasFailedContractOperation bool
}
