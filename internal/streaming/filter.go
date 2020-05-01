package streaming

import (
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"gitlab.flora.loc/mills/tondb/internal/ton/view/feed"
)

var rangeRegexp = regexp.MustCompile(`(?m)\[(\d+), *(\d+)\]`)

const (
	FeedNameBlocks       FeedName = "blocks"
	FeedNameTransactions FeedName = "transactions"
	FeedNameMessages     FeedName = "messages"

	FieldTrxTotalNanogam CustomFilterField = "total_nanograms"
	FieldMsgValueNanogam CustomFilterField = "value_nanogram"

	OpEq    CustomFilterOperation = "eq"
	OpLt    CustomFilterOperation = "lt"
	OpGt    CustomFilterOperation = "gt"
	OpRange CustomFilterOperation = "range"

	MessageDirectionIn  MessageDirection = "in"
	MessageDirectionOut MessageDirection = "out"
)

type (
	FeedName         string
	Addr             string
	FilterHash       string
	MessageDirection string

	CustomFilterField     string
	CustomFilterOperation string

	CustomFilters []CustomFilter

	Filter struct {
		FeedName    FeedName `json:"feed_name"`
		WorkchainId *int32   `json:"workchain_id,omitempty"`
		Shard       *uint64  `json:"shard,omitempty"`
		AccountAddr *string  `json:"account_addr,omitempty"`

		MessageDirection *MessageDirection `json:"message_direction,omitempty"`

		CustomFilters CustomFilters `json:"custom_filters,omitempty"`
	}

	CustomFilter struct {
		Field       CustomFilterField     `json:"field"`
		Operation   CustomFilterOperation `json:"operation"`
		ValueString string                `json:"value_string"`
	}
)

func (f *Filter) Hash() FilterHash {
	sb := strings.Builder{}

	sb.WriteString("feed_name=" + string(f.FeedName) + ",")

	if f.WorkchainId != nil {
		sb.WriteString(fmt.Sprintf("workchain_id=%d,", *f.WorkchainId))
	}

	if f.Shard != nil {
		sb.WriteString(fmt.Sprintf("shard=%d,", *f.Shard))
	}

	if f.AccountAddr != nil && len(*f.AccountAddr) != 0 {
		sb.WriteString("account_addr=%s," + *f.AccountAddr + ",")
	}

	if f.MessageDirection != nil {
		sb.WriteString("message_direction=" + string(*f.MessageDirection) + ",")
	}

	f.CustomFilters.Sort()
	for i, filter := range f.CustomFilters {
		if i == 0 {
			sb.WriteString("custom_filters=")
		}

		sb.WriteString(string(filter.Field) + " " + string(filter.Operation) + " " + filter.ValueString)

		if i < len(f.CustomFilters)-1 {
			sb.WriteString(";")
		}
	}
	return FilterHash(sb.String())
}

func (f *Filter) MatchWorkchainAndShard(block *feed.BlockInFeed) bool {
	return (f.WorkchainId == nil || (f.WorkchainId != nil && block.WorkchainId == *f.WorkchainId)) &&
		(f.Shard == nil || (f.Shard != nil && block.Shard == *f.Shard))
}

func (cf *CustomFilter) ParseRange() (first, second uint64, err error) {
	if cf.Operation != OpRange {
		return 0, 0, errors.New("cant parse range, invalid operation")
	}

	rng := rangeRegexp.FindAllString(cf.ValueString, -1)
	if len(rng) != 2 {
		return 0, 0, errors.New("invalid range")
	}

	first, err = strconv.ParseUint(rng[0], 10, 64)
	second, err = strconv.ParseUint(rng[1], 10, 64)

	// we expect ranges [lessThanOrEqual, greaterThanOrEqual] and [greaterThanOrEqual, lessThanOrEqual]
	// but need to return [greaterOrEqual, lessThan) for btree.AscendRange()
	if first > second {
		first, second = second, first
	}

	second += 1

	return
}

func (cf CustomFilters) Sort() {
	sort.SliceStable(cf, func(i, j int) bool {
		return cf[i].Field < cf[j].Field || cf[i].Operation < cf[j].Operation || cf[i].ValueString < cf[j].ValueString
	})
}
