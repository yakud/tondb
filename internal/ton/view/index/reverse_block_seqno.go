package index

import (
	"database/sql"

	"gitlab.flora.loc/mills/tondb/internal/ton"

	"gitlab.flora.loc/mills/tondb/internal/ton/view"
)

const (

	// Store index like:
	// SeqNo: [(WorkchainId,Shard), (WorkchainId,Shard), ... ]
	createIndexReverseBlockSeqNo = `
	CREATE MATERIALIZED VIEW IF NOT EXISTS _view_index_ReverseBlockSeqNo
	ENGINE = SummingMergeTree()
	ORDER BY (SeqNo)
	PARTITION BY (round(SeqNo / 20000000))
	SETTINGS index_granularity = 32
	POPULATE 
	AS
	SELECT 
		SeqNo,
		groupArrayState((WorkchainId, Shard)) as Blocks
	FROM blocks
	GROUP BY SeqNo;
`
	dropIndexReverseBlockSeqNo = `DROP TABLE _view_index_NextBlock`

	selectBlocksBySeqNo = `
	SELECT 
		toInt32(Blocks.1) as WorkchainId,
		toUInt64(Blocks.2) as Shard,
		SeqNo
	FROM (
		SELECT 
			SeqNo, 
			arrayJoin(groupArrayMerge(Blocks)) as Blocks
		FROM ".inner._view_index_ReverseBlockSeqNo" FINAL
		PREWHERE SeqNo = ?
		GROUP BY SeqNo
	) ORDER BY WorkchainId, Shard
`
)

type IndexReverseBlockSeqNo struct {
	view.View
	conn *sql.DB
}

func (t *IndexReverseBlockSeqNo) CreateTable() error {
	_, err := t.conn.Exec(createIndexReverseBlockSeqNo)
	return err
}

func (t *IndexReverseBlockSeqNo) DropTable() error {
	_, err := t.conn.Exec(dropIndexReverseBlockSeqNo)
	return err
}

func (t *IndexReverseBlockSeqNo) SelectBlocksBySeqNo(seqNo uint64) ([]*ton.BlockId, error) {
	rows, err := t.conn.Query(selectBlocksBySeqNo, seqNo)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := make([]*ton.BlockId, 0)
	for rows.Next() {
		blockId := &ton.BlockId{}
		err := rows.Scan(
			&blockId.WorkchainId,
			&blockId.Shard,
			&blockId.SeqNo,
		)
		if err != nil {
			return nil, err
		}

		res = append(res, blockId)
	}

	return res, nil
}

func NewIndexReverseBlockSeqNo(conn *sql.DB) *IndexReverseBlockSeqNo {
	return &IndexReverseBlockSeqNo{
		conn: conn,
	}
}
