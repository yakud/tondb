package ton

type ShardDescr struct {
	MasterShard      uint64 `json:"master_shard_prefix"`
	MasterSeqNo      uint64 `json:"master_seq_no"`
	ShardWorkchainId int32  `json:"shard_workchain_id"` // 0 temporary
	ShardPrefix      uint64 `json:"shard_prefix"`       // next_validator_shard
	ShardSeqNo       uint64 `json:"shard_seq_no"`
}
