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
		BeforeSplit,
		ValueFlowFromPrevBlk,
		ValueFlowToNextBlk,
		ValueFlowImported,
		ValueFlowExported,
		ValueFlowFeesCollected,
		ValueFlowFeesImported,
		ValueFlowRecovered,
		ValueFlowCreated,
		ValueFlowMinted,
		BlockStatsTrxCount,
		BlockStatsMsgCount,
		BlockStatsSentNanograms,
		BlockStatsTrxTotalFeesNanograms,
		BlockStatsMsgIhrFeeNanograms,
		BlockStatsMsgImportFeeNanograms,
		BlockStatsMsgFwdFeeNanograms,
	    SeqNo-1 as PrevSeqNo,
	    NextSeqNo
	FROM (
		SELECT 
			WorkchainId,
			Shard,
			SeqNo,
			Time,
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
			BeforeSplit,
			ValueFlowFromPrevBlk,
			ValueFlowToNextBlk,
			ValueFlowImported,
			ValueFlowExported,
			ValueFlowFeesCollected,
			ValueFlowFeesImported,
			ValueFlowRecovered,
			ValueFlowCreated,
			ValueFlowMinted,
			BlockStatsTrxCount,
			BlockStatsMsgCount,
			BlockStatsSentNanograms,
			BlockStatsTrxTotalFeesNanograms,
			BlockStatsMsgIhrFeeNanograms,
			BlockStatsMsgImportFeeNanograms,
			BlockStatsMsgFwdFeeNanograms
		FROM blocks
		WHERE %s
		LIMIT 100
	) ANY LEFT JOIN (
	    SELECT
			WorkchainId,
		    Shard,
		    SeqNo,
			NextSeqNo
		FROM ".inner._view_index_NextBlock"
		WHERE %s
	) USING (WorkchainId, Shard, SeqNo)
`
)

type GetBlockInfo struct {
	conn *sql.DB
}

func (q *GetBlockInfo) GetBlockInfo(f filter.Filter) ([]*ton.BlockInfo, error) {
	query, args, err := filter.RenderQuery(queryGetBlockInfo, f, f)
	if err != nil {
		return nil, err
	}

	rows, err := q.conn.Query(query, args...)
	if err != nil {
		if rows != nil {
			rows.Close()
		}
		return nil, err
	}

	res := make([]*ton.BlockInfo, 0)
	for rows.Next() {
		blockInfo := &ton.BlockInfo{
			Prev1Ref:   &ton.BlockRef{},
			Prev2Ref:   &ton.BlockRef{},
			MasterRef:  &ton.BlockRef{},
			ValueFlow:  &ton.ValueFlow{},
			BlockStats: &ton.BlockStats{},
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
			&blockInfo.ValueFlow.FromPrevBlk,
			&blockInfo.ValueFlow.ToNextBlk,
			&blockInfo.ValueFlow.Imported,
			&blockInfo.ValueFlow.Exported,
			&blockInfo.ValueFlow.FeesCollected,
			&blockInfo.ValueFlow.FeesImported,
			&blockInfo.ValueFlow.Recovered,
			&blockInfo.ValueFlow.Created,
			&blockInfo.ValueFlow.Minted,
			&blockInfo.BlockStats.TrxCount,
			&blockInfo.BlockStats.MsgCount,
			&blockInfo.BlockStats.SentNanograms,
			&blockInfo.BlockStats.TrxTotalFeesNanograms,
			&blockInfo.BlockStats.MsgIhrFeeNanograms,
			&blockInfo.BlockStats.MsgImportFeeNanograms,
			&blockInfo.BlockStats.MsgFwdFeeNanograms,
			&blockInfo.PrevSeqNo,
			&blockInfo.NextSeqNo,
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
