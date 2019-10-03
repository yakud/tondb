package ton

/*
"info":{
      "@type":"block_info",
      "after_merge":"0",
      "after_split":"0",
      "before_split":"0",
      "end_lt":"295011000004",
      "flags":"0",
      "gen_catchain_seqno":"2959",
      "gen_utime":"1569593831",
      "gen_validator_list_hash_short":"1963064503",
      "key_block":"0",
      "master_ref":{
         "@type":"master_info",
         "master":{
            "@type":"ext_blk_ref",
            "end_lt":"295009000004",
            "file_hash":"x07F6919B1269341AA6F36282A526BB7E4C77D2996E0E9D2B3ABCA6DC84FF50DD",
            "root_hash":"x8AAFE98D2F80476929ED86EF31D0CF9331D2DEE579D5BA66FF9CB1D245802EB0",
            "seq_no":"171159"
         }
      },
      "min_ref_mc_seqno":"171159",
      "not_master":"1",
      "prev_key_block_seqno":"170510",
      "prev_ref":{
         "@type":"prev_blk_info",
         "prev":{
            "@type":"ext_blk_ref",
            "end_lt":"295010000001",
            "file_hash":"xBC92309A2A0BD16C03C40AD0845DD5F1D51984F4B55FBC7EA785E0B077F98029",
            "root_hash":"x6667E525FF8065AEB09F7D871BF772C6B3AAFF6446AFF8AE4CF8651E9E770373",
            "seq_no":"279098"
         }
      },
      "seq_no":"279099",
      "shard":{
         "@type":"shard_ident",
         "shard_pfx_bits":"2",
         "shard_prefix":"13835058055282163712",
         "workchain_id":"0"
      },
      "start_lt":"295011000000",
      "version":"0",
      "vert_seq_no":"0",
      "vert_seqno_incr":"0",
      "want_merge":"1",
      "want_split":"0"
   },
*/
type Block struct {
	Info *BlockInfo `json:"block_info"`

	Transactions []*Transaction `json:"transactions"` // starts from extra.account_blocks
}

type BlockInfo struct {
	ShardWorkchainId int32  `json:"shard_workchain_id"`
	ShardPrefix      uint64 `json:"shard_prefix"`
	ShardPfxBits     uint8  `json:"shard_pfx_bits"`
	SeqNo            uint64 `json:"seq_no"`

	MinRefMcSeqno     uint32    `json:"min_ref_mc_seqno"`
	PrevKeyBlockSeqno uint32    `json:"prev_key_block_seqno"`
	GenCatchainSeqno  uint32    `json:"gen_catchain_seqno"`
	GenUtime          uint32    `json:"gen_utime"`
	PrevRef           *BlockRef `json:"prev_ref"`
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
