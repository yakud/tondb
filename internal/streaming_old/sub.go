package streaming_old

import (
	"errors"
	"github.com/google/uuid"
	"gitlab.flora.loc/mills/tondb/internal/ton"
	"net"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var rangeRegexp = regexp.MustCompile(`(?m)\[(\d+), *(\d+)\]`)

const (
	EqOp    = "eq"
	LtOp    = "lt"
	GtOp    = "gt"
	RangeOp = "range"
)

type Params struct {
	Filter

	FetchFromDb   *uint32       `json:"fetch_from_db"`
	CustomFilters CustomFilters `json:"custom_filters"`
}

type Filter struct {
	FeedName      string  `json:"feed_name"`
	WorkchainId   *int32  `json:"workchain_id"`
	Shard         *uint64 `json:"shard"`
	AccountAddr   *string `json:"account_addr"`

	customFilters string
}

type CustomFilters []CustomFilter

type CustomFilter struct {
	Field       string `json:"field"`
	Operation   string `json:"operation"` // eq, lt, gt, range
	ValueString string `json:"value_string"`
}

type Sub struct {
	Conn   net.Conn
	Filter Filter
	Uuid   string
}

func (f *Filter) MatchWorkchainAndShard(block *ton.Block) bool {
	return (f.WorkchainId == nil || (f.WorkchainId != nil && block.Info.WorkchainId == *f.WorkchainId)) &&
		(f.Shard == nil || (f.Shard != nil && block.Info.Shard == *f.Shard))
}

func NewSub(conn net.Conn, filter Filter) *Sub {
	return &Sub{
		Conn: conn,
		Filter: filter,
		Uuid: uuid.New().String(),
	}
}

func NewSubUuid(conn net.Conn, filter Filter, id string) *Sub {
	return &Sub{
		Conn: conn,
		Filter: filter,
		Uuid: id,
	}
}

func (cf *CustomFilter) ParseRange() (first, second uint64, err error) {
	if cf.Operation != RangeOp {
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

func (cf CustomFilters) String() string {
	sb := strings.Builder{}
	for i, filter := range cf {
		sb.WriteString(filter.Field+" "+filter.Operation+" "+filter.ValueString)
		if i < len(cf) - 1 {
			sb.WriteString(";")
		}
	}
	return sb.String()
}

func ParseCustomFilters(str string) CustomFilters {
	filters := strings.Split(str, ";")
	res := make(CustomFilters, 0, len(filters))

	for _, filter := range filters {
		fields := strings.SplitN(filter, " ", 3)
		if len(fields) == 3 {
			res = append(res, CustomFilter{Field: fields[0], Operation: fields[1], ValueString: fields[2]})
		}
	}

	return res
}