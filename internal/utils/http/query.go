package http

import (
	"fmt"
	"net/url"
	"strconv"
)

func GetQueryValueString(url *url.URL, key string) (string, error) {
	values, ok := url.Query()[key]
	if !ok || len(values) > 1 {
		return "", fmt.Errorf("should be set exactly one %s field", key)
	}

	return values[0], nil
}

func GetQueryValueInt(url *url.URL, key string, bitsize int) (int64, error) {
	valStr, err := GetQueryValueString(url, key)
	if err != nil {
		return 0, err
	}

	valInt, err := strconv.ParseInt(valStr, 10, bitsize)
	if err != nil {
		return 0, err
	}

	return valInt, nil
}

func GetQueryValueUint(url *url.URL, key string, bitsize int) (uint64, error) {
	valStr, err := GetQueryValueString(url, key)
	if err != nil {
		return 0, err
	}

	valInt, err := strconv.ParseUint(valStr, 10, bitsize)
	if err != nil {
		return 0, err
	}

	return valInt, nil
}

func GetQueryValueInt16(url *url.URL, key string) (int16, error) {
	val64, err := GetQueryValueInt(url, key, 16)
	if err != nil {
		return 0, err
	}

	return int16(val64), nil
}

func GetQueryValueInt32(url *url.URL, key string) (int32, error) {
	val64, err := GetQueryValueInt(url, key, 32)
	if err != nil {
		return 0, err
	}

	return int32(val64), nil
}

func GetQueryValueUint16(url *url.URL, key string) (uint16, error) {
	val64, err := GetQueryValueUint(url, key, 16)
	if err != nil {
		return 0, err
	}

	return uint16(val64), nil
}

func GetQueryValueUint32(url *url.URL, key string) (uint32, error) {
	val64, err := GetQueryValueUint(url, key, 32)
	if err != nil {
		return 0, err
	}

	return uint32(val64), nil
}

func GetQueryValueArrString(url *url.URL, key string) ([]string, error) {
	values, ok := url.Query()[key]
	if !ok || len(values) == 0 {
		return nil, fmt.Errorf("should be set at least one %s field", key)
	}

	return values, nil
}
