package feed

import (
	"encoding/base64"
	"encoding/json"
)

func PackScrollId(scrollId interface{}) (string, error) {
	scrollJson, err := json.Marshal(scrollId)
	if err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(scrollJson), nil
}

func UnpackScrollId(packedScrollId string, scrollId interface{}) error {
	scrollJson, err := base64.RawURLEncoding.DecodeString(packedScrollId)
	if err != nil {
		return err
	}
	return json.Unmarshal(scrollJson, scrollId)
}
