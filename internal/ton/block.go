package ton

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"gitlab.flora.loc/mills/tondb/internal/utils"
)

const (
	WorkchainMasterId = -1
	WorkchainTonId    = 0
)

type Block struct {
	Info         *BlockInfo     `json:"block_info"`
	ShardDescr   []*ShardDescr  `json:"shard_descr"`
	Transactions []*Transaction `json:"transactions"` // starts from extra.account_blocks
}

type WorkchainId int32

type BlockInfo struct {
	BlockId
	BlockHeader

	PrevSeqNo uint64 `json:"prev_seq_no"` // virtual field. filled only in get block info query by join.
	NextSeqNo uint64 `json:"next_seq_no"` // virtual field. same

	MinRefMcSeqno     uint32    `json:"min_ref_mc_seqno"`
	PrevKeyBlockSeqno uint32    `json:"prev_key_block_seqno"`
	GenCatchainSeqno  uint32    `json:"gen_catchain_seqno"`
	GenUtime          uint32    `json:"gen_utime"`
	Prev1Ref          *BlockRef `json:"prev1_ref"`
	Prev2Ref          *BlockRef `json:"prev2_ref"`

	// todo: vert_seqno_incr
	// todo: prev_vert_ref
	// todo: vert_seq_no

	MasterRef *BlockRef `json:"master_ref,omitempty"`

	StartLt uint64 `json:"start_lt"`
	EndLt   uint64 `json:"end_lt"`
	Version uint32 `json:"version"`

	Flags       uint8 `json:"flags"`
	KeyBlock    bool  `json:"key_block"`
	NotMaster   bool  `json:"not_master"`
	WantMerge   bool  `json:"want_merge"`
	WantSplit   bool  `json:"want_split"`
	AfterMerge  bool  `json:"after_merge"`
	AfterSplit  bool  `json:"after_split"`
	BeforeSplit bool  `json:"before_split"`

	BlockStats *BlockStats `json:"block_stats"`
	ValueFlow  *ValueFlow  `json:"value_flow"`
}

type BlockStats struct {
	TrxCount              uint16 `json:"trx_count"`
	MsgCount              uint16 `json:"msg_count"`
	SentNanograms         uint64 `json:"sent_nanograms"`
	TrxTotalFeesNanograms uint64 `json:"trx_total_fees_nanograms"`
	MsgIhrFeeNanograms    uint64 `json:"msg_ihr_fee_nanograms"`
	MsgImportFeeNanograms uint64 `json:"msg_import_fee_nanograms"`
	MsgFwdFeeNanograms    uint64 `json:"msg_fwd_fee_nanograms"`
}

type ValueFlow struct {
	FromPrevBlk   uint64 `json:"from_prev_blk"`
	ToNextBlk     uint64 `json:"to_next_blk"`
	Imported      uint64 `json:"imported"`
	Exported      uint64 `json:"exported"`
	FeesCollected uint64 `json:"fees_collected"`
	FeesImported  uint64 `json:"fees_imported"`
	Recovered     uint64 `json:"recovered"`
	Created       uint64 `json:"created"`
	Minted        uint64 `json:"minted"`
}

type BlockHeader struct {
	RootHash string `json:"root_hash"`
	FileHash string `json:"file_hash"`
}

type BlockRef struct {
	EndLt    uint64 `json:"end_lt,omitempty"`
	SeqNo    uint64 `json:"seq_no,omitempty"`
	FileHash string `json:"file_hash,omitempty"`
	RootHash string `json:"root_hash,omitempty"`
}

type BlockId struct {
	WorkchainId int32  `json:"workchain_id"`
	Shard       uint64 `json:"shard"`
	SeqNo       uint64 `json:"seq_no"`
}

func (b *BlockId) String() string {
	return fmt.Sprintf("(%d,%s,%d)",
		b.WorkchainId,
		strings.ToUpper(utils.DecToHex(b.Shard)),
		b.SeqNo,
	)
}

func ParseBlockId(b string) (*BlockId, error) {
	chunks := strings.Split(strings.Trim(b, "() "), ",")
	if len(chunks) != 3 {
		return nil, errors.New("wrong format BlockId. expected like: (WorkchainId,ShardHex,SeqNo)")
	}

	blockId := &BlockId{}

	// Input data
	wId, err := strconv.ParseInt(chunks[0], 10, 32)
	if err != nil {
		return nil, errors.New("workchain_id parse error")
	}
	blockId.WorkchainId = int32(wId)

	if blockId.Shard, err = utils.HexToDec(chunks[1]); err != nil {
		return nil, errors.New("shard is not hex")
	}
	blockId.SeqNo, err = strconv.ParseUint(chunks[2], 10, 64)
	if err != nil {
		return nil, errors.New("seq_no parse error")
	}

	return blockId, nil
}
