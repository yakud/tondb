package ton

const (
	WorkchainMasterId = -1
	WorkchainTonId    = 0
)

type BlockId struct {
	WorkchainId int32  `json:"workchain_id"`
	Shard       uint64 `json:"shard"`
	SeqNo       uint64 `json:"seq_no"`
}

type Block struct {
	Info         *BlockInfo     `json:"block_info"`
	ShardDescr   []*ShardDescr  `json:"shard_descr"`
	Transactions []*Transaction `json:"transactions"` // starts from extra.account_blocks
}

type BlockInfo struct {
	WorkchainId int32  `json:"workchain_id"`
	Shard       uint64 `json:"shard"`
	SeqNo       uint64 `json:"seq_no"`

	MinRefMcSeqno     uint32    `json:"min_ref_mc_seqno"`
	PrevKeyBlockSeqno uint32    `json:"prev_key_block_seqno"`
	GenCatchainSeqno  uint32    `json:"gen_catchain_seqno"`
	GenUtime          uint32    `json:"gen_utime"`
	Prev1Ref          *BlockRef `json:"prev1_ref"`
	Prev2Ref          *BlockRef `json:"prev2_ref"`
	MasterRef         *BlockRef `json:"master_ref,omitempty"`

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
}

type BlockRef struct {
	EndLt    uint64 `json:"end_lt,omitempty"`
	SeqNo    uint64 `json:"seq_no,omitempty"`
	FileHash string `json:"file_hash,omitempty"`
	RootHash string `json:"root_hash,omitempty"`
}
