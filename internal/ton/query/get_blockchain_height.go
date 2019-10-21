package query

import (
	"database/sql"

	"gitlab.flora.loc/mills/tondb/internal/ton"
)

const (
	queryGetBlockchainHeight = `
	SELECT 
		max(SeqNo) as SeqNo
	FROM blocks
	WHERE WorkchainId = -1
`
)

type GetBlockchainHeight struct {
	conn *sql.DB
}

func (q *GetBlockchainHeight) GetBlockchainHeight() (*ton.BlockId, error) {
	row := q.conn.QueryRow(queryGetBlockchainHeight)

	var lastMasterHeight uint64
	if err := row.Scan(&lastMasterHeight); err != nil {
		return nil, err
	}

	return &ton.BlockId{
		WorkchainId: -1,
		Shard:       0,
		SeqNo:       lastMasterHeight,
	}, nil
}

func NewGetBlockchainHeight(conn *sql.DB) *GetBlockchainHeight {
	return &GetBlockchainHeight{
		conn: conn,
	}
}
