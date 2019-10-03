package tlb_pretty

import (
	"bytes"
	"strconv"
)

// Parser TLB pretty schema to AST tree
type Parser struct {
}

const (
	OpenBrace   = byte('(')
	CloseBrace  = byte(')')
	Space       = byte(' ')
	NewLine     = byte('\n')
	KVDelimiter = byte(':')
	Up          = byte('^')
)

var (
	RawType   = []byte("raw@")
	FieldType = "@type"
	EmptyType = "empty"
)

const (
	StateInit = iota
	StateLookingType
	StateFinishingLookingType
	StateLookingKey
	StateLookingValue
	StateFinishingLookingValue
	StateFinishingLookingNode
)

func (t *Parser) Parse(data []byte) *AstNode {
	var node *AstNode
	var state = StateInit
	var key = make([]byte, 0)
	var chunk = make([]byte, 0)
	var braceCount = 0
	var typeBraceCount = 0
	var isRawType = false

	for _, b := range data {
		if b == Up {
			continue
		}

		switch state {
		case StateInit:
			switch b {
			case OpenBrace:
				node = &AstNode{
					Fields: make(map[string]interface{}),
					Level:  braceCount,
				}
				braceCount++
				state = StateLookingType
			}

		case StateLookingType:
			switch b {
			case OpenBrace:
				if bytes.HasPrefix(chunk, RawType) {
					chunk = append(chunk, b)
					typeBraceCount++
				}
			case Space:
				if typeBraceCount > 0 {
					chunk = append(chunk, b)
				} else {
					state = StateFinishingLookingType
				}
			case NewLine:
				state = StateFinishingLookingType
			case CloseBrace:
				if bytes.HasPrefix(chunk, RawType) {
					chunk = append(chunk, b)
					typeBraceCount--
				}
			case KVDelimiter:
				node.Fields[FieldType] = EmptyType
				key = make([]byte, len(chunk))
				copy(key, chunk)
				chunk = chunk[:0]
				state = StateLookingValue
			default:
				chunk = append(chunk, b)
			}

			if state == StateFinishingLookingType {
				node.Fields[FieldType] = string(chunk)

				if bytes.HasPrefix(chunk, RawType) {
					isRawType = true
					key = []byte("value")
					state = StateLookingValue
				} else {
					state = StateLookingKey
				}

				chunk = chunk[:0]
			}

		case StateLookingKey:
			switch b {
			case Space:
				continue
			case NewLine:
				continue
			case OpenBrace:
				if len(chunk) == 0 {
					key = []byte("value_" + strconv.Itoa(node.EmptyKeys))
					node.EmptyKeys++
					newNode := &AstNode{
						Parent:    node,
						Fields:    make(map[string]interface{}),
						Level:     braceCount,
						ParentKey: string(key),
					}
					node = newNode
					braceCount++
					state = StateLookingType
				} else {
					state = StateLookingValue
				}
			case CloseBrace:
				braceCount--
				state = StateLookingKey
				isRawType = false
				if node != nil && node.Parent != nil {
					node.Parent.Fields[node.ParentKey] = node
					node = node.Parent
				}

			case KVDelimiter:
				key = make([]byte, len(chunk))
				copy(key, chunk)
				chunk = chunk[:0]
				state = StateLookingValue
			default:
				chunk = append(chunk, b)
			}

		case StateLookingValue:
			switch b {
			case OpenBrace:
				newNode := &AstNode{
					Parent:    node,
					Fields:    make(map[string]interface{}),
					Level:     braceCount,
					ParentKey: string(key),
				}
				node = newNode
				braceCount++
				state = StateLookingType

			case CloseBrace:
				braceCount--
				state = StateFinishingLookingNode

			case Space:
				if isRawType {
					if len(chunk) > 0 {
						chunk = append(chunk, b)
					}
				} else {
					state = StateFinishingLookingValue
				}

			case NewLine:
				if isRawType {
					if len(chunk) > 0 {
						chunk = append(chunk, b)
					}
				} else {
					state = StateFinishingLookingValue
				}

			default:
				chunk = append(chunk, b)
			}

			if state == StateFinishingLookingValue || state == StateFinishingLookingNode {
				node.Fields[string(key)] = string(bytes.TrimSpace(chunk))
				chunk = chunk[:0]
				key = key[:0]
				isRawType = false

				if state == StateFinishingLookingNode {
					if node != nil && node.Parent != nil {
						node.Parent.Fields[node.ParentKey] = node
						node = node.Parent
					}
				}
				state = StateLookingKey
			}
		}
	}

	return node
}

func NewParser() *Parser {
	return &Parser{}
}
