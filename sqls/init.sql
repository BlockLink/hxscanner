CREATE TABLE "blocks" (
    id integer NOT NULL,
    number bigint NOT NULL,
    previous text NULL,
    timestamp text NULL,
    trxfee numeric(20,0) NOT NULL,
    miner text NULL,
    transaction_merkle_root text NULL,
    next_secret_hash text NULL,
    block_id text NULL,
    reward numeric(20,0) NOT NULL,
    txs_count integer NOT NULL,
    CONSTRAINT "pk_blocks" PRIMARY KEY (id)
);

CREATE INDEX blocks_idx ON blocks(number);

CREATE TABLE "citizen_infos" (
    id serial NOT NULL,
    citizen_id text NULL,
    account_id text NULL,
    last_citizen_fee integer NOT NULL,
    last_mined_block_num bigint NOT NULL,
    last_mined_block_id text NULL,
    last_collect_miss_time timestamp without time zone NOT NULL,
    last_collect_miss_count integer NOT NULL,
    last_collect_produced_count integer NOT NULL,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL,
    CONSTRAINT "pk_citizen_infos" PRIMARY KEY (id)
);

CREATE TABLE "operations" (
    serial_id serial NOT NULL,
    id text NOT NULL,
    txid text NULL,
    tx_block_number bigint NOT NULL,
    tx_index_in_block bigint NOT NULL,
    operation_type bigint NOT NULL,
    operation_type_name varchar(255) NOT NULL,
    operation_json text NULL,
    addr text NULL,
    CONSTRAINT "pk_operations" PRIMARY KEY (serial_id)
);

CREATE INDEX operations_txid_idx ON operations (txid);

CREATE TABLE "scan_configs" (
    id serial NOT NULL,
    config_key text NULL,
    config_value text NULL,
    CONSTRAINT "pk_scan_configs" PRIMARY KEY (id)
);

CREATE INDEX scan_configs_key_idx ON scan_configs (config_key);

CREATE TABLE "transactions" (
    serial_id serial NOT NULL,
    block_number integer NOT NULL,
    id text NOT NULL,
    ref_block_num bigint NOT NULL,
    ref_block_prefix numeric(20,0) NOT NULL,
    expiration text NULL,
    operations_count integer NOT NULL,
    index_in_block bigint NOT NULL,
    first_operation_type bigint NOT NULL,
    txid text NULL,
    CONSTRAINT "pk_transactions" PRIMARY KEY (serial_id)
);

CREATE INDEX transactions_idx ON transactions(txid);

CREATE TABLE "update_account_options_operations" (
    id serial NOT NULL,
    txid text NULL,
    index_in_tx integer NOT NULL,
    operation_id text NULL,
    account_id text NULL,
    memo_key text NULL,
    voting_account text NULL,
    miner_pledge_pay_back integer NOT NULL,
    trx_time timestamp without time zone NOT NULL,
    CONSTRAINT "pk_update_account_options_operations" PRIMARY KEY (id)
);

CREATE TABLE "contract_operation_receipt" (
  id serial NOT NULL,
  trxid text NOT NULL,
	block_num integer NOT NULL,
	op_num integer NOT NULL,
	api_result text NOT NULL,
	events text NULL,
	exec_succeed bool NOT NULL,
	actual_fee bigint NOT NULL,
	invoker text NOT NULL,
	contract_registered text NULL,
	contract_withdraw_info text NULL,
	contract_balance_changes text NULL,
	deposit_to_address_changes text NULL,
	deposit_to_contract_changes text NULL,
	transfer_fees TEXT NULL,
  CONSTRAINT "pk_contract_operation_receipt" PRIMARY KEY (id)
);

CREATE INDEX contract_operation_receipt_idx ON contract_operation_receipt (trxid, op_num);

CREATE TABLE "contract_operation_receipt_event" (
  id serial NOT NULL,
  trxid text NOT NULL,
	block_num integer NOT NULL,
	op_num integer NOT NULL,
  caller_addr text NOT NULL,
  contract_address text NOT NULL,
  event_arg text NOT NULL,
  event_name text NOT NULL,
  CONSTRAINT "pk_contract_operation_receipt_event" PRIMARY KEY (id)
);

