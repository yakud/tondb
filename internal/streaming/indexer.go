package streaming

import (
	"github.com/google/btree"
)

type Indexer interface {
	Add(value interface{}, valueJson []byte) error
	Filter(filter Filter) []byte
	Init()
}

type IndexItem interface {
	Less(item btree.Item) bool
	GetJsons() map[string]struct{}
	JsonsUnion(jsons map[string]struct{})
}

type UInt64Index struct {
	value uint64
	jsons map[string]struct{}
}

func (i UInt64Index) Less(item btree.Item) bool {
	return i.value < item.(UInt64Index).value
}

func (i UInt64Index) GetJsons() map[string]struct{} {
	return i.jsons
}

func (i UInt64Index) JsonsUnion(jsons map[string]struct{}) {
	for key := range jsons {
		i.jsons[key] = struct{}{}
	}
}

func NewUInt64Index(value uint64, json []byte) UInt64Index {
	return UInt64Index{
		value: value,
		jsons: map[string]struct{}{string(json):{}},
	}
}

type IndexIterator struct {
	jsons map[string]struct{}
}

func (it *IndexIterator) Iterate(itemRaw btree.Item) bool {
	item := itemRaw.(IndexItem)
	for key := range item.GetJsons() {
		it.jsons[key] = struct{}{}
	}

	return true
}

func (it *IndexIterator) GetJsons() map[string]struct{} {
	return it.jsons
}

func NewIndexIterator() *IndexIterator {
	return &IndexIterator{
		jsons: make(map[string]struct{}),
	}
}

