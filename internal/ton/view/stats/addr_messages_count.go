package stats

import (
	"database/sql"

	"gitlab.flora.loc/mills/tondb/internal/utils"
)

const (
	createAddrMessagesCount = `
	CREATE MATERIALIZED VIEW IF NOT EXISTS _view_stats_AddrMessagesCountTop
	ENGINE = SummingMergeTree() 
	PARTITION BY tuple()
	ORDER BY (Direction, Addr, WorkchainId)
	POPULATE 
	AS
	SELECT
	    Messages.Direction as Direction,
		if(Direction = 'in', Messages.DestAddr, Messages.SrcAddr) AS Addr,
	    WorkchainId,
		count() as Count
	FROM transactions
	ARRAY JOIN Messages
	WHERE Type = 'trans_ord' AND Messages.Type = 'int_msg_info'
	GROUP BY Direction, Addr, WorkchainId
`

	selectAddrMessagesCountTop = `SELECT
    Direction,
    WorkchainId,
    Addr,
    sum(Count) AS cnt
FROM ".inner._view_stats_AddrMessagesCountTop"
GROUP BY
    Direction,
    Addr,
    WorkchainId
ORDER BY cnt DESC
LIMIT ? BY Direction
`

	dropAddrMessagesCount = `DROP TABLE _view_stats_AddrMessagesCountTop`
)

type AddrCount struct {
	WorkchainId int32  `json:"workchain_id"`
	Addr        string `json:"addr"`
	AddrUf      string `json:"addr_uf"`
	Count       int64  `json:"count"`
}

type AddrMessagesCount struct {
	conn *sql.DB
}

func (t *AddrMessagesCount) CreateTable() error {
	_, err := t.conn.Exec(createAddrMessagesCount)
	return err
}

func (t *AddrMessagesCount) DropTable() error {
	_, err := t.conn.Exec(dropAddrMessagesCount)
	return err
}

// Select in and out top addr by messages count
func (t *AddrMessagesCount) SelectTopMessagesCount(topN int) ([]AddrCount, []AddrCount, error) {
	rows, err := t.conn.Query(selectAddrMessagesCountTop, topN)
	if err != nil {
		if rows != nil {
			rows.Close()
		}

		return nil, nil, err
	}

	var direction string
	topIn := make([]AddrCount, 0, topN)
	topOut := make([]AddrCount, 0, topN)
	for rows.Next() {
		row := AddrCount{}
		err := rows.Scan(
			&direction,
			&row.WorkchainId,
			&row.Addr,
			&row.Count,
		)
		if err != nil {
			rows.Close()
			return nil, nil, err
		}

		row.Addr = utils.NullAddrToString(row.Addr)

		if row.AddrUf, err = utils.ComposeRawAndConvertToUserFriendly(row.WorkchainId, row.Addr); err != nil {
			// Maybe we shouldn't fail here
			return nil, nil, err
		}

		switch direction {
		case "in":
			topIn = append(topIn, row)
		case "out":
			topOut = append(topOut, row)
		}
	}

	if rows != nil {
		rows.Close()
	}

	return topIn, topOut, err
}

func NewAddrMessagesCount(conn *sql.DB) *AddrMessagesCount {
	return &AddrMessagesCount{
		conn: conn,
	}
}
