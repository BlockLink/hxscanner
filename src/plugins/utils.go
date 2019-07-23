package plugins

import (
	"encoding/json"
	"github.com/blocklink/hxscanner/src/log"
)

var logger = log.GetLogger()

func isStringInArray(item string, arr []string) bool {
	for _, val := range arr {
		if val == item {
			return true
		}
	}
	return false
}

func isAllInArray(items []string, arr []string) bool {
	for _, item := range items {
		if !isStringInArray(item, arr) {
			return false
		}
	}
	return true
}

func objArrayToStringArray(src []interface{}) (result []string, ok bool) {
	for _, item := range src {
		itemStr, isStr := item.(string)
		if !isStr {
			ok = false
			return
		}
		result = append(result, itemStr)
	}
	ok = true
	return
}

func objToStringArray(src interface{}) (result []string, ok bool) {
	objArray, ok := src.([]interface{})
	if !ok {
		return
	}
	result, ok = objArrayToStringArray(objArray)
	return
}

func getStringPropFromJSONObj(jsonObj map[string]interface{}, prop string) (result string, ok bool) {
	item, ok := jsonObj[prop]
	if !ok {
		return
	}
	result, ok = item.(string)
	return
}

func getIntPropFromJSONObj(jsonObj map[string]interface{}, prop string) (result int64, isInt bool) {
	itemObj, ok := jsonObj[prop]
	if !ok {
		return
	}
	if item, ok := itemObj.(int64); ok {
		result = item
		isInt = true
		return
	}
	if item, ok := itemObj.(int32); ok {
		result = int64(item)
		isInt = true
		return
	}
	if item, ok := itemObj.(json.Number); ok {
		itemInt, err := item.Int64()
		if err != nil {
			isInt = false
			return
		}
		result = itemInt
		isInt = true
		return
	}
	isInt = false
	return
}

