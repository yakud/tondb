package filter

import (
	"errors"
	"net/http"
	"strconv"

	"gitlab.flora.loc/mills/tondb/internal/ton"

	"gitlab.flora.loc/mills/tondb/internal/ton/query/filter"
)

func AddrAndLtFromRequest(r *http.Request) (filter.Filter, error) {
	addrQuery, ok := r.URL.Query()["addr"]
	if !ok || len(addrQuery) == 0 {
		return nil, nil
	}
	ltQuery, ok := r.URL.Query()["lt"]
	if !ok || len(ltQuery) == 0 {
		return nil, nil
	}

	if len(addrQuery) != len(ltQuery) {
		return nil, errors.New("different count values of addr and lt fields")
	}

	or := filter.NewOr()

	for i, addr := range addrQuery {
		accAddr, err := ton.ParseAccountAddress(addr)
		if err != nil {
			return nil, err
		}

		lt, err := strconv.ParseUint(ltQuery[i], 10, 64)
		if err != nil {
			return nil, err
		}

		f := filter.NewAnd(
			filter.NewKV("AccountAddr", accAddr.Addr),
			filter.NewKV("Lt", lt),
		)

		or.Or(f)
	}

	return or, nil
}

func AddrAndLtFromParams(addr, lt *[]string) (filter.Filter, error) {
	if addr == nil || len(*addr) == 0 {
		return nil, nil
	}

	if lt == nil || len(*lt) == 0 {
		return nil, nil
	}

	if len(*addr) != len(*lt) {
		return nil, errors.New("different count values of addr and lt fields")
	}

	or := filter.NewOr()

	for i, addr := range *addr {
		accAddr, err := ton.ParseAccountAddress(addr)
		if err != nil {
			return nil, err
		}

		lt, err := strconv.ParseUint((*lt)[i], 10, 64)
		if err != nil {
			return nil, err
		}

		f := filter.NewAnd(
			filter.NewKV("AccountAddr", accAddr.Addr),
			filter.NewKV("Lt", lt),
		)

		or.Or(f)
	}

	return or, nil
}
