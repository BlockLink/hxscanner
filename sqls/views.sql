
create or replace view citizen_fee_change as
(select r.addr, r.block_num, r.trxid, r.account, (r.new_options->>'miner_pledge_pay_back')::int as miner_pledge_pay_back,
 r.new_options from (select r.addr, r.block_num, r.trxid, r.new_options::jsonb, r.account from tbl_account_update_operation r order by r.block_num desc  ) r);

select cfc.addr, aco.name as account_name, max(miner_pledge_pay_back) as max_miner_pledge_pay_back, count(cfc.addr) as change_count, max(cfc.block_num) as max_block_num, max(cfc.account) as account from citizen_fee_change cfc left join tbl_account_create_operation aco on aco.payer=cfc.addr where miner_pledge_pay_back >10 group by cfc.addr, aco.name order by max_block_num desc;

select o.*, (o.amount::jsonb ->'amount')::bigint as transfer_amount, (o.amount::jsonb ->> 'asset_id')::text as transfer_asset_id from tbl_transfer_operation o limit 10;

create or replace view view_transfer_operation as (select o.*, (o.amount::jsonb ->>'amount')::bigint as transfer_amount, (o.amount::jsonb ->> 'asset_id')::text as transfer_asset_id from tbl_transfer_operation o);

select o.*, (o.amount::jsonb ->>'amount')::bigint as transfer_amount, (o.amount::jsonb ->> 'asset_id')::text as transfer_asset_id from tbl_transfer_operation o limit 100;
