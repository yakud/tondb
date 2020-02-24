package query

import (
	"database/sql"

	"gitlab.flora.loc/mills/tondb/internal/ton/query/filter"

	"gitlab.flora.loc/mills/tondb/internal/ton"
)

const (
	queryGetBlockInfo = `
	SELECT 
		WorkchainId,
		Shard,
		SeqNo,
		toUInt64(Time),
		RootHash,
		FileHash,
		MinRefMcSeqno,
		PrevKeyBlockSeqno,
		GenCatchainSeqno,
		Prev1RefEndLt,
		Prev1RefSeqNo,
		Prev1RefFileHash,
		Prev1RefRootHash,
		Prev2RefEndLt,
		Prev2RefSeqNo,
		Prev2RefFileHash,
		Prev2RefRootHash,
		MasterRefEndLt,
		MasterRefSeqNo,
		MasterRefFileHash,
		MasterRefRootHash,
		StartLt,
		EndLt,
		Version,
		Flags,
		KeyBlock,
		NotMaster,
		WantMerge,
		WantSplit,
		AfterMerge,
		AfterSplit,
		BeforeSplit
	FROM blocks
	WHERE %s
	LIMIT 100;
`
)

type GetBlockInfo struct {
	conn *sql.DB
}

func (q *GetBlockInfo) GetBlockInfo(f filter.Filter) ([]*ton.BlockInfo, error) {
	query, args, err := filter.RenderQuery(queryGetBlockInfo, f)
	if err != nil {
		return nil, err
	}

	rows, err := q.conn.Query(query, args...)
	if err != nil {
		rows.Close()
		return nil, err
	}

	res := make([]*ton.BlockInfo, 0)
	for rows.Next() {
		blockInfo := &ton.BlockInfo{
			Prev1Ref:  &ton.BlockRef{},
			Prev2Ref:  &ton.BlockRef{},
			MasterRef: &ton.BlockRef{},
		}

		err = rows.Scan(
			&blockInfo.WorkchainId,
			&blockInfo.Shard,
			&blockInfo.SeqNo,
			&blockInfo.GenUtime,
			&blockInfo.RootHash,
			&blockInfo.FileHash,
			&blockInfo.MinRefMcSeqno,
			&blockInfo.PrevKeyBlockSeqno,
			&blockInfo.GenCatchainSeqno,
			&blockInfo.Prev1Ref.EndLt,
			&blockInfo.Prev1Ref.SeqNo,
			&blockInfo.Prev1Ref.FileHash,
			&blockInfo.Prev1Ref.RootHash,
			&blockInfo.Prev2Ref.EndLt,
			&blockInfo.Prev2Ref.SeqNo,
			&blockInfo.Prev2Ref.FileHash,
			&blockInfo.Prev2Ref.RootHash,
			&blockInfo.MasterRef.EndLt,
			&blockInfo.MasterRef.SeqNo,
			&blockInfo.MasterRef.FileHash,
			&blockInfo.MasterRef.RootHash,
			&blockInfo.StartLt,
			&blockInfo.EndLt,
			&blockInfo.Version,
			&blockInfo.Flags,
			&blockInfo.KeyBlock,
			&blockInfo.NotMaster,
			&blockInfo.WantMerge,
			&blockInfo.WantSplit,
			&blockInfo.AfterMerge,
			&blockInfo.AfterSplit,
			&blockInfo.BeforeSplit,
		)
		if err != nil {
			rows.Close()
			return nil, err
		}

		if blockInfo.Prev1Ref.SeqNo == 0 {
			blockInfo.Prev1Ref = nil
		}
		if blockInfo.Prev2Ref.SeqNo == 0 {
			blockInfo.Prev2Ref = nil
		}
		if blockInfo.MasterRef.SeqNo == 0 {
			blockInfo.MasterRef = nil
		}

		res = append(res, blockInfo)
	}
	rows.Close()

	return res, nil
}

func NewGetBlockInfo(conn *sql.DB) *GetBlockInfo {
	return &GetBlockInfo{
		conn: conn,
	}
}
