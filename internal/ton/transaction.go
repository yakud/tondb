package ton

type Transaction struct {
	BlockShardWorkchainId int32  `json:"block_shard_workchain_id"`
	BlockShardPrefix      uint64 `json:"block_shard_prefix"`
	BlockSeqNo            uint64 `json:"block_seq_no"`

	Type                  string `json:"type"`
	Lt                    uint64 `json:"lt"`
	Now                   uint64 `json:"now"`
	TotalFeesNanograms    uint64 `json:"total_fees_nanograms"`
	TotalFeesNanogramsLen uint8  `json:"total_fees_nanograms_len"`
	AccountAddr           string `json:"account_addr"`
	OrigStatus            string `json:"orig_status"`
	EndStatus             string `json:"end_status"`

	PrevTransLt   uint64 `json:"prev_trans_lt"`
	PrevTransHash string `json:"prev_trans_hash"`

	StateUpdateNewHash string `json:"state_update_new_hash"`
	StateUpdateOldHash string `json:"state_update_old_hash"`

	InMsg   *TransactionMessage   `json:"in_msg"`
	OutMsgs []*TransactionMessage `json:"out_msgs"`
}
