package ch

import (
	"database/sql"
	"fmt"

	_ "github.com/mailru/go-clickhouse"
)

func Connect(addr *string) (*sql.DB, error) {
	if addr == nil {
		defaultAddr := "http://0.0.0.0:8123/default?compress=false&debug=false"
		addr = &defaultAddr
	}

	connect, err := sql.Open("clickhouse", *addr)
	if err != nil {
		return nil, fmt.Errorf("CH connect %s error: %s", *addr, err)
	}

	if err := connect.Ping(); err != nil {
		return nil, fmt.Errorf("CH ping %s error: %s", *addr, err)
	}

	return connect, nil
}
