package query

import (
	"database/sql"
	"gitlab.flora.loc/mills/tondb/swagger/tonapi"
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

func (q *GetBlockchainHeight) GetBlockchainHeight() (*tonapi.BlockId, error) {
	row := q.conn.QueryRow(queryGetBlockchainHeight)

	var lastMasterHeight uint64
	if err := row.Scan(&lastMasterHeight); err != nil {
		return nil, err
	}

	return &tonapi.BlockId{
		WorkchainId: -1,
		Shard:       0,
		SeqNo:       tonapi.Uint64(lastMasterHeight),
	}, nil
}

func NewGetBlockchainHeight(conn *sql.DB) *GetBlockchainHeight {
	return &GetBlockchainHeight{
		conn: conn,
	}
}
