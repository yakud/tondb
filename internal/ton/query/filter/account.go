package filter

import (
	"gitlab.flora.loc/mills/tondb/internal/ton"
)

type Account struct {
	addr ton.AddrStd
}

func (f *Account) Addr() ton.AddrStd {
	return f.addr
}

func (f *Account) Build() (string, []interface{}, error) {
	accountFilter := NewAnd(
		NewKV("WorkchainId", f.addr.WorkchainId),
		NewKV("Addr", f.addr.Addr),
	)

	return accountFilter.Build()
}

func NewAccount(addr ton.AddrStd) *Account {
	f := &Account{
		addr: addr,
	}

	return f
}
