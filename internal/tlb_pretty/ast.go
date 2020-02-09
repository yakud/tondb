package tlb_pretty

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
)

const NOTHING = "nothing"

type AstNode struct {
	Parent    *AstNode               `json:"-"`
	Fields    map[string]interface{} `json:"data"`
	Level     int
	ParentKey string
	EmptyKeys int
}

func (t *AstNode) ToJSON() ([]byte, error) {
	data := t.PureFields()
	return json.Marshal(data)
}

// ConvertToBlock AST node to map
func (t *AstNode) PureFields() map[string]interface{} {
	var fields = make(map[string]interface{})
	for k, v := range t.Fields {
		if vv, ok := v.(*AstNode); ok {
			fields[k] = vv.PureFields()
		} else if vv, ok := v.([]*AstNode); ok {
			newSlice := make([]map[string]interface{}, 0, len(vv))
			for _, sliceV := range vv {
				newSlice = append(newSlice, sliceV.PureFields())
			}

			fields[k] = newSlice
		} else {
			fields[k] = v
		}
	}

	return fields
}

func (t *AstNode) Get(key ...string) (interface{}, error) {
	if len(key) <= 0 {
		return nil, errors.New("empty key")
	}

	if len(key) == 1 {
		v, ok := t.Fields[key[0]]
		if !ok {
			return nil, fmt.Errorf("key %s not found", key[0])
		}

		return v, nil
	}

	child, err := t.GetNode(key[0])
	if err != nil {
		return nil, fmt.Errorf("get key %+v error: %+v", key, err)
	}

	v, err := child.Get(key[1:]...)
	if err != nil {
		return nil, fmt.Errorf("get key %+v error: %+v", key, err)
	}

	return v, err
}

func (t *AstNode) GetNode(key ...string) (*AstNode, error) {
	if v, err := t.Get(key...); err != nil {
		return nil, err
	} else {
		if vv, ok := v.(*AstNode); ok {
			return vv, nil
		}

		return nil, fmt.Errorf("value under key %s is not *AstNode", key)
	}
}

func (t *AstNode) GetString(key ...string) (string, error) {
	if v, err := t.Get(key...); err != nil {
		return "", err
	} else {
		if vv, ok := v.(string); ok {
			return vv, nil
		}

		return "", fmt.Errorf("value under key %s is not string", key)
	}
}

func (t *AstNode) GetSlice(key ...string) ([]*AstNode, error) {
	if v, err := t.Get(key...); err != nil {
		return nil, err
	} else {
		if vv, ok := v.([]*AstNode); ok {
			return vv, nil
		}

		return nil, fmt.Errorf("value under key %s is not []*AstNode", key)
	}
}

func (t *AstNode) EachNode(each func(i int, el *AstNode) error, key ...string) error {
	if vSlice, err := t.GetSlice(key...); err != nil {
		return err
	} else {
		for i, v := range vSlice {
			if err := each(i, v); err != nil {
				return err
			}
		}

		return nil
	}
}

func (t *AstNode) GetBool(key ...string) (bool, error) {
	if v, err := t.GetUint8(key...); err != nil {
		return false, err
	} else {
		return v == 1, nil
	}
}

func (t *AstNode) IsType(checkType string) bool {
	if nodeType, err := t.Type(); err == nil && nodeType == checkType {
		return true
	}
	return false
}

func (t *AstNode) Type() (string, error) {
	return t.GetString("@type")
}

func (t *AstNode) GetUint(bitSize int, key ...string) (uint64, error) {
	if v, err := t.GetString(key...); err != nil {
		return 0, err
	} else {
		vUint, err := strconv.ParseUint(v, 10, bitSize)
		if err != nil {
			return 0, fmt.Errorf("error parse value by key %s to uint%d: %+v", key, bitSize, err)
		}

		return vUint, nil
	}
}

func (t *AstNode) GetInt(bitSize int, key ...string) (int64, error) {
	if v, err := t.GetString(key...); err != nil {
		return 0, err
	} else {
		vInt, err := strconv.ParseInt(v, 10, bitSize)
		if err != nil {
			return 0, fmt.Errorf("error parse value by key %s to int%d: %+v", key, bitSize, err)
		}

		return vInt, nil
	}
}

func (t *AstNode) GetUint8(key ...string) (uint8, error) {
	if v, err := t.GetUint(8, key...); err != nil {
		return 0, err
	} else {
		return uint8(v), nil
	}
}

func (t *AstNode) GetUint16(key ...string) (uint16, error) {
	if v, err := t.GetUint(16, key...); err != nil {
		return 0, err
	} else {
		return uint16(v), nil
	}
}

func (t *AstNode) GetUint32(key ...string) (uint32, error) {
	if v, err := t.GetUint(32, key...); err != nil {
		return 0, err
	} else {
		return uint32(v), nil
	}
}
func (t *AstNode) GetInt32(key ...string) (int32, error) {
	if v, err := t.GetInt(32, key...); err != nil {
		return 0, err
	} else {
		return int32(v), nil
	}
}

func (t *AstNode) GetValueOrNothingInt32(key ...string) (value int32, err error) {
	key = append(key, "value")
	if value, err = t.GetInt32(key...); err != nil {
		if resultArgStr, err := t.GetString(key[0]); err != nil || resultArgStr != NOTHING {
			return 0, err
		}
	}

	return value, nil
}

func (t *AstNode) GetValueOrNothingUint64(key ...string) (value uint64, err error) {
	key = append(key, "value")
	if value, err = t.GetUint64(key...); err != nil {
		if resultArgStr, err := t.GetString(key[0]); err != nil || resultArgStr != NOTHING {
			return 0, err
		}
	}

	return value, nil
}

func (t *AstNode) GetInt8(key ...string) (int8, error) {
	if v, err := t.GetInt(8, key...); err != nil {
		return 0, err
	} else {
		return int8(v), nil
	}
}

func (t *AstNode) GetUint64(key ...string) (uint64, error) {
	if v, err := t.GetUint(64, key...); err != nil {
		return 0, err
	} else {
		return v, nil
	}
}
