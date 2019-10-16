package ton

type ShardDescr struct {
	MasterShard      uint64 `json:"master_shard"`
	MasterSeqNo      uint64 `json:"master_seq_no"`
	ShardWorkchainId int32  `json:"shard_workchain_id"` // 0 temporary
	Shard            uint64 `json:"shard"`              // next_validator_shard
	ShardSeqNo       uint64 `json:"shard_seq_no"`
}