CREATE TABLE "token_contract" (
  id serial NOT NULL,
  block_num integer NOT NULL,
  block_time varchar(100) NOT NULL,
  txid varchar(100) NOT NULL,
  contract_id varchar(100) NOT NULL,
  contract_type varchar(100) NOT NULL,
  owner_pubkey varchar(200) NOT NULL,
  owner_addr varchar(100) NOT NULL,
  register_time varchar(100) NOT NULL,
  inherit_from varchar(100) NOT NULL,
  gas_price integer NOT NULL,
  gas_limit integer NOT NULL,
  state varchar(100) NULL,
  total_supply varchar(100) NULL,
  precision integer NULL,
  token_symbol varchar(255) NULL,
  token_name varchar(255) NULL,
  logo varchar(255) NULL,
  url varchar(255) NULL,
  description text NULL,
  CONSTRAINT "pk_token_contract" PRIMARY KEY (id)
);

CREATE INDEX token_contract_contract_id_idx ON token_contract (contract_id);
CREATE INDEX token_contract_contract_owner_addr_idx ON token_contract (owner_addr);
CREATE INDEX token_contract_block_num_idx ON token_contract (block_num);

CREATE TABLE "token_balance" (
  id serial NOT NULL,
  contract_addr varchar(100) NOT NULL,
  owner_addr varchar(100) NOT NULL,
  amount decimal(36, 18) NOT NULL,
  created_at bigint NOT NULL,
  updated_at bigint NOT NULL,
  CONSTRAINT "pk_token_balance" PRIMARY KEY (id)
);

CREATE INDEX token_balance_contract_addr_idx ON token_balance (contract_addr);
CREATE INDEX token_balance_contract_addr_and_owner_addr_idx ON token_balance (contract_addr, owner_addr);

CREATE TABLE "token_contract_transfer_history" (
  id serial NOT NULL,
  contract_addr varchar(100) NOT NULL,
  from_addr varchar(100) NOT NULL,
  to_addr varchar(100) NOT NULL,
  amount decimal(36, 18) NOT NULL,
  block_num integer NOT NULL,
  txid varchar(100) NOT NULL,
  op_num integer NOT NULL,
  event_name varchar(100) NOT NULL,
  tx_time bigint NOT NULL,
  created_at bigint NOT NULL,
  updated_at bigint NOT NULL,
  CONSTRAINT "pk_token_contract_transfer_history" PRIMARY KEY (id)
);

CREATE TABLE "asset" (
  asset_id varchar(10) NOT NULL,
  symbol varchar(20) NOT NULL,
  precision integer NOT NULL,
  created_at bigint NOT NULL,
  updated_at bigint NOT NULL,
  CONSTRAINT "pk_asset" PRIMARY KEY (asset_id)
);

CREATE TABLE "address_balance" (
  id serial NOT NULL,
  owner_addr varchar(100) NOT NULL,
  asset_id varchar(10) NOT NULL,
  amount decimal(36, 18) NOT NULL,
  created_at bigint NOT NULL,
  updated_at bigint NOT NULL,
  CONSTRAINT "pk_address_balance" PRIMARY KEY (id)
);

CREATE TABLE "account" (
  id serial NOT NULL,
  owner_addr varchar(100) NOT NULL,
  account_name varchar(100) NOT NULL,
  created_at bigint NOT NULL,
  updated_at bigint NOT NULL,
  CONSTRAINT "pk_account" PRIMARY KEY (id)
);

CREATE INDEX token_contract_transfer_history_contract_addr_idx ON token_contract_transfer_history (contract_addr);

CREATE INDEX token_contract_transfer_history_txid_op_num_idx ON token_contract_transfer_history (txid, op_num);

CREATE INDEX token_contract_transfer_history_from_addr_and_to_addr_idx ON token_contract_transfer_history (from_addr, to_addr);

CREATE INDEX token_contract_transfer_history_contract_addr_and_from_addr_and_to_addr_idx ON token_contract_transfer_history (contract_addr, from_addr, to_addr);

CREATE INDEX address_balance_owner_addr_idx ON address_balance (owner_addr);
CREATE INDEX address_balance_asset_id_idx ON address_balance (asset_id);
CREATE INDEX address_balance_owner_addr_and_asset_id_idx ON address_balance (owner_addr, asset_id);

CREATE INDEX account_owner_addr_idx ON account (owner_addr);
CREATE INDEX account_account_name_idx ON account (account_name);
