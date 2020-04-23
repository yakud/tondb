package streaming

import (
	"errors"
	"github.com/google/btree"
	"gitlab.flora.loc/mills/tondb/internal/ton/view/feed"
	"strconv"
	"strings"
)

const (
	FeedTransactions = "transactions"
	FeedMessages     = "messages"
)

var MessagesFieldsIndices     = [...]string{"lt", "time", "message_lt", "value_nanogram", "total_fee_nanogram"}
var TransactionsFieldsIndices = [...]string{"lt", "time", "msg_in_created_lt", "total_nanograms", "total_fees_nanograms",
											"total_fwd_fee_nanograms", "total_ihr_fee_nanograms", "total_import_fee_nanograms"}

type TransactionsAndMessagesIndexer struct {
	addrIndex map[string]map[string]struct{}
	indices   map[string]*btree.BTree
}

func (fi *TransactionsAndMessagesIndexer) Add(value interface{}, valueJson []byte) error {
	switch value.(type) {
	case *feed.MessageInFeed:
		fi.AddMessage(value.(*feed.MessageInFeed), valueJson)
		return nil
	case *feed.TransactionInFeed:
		fi.AddTransaction(value.(*feed.TransactionInFeed), valueJson)
		return nil
	default:
		return errors.New("couldn't add value to indices because it's of wrong type")
	}
}

func (fi *TransactionsAndMessagesIndexer) AddMessage(msg *feed.MessageInFeed, msgJson []byte)  {
	fi.addAddrToMap(FeedMessages, msg.WorkchainId, msg.Src, msgJson)
	fi.addAddrToMap(FeedMessages, msg.WorkchainId, msg.Dest, msgJson)

	fi.InsertIndex(FeedMessages+"_"+"lt", NewUInt64Index(msg.Lt, msgJson))
	fi.InsertIndex(FeedMessages+"_"+"time", NewUInt64Index(msg.Time, msgJson))
	fi.InsertIndex(FeedMessages+"_"+"message_lt", NewUInt64Index(msg.MessageLt, msgJson))
	fi.InsertIndex(FeedMessages+"_"+"value_nanogram", NewUInt64Index(msg.ValueNanogram, msgJson))
	fi.InsertIndex(FeedMessages+"_"+"total_fee_nanogram", NewUInt64Index(msg.TotalFeeNanogram, msgJson))
}

func (fi *TransactionsAndMessagesIndexer) AddTransaction(trx *feed.TransactionInFeed, trxJson []byte) {
	fi.addAddrToMap(FeedTransactions, trx.WorkchainId, trx.AccountAddr, trxJson)

	fi.InsertIndex(FeedTransactions+"_"+"lt", NewUInt64Index(trx.Lt, trxJson))
	fi.InsertIndex(FeedTransactions+"_"+"time", NewUInt64Index(trx.TimeUnix, trxJson))
	fi.InsertIndex(FeedTransactions+"_"+"msg_in_created_lt", NewUInt64Index(trx.MsgInCreatedLt, trxJson))
	fi.InsertIndex(FeedTransactions+"_"+"total_nanograms", NewUInt64Index(trx.TotalNanograms, trxJson))
	fi.InsertIndex(FeedTransactions+"_"+"total_fees_nanograms", NewUInt64Index(trx.TotalFeesNanograms, trxJson))
	fi.InsertIndex(FeedTransactions+"_"+"total_fwd_fee_nanograms", NewUInt64Index(trx.TotalFwdFeeNanograms, trxJson))
	fi.InsertIndex(FeedTransactions+"_"+"total_ihr_fee_nanograms", NewUInt64Index(trx.TotalIhrFeeNanograms, trxJson))
	fi.InsertIndex(FeedTransactions+"_"+"total_import_fee_nanograms", NewUInt64Index(trx.TotalImportFeeNanograms, trxJson))
}

func (fi *TransactionsAndMessagesIndexer) Init() {
	fi.addrIndex = make(map[string]map[string]struct{})
	fi.indices = make(map[string]*btree.BTree)

	for _, field := range MessagesFieldsIndices {
		fi.indices[FeedMessages+"_"+field] = btree.New(2)
	}

	for _, field := range TransactionsFieldsIndices {
		fi.indices[FeedTransactions+"_"+field] = btree.New(2)
	}
}

func (fi *TransactionsAndMessagesIndexer) Filter(filter Filter) []byte {
	customFilters := ParseCustomFilters(filter.customFilters)

	if filter.AccountAddr != nil {
		if jsons, ok := fi.addrIndex[filter.FeedName+"_"+*filter.AccountAddr]; ok {
			return fi.filterWithIntersection(filter.FeedName, customFilters, jsons)
		} else {
			return []byte{}
		}
	}

	return fi.filterWithIntersection(filter.FeedName, customFilters, nil)
}

