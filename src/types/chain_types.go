package types


type TxidArgs struct {
	Txid string `json:"txid"`
}

type HxTransaction struct {
	BlockNum         uint32          `json:"block_num"`
	Trxid            string          `json:"trxid"`
	ContractId       string          `json:"contract_id"`
	IndexInBlock     int             `json:"index_in_block"`
	Expiration       string          `json:"expiration"`
	Extensions       []interface{}   `json:"extensions"`
	OperationResults []interface{}   `json:"operation_results"`
	Operations       [][]interface{} `json:"operations"` // every item is [operationTypeInt, operationJson]
	RefBlockNum      uint32          `json:"ref_block_num"`
	RefBlockPrefix   uint64          `json:"ref_block_prefix"`
	Signatures       []string        `json:"signatures"`
}

type HxFullTransactionExtraInfo struct {
	BlockNum   uint32 `json:"block_num"`
	Trxid      string `json:"trxid"`
	ContractId string `json:"contract_id"`
}

type HxBlock struct {
	BlockNumber           int              `json:"block_number"`
	Extensions            []interface{}    `json:"extensions"`
	Miner                 string           `json:"miner"`
	MinerSignature        string           `json:"miner_signature"`
	NextSecretHash        string           `json:"next_secret_hash"`
	Previous              string           `json:"previous"`
	PreviousSecret        string           `json:"previous_secret"`
	Timestamp             string           `json:"timestamp"`
	TransactionMerkleRoot string           `json:"transaction_merkle_root"`
	Transactions          []*HxTransaction `json:"transactions"`
	TransactionIds        []string         `json:"transaction_ids"`
	Trxfee                int              `json:"trxfee"`
}


