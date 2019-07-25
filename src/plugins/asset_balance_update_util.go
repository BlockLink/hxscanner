package plugins

import (
	"github.com/blocklink/hxscanner/src/nodeservice"
	"github.com/blocklink/hxscanner/src/db"
	"time"
	"github.com/shopspring/decimal"
	"math"
	"errors"
)

var assetsCache = make(map[string]*db.AssetEntity) // assetId => assetInfo

func findAssetFromCacheOrDb(assetId string) (result *db.AssetEntity, err error) {
	item, ok := assetsCache[assetId]
	if ok {
		result = item
		return
	}
	result, err = db.FindAsset(assetId)
	if err != nil {
		return
	}
	if result == nil {
		// 没找到就 list_assets 0, 100获取资产列表
		var nodeAssets []*db.AssetEntity
		nodeAssets, err = nodeservice.ListAssets(0, 100)
		if err != nil {
			return
		}
		for _, item := range nodeAssets {
			var existItem *db.AssetEntity
			existItem, err = db.FindAsset(item.AssetId)
			if err != nil {
				return
			}
			if existItem == nil {
				err = db.SaveAsset(item)
				if err != nil {
					return
				}
			}
			if item.AssetId == assetId {
				result = item
			}
		}
	}
	return
}

func updateAddressBalance(addr string, assetId string) (err error) {
	addrBalances, err := nodeservice.GetAddressBalances(addr)
	if err != nil {
		return
	}
	newBalance, ok := addrBalances[assetId]
	if !ok {
		newBalance = 0
	}
	assetEntry, err := findAssetFromCacheOrDb(assetId)
	if err != nil {
		return
	}
	if assetEntry == nil {
		err = errors.New("unknown asset id " + assetId)
		return
	}
	precisonFull := int64(math.Pow10(int(assetEntry.Precision)))
	newBalanceBn := decimal.New(int64(newBalance), 0).Div(decimal.New(precisonFull, 0))
	// update addr balance to db
	record, err := db.FindAddressBalanceByOwnerAddrAndAssetId(addr, assetId)
	if err != nil {
		return
	}
	now := time.Now()
	if record == nil {
		record = &db.AddressBalanceEntity{OwnerAddr:addr, AssetId:assetId, Amount: newBalanceBn, CreatedAt:now, UpdatedAt:now}
		err = db.SaveAddressBalance(record)
		if err != nil {
			return
		}
	} else {
		record.Amount = newBalanceBn
		record.UpdatedAt = now
		err = db.UpdateAddressBalance(record)
		if err != nil {
			return
		}
	}
	return
}
