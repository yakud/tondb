package ton

type Transaction struct {
	BlockId

	Hash                  string `json:"hash"`
	Type                  string `json:"type"`
	Lt                    uint64 `json:"lt"`
	Now                   uint64 `json:"now"`
	TotalFeesNanograms    uint64 `json:"total_fees_nanograms"`
	TotalFeesNanogramsLen uint8  `json:"total_fees_nanograms_len"`
	AccountAddr           string `json:"account_addr"`
	AccountAddrUf         string `json:"account_addr_uf"`
	OrigStatus            string `json:"orig_status"`
	EndStatus             string `json:"end_status"`

	PrevTransLt   uint64 `json:"prev_trans_lt"`
	PrevTransHash string `json:"prev_trans_hash"`

	StateUpdateNewHash string `json:"state_update_new_hash"`
	StateUpdateOldHash string `json:"state_update_old_hash"`

	// virtual field. calculates only when retrieved from db
	TotalNanograms uint64 `json:"total_nanograms"`
	IsTock         bool   `json:"is_tock"`

	InMsg   *TransactionMessage   `json:"in_msg,omitempty"`
	OutMsgs []*TransactionMessage `json:"out_msgs,omitempty"`
}
