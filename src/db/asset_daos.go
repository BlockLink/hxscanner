package db

import (
	"time"
	"github.com/shopspring/decimal"
)

func SaveAsset(asset *AssetEntity) error {
	now := time.Now()
	stmt, err := dbConn.Prepare("INSERT INTO public.asset (asset_id, symbol," +
		" precision, created_at, updated_at)" +
		" VALUES (($1),($2),($3),($4),($5))")
	if err != nil {
		return err
	}
	defer stmt.Close()
	res, err := stmt.Exec(asset.AssetId, asset.Symbol, asset.Precision, now.Unix(), now.Unix())
	if err != nil {
		return err
	}
	_ = res
	return nil
}

func FindAsset(assetId string) (result *AssetEntity, err error) {
	rows, err := dbConn.Query("SELECT asset_id, symbol, precision, created_at, updated_at" +
		" FROM public.asset where asset_id=$1", assetId)
	if err != nil {
		return
	}
	defer rows.Close()
	if rows.Next() {
		result = new(AssetEntity)
		var createdAtUnix, updatedAtUnix int64
		err = rows.Scan(&result.AssetId, &result.Symbol, &result.Precision,
			&createdAtUnix, &updatedAtUnix)
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

func SaveAddressBalance(addressBalance *AddressBalanceEntity) error {
	now := time.Now()
	stmt, err := dbConn.Prepare("INSERT INTO public.address_balance (owner_addr, asset_id, " +
		" amount, created_at, updated_at)" +
		" VALUES (($1),($2),($3),($4),($5))")
	if err != nil {
		return err
	}
	defer stmt.Close()
	res, err := stmt.Exec(addressBalance.OwnerAddr, addressBalance.AssetId, addressBalance.Amount.String(), now.Unix(), now.Unix())
	if err != nil {
		return err
	}
	_ = res
	return nil
}


func UpdateAddressBalance(addressBalance *AddressBalanceEntity) error {
	stmt, err := dbConn.Prepare("UPDATE public.address_balance SET owner_addr = $1, asset_id=$2, amount = $3," +
		"created_at = $4, updated_at = $5 WHERE id=$6")
	if err != nil {
		return err
	}
	defer stmt.Close()
	res, err := stmt.Exec(addressBalance.OwnerAddr, addressBalance.AssetId, addressBalance.Amount.String(), addressBalance.CreatedAt.Unix(), addressBalance.UpdatedAt.Unix(), addressBalance.Id)
	if err != nil {
		return err
	}
	_ = res
	return nil
}

func FindAddressBalanceByOwnerAddrAndAssetId(ownerAddr string, assetId string) (result *AddressBalanceEntity, err error) {
	rows, err := dbConn.Query("SELECT id, owner_addr, asset_id, amount, created_at, updated_at FROM public.address_balance where owner_addr=$1 and asset_id=$2", ownerAddr, assetId)
	if err != nil {
		return
	}
	defer rows.Close()
	if rows.Next() {
		result = new(AddressBalanceEntity)
		var amountStr string
		var createdAtUnix, updatedAtUnix int64
		err = rows.Scan(&result.Id, &result.OwnerAddr, &result.AssetId, &amountStr,
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