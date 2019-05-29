package config

type Config struct {
	NodeApiUrl string
	DbConnectionString string
	CallerPubKeyString string
}

var SystemConfig *Config

const LastScannedBlockNumberConfigKey = "last_scanned_block_number"