func (fi *TransactionsAndMessagesIndexer) filterWithIntersection(feed string, filters []CustomFilter, intersection map[string]struct{}) []byte {
	if intersection != nil && len(intersection) == 0 {
		return []byte{}
	}

	iter := NewIndexIterator()
	if len(filters) > 0	 {
		filter := filters[0]
		if tree, ok := fi.indices[feed+"_"+filter.Field]; ok {
			switch filter.Operation {
			case EqOp:
				if v, err := strconv.ParseUint(filter.ValueString, 10, 64); err == nil {

					if raw := tree.Get(UInt64Index{value: v}); raw != nil {

						if len(filters) == 1 {
							// it is last filter so final result is intersection of current result and previous intersection
							return fi.setToBytes(fi.intersect(intersection, raw.(IndexItem).GetJsons()))
						} else {
							// it is not the last filter so we need to go deeper in recursion, passing filters without
							// current filter and with intersection of previous step intersection with this step result
							return fi.filterWithIntersection(feed, filters[1:], fi.intersect(intersection, raw.(IndexItem).GetJsons()))
						}
					} else {
						return []byte{}
					}
				} else {
					return []byte{}
				}
			case LtOp:
				if v, err := strconv.ParseUint(filter.ValueString, 10, 64); err == nil {

					tree.AscendLessThan(UInt64Index{value: v}, iter.Iterate)

					if len(filters) == 1 {
						return fi.setToBytes(fi.intersect(intersection, iter.GetJsons()))
					} else {
						return fi.filterWithIntersection(feed, filters[1:], fi.intersect(intersection, iter.GetJsons()))
					}
				} else {
					return []byte{}
				}
			case GtOp:
				if v, err := strconv.ParseUint(filter.ValueString, 10, 64); err == nil {

					tree.DescendGreaterThan(UInt64Index{value: v}, iter.Iterate)

					if len(filters) == 1 {
						return fi.setToBytes(fi.intersect(intersection, iter.GetJsons()))
					} else {
						return fi.filterWithIntersection(feed, filters[1:], fi.intersect(intersection, iter.GetJsons()))
					}
				} else {
					return []byte{}
				}
			case RangeOp:
				if first, second, err := filter.ParseRange(); err == nil {

					tree.DescendRange(UInt64Index{value: first}, UInt64Index{value: second}, iter.Iterate)

					if len(filters) == 1 {
						return fi.setToBytes(fi.intersect(intersection, iter.GetJsons()))
					} else {
						return fi.filterWithIntersection(feed, filters[1:], fi.intersect(intersection, iter.GetJsons()))
					}
				} else {
					return []byte{}
				}
			default:
				return []byte{}
			}
		} else {
			return []byte{}
		}
	}

	return fi.setToBytes(intersection)
}

func (fi *TransactionsAndMessagesIndexer) intersect(first, second map[string]struct{}) map[string]struct{} {
	// if we get nil as any parameter, we don't intersect sets and just return other parameter
	if first == nil {
		return second
	} else if second == nil {
		return first
	}

	// iterating over smaller set
	if len(first) > len(second) {
		res := make(map[string]struct{}, len(second))
		for key := range second {
			if _, ok := first[key]; ok {
				res[key] = struct{}{}
			}
		}
	} else {
		res := make(map[string]struct{}, len(first))
		for key := range first {
			if _, ok := second[key]; ok {
				res[key] = struct{}{}
			}
		}
	}

	// code never goes here
	return nil
}

func (fi *TransactionsAndMessagesIndexer) setToBytes(set map[string]struct{}) []byte {
	if len(set) == 0 {
		return []byte{}
	}

	res := make([]byte, 0, len(set)*64)
	res = append(res, '[')

	for key := range set {
		// check if key is json array (but it shouldn't be so, so maybe this check is redundant)
		if strings.HasPrefix(key, "[") && strings.HasSuffix(key, "]") {
			// if res has no items yet (just [) we dont need to separate array entries
			if len(res) > 1 {
				res = append(res, ',')
			}
			res = append(res, key[1:len(key)-2]...)
		} else {
			if len(res) > 1 {
				res = append(res, ',')
			}
			res = append(res, key...)
		}
	}

	res = append(res, ']')

	return res
}

func (fi *TransactionsAndMessagesIndexer) addJsonToAddrMap(key string, json []byte) {
	if v, ok := fi.addrIndex[key]; ok {
		v[string(json)] = struct{}{}
	} else {
		v = map[string]struct{}{string(json):{}}
	}
}

func (fi *TransactionsAndMessagesIndexer) addAddrToMap(feed string, wcId int32, addr string, json []byte) {
	if len(addr) > 1 {
		key := feed+"_"+strconv.FormatInt(int64(wcId), 10)+":"+addr
		fi.addJsonToAddrMap(key, json)
	}
}

func (fi *TransactionsAndMessagesIndexer) InsertIndex(key string, item IndexItem) {
	// ReplaceOrInsert returns nil if item is inserted and Item if it was replaced.
	// We want not to replace items, but to add data to their jsons
	if rawIndex := fi.indices[key].ReplaceOrInsert(item); rawIndex != nil {
		index := rawIndex.(IndexItem)
		index.JsonsUnion(item.GetJsons())
		fi.indices[key].ReplaceOrInsert(index)
	}
}

func NewTransactionsAndMessagesIndexer() *TransactionsAndMessagesIndexer {
	indexer := &TransactionsAndMessagesIndexer{}
	indexer.Init()
	return indexer
}
