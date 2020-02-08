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
	OrigStatus            string `json:"orig_status"`
	EndStatus             string `json:"end_status"`

	PrevTransLt   uint64 `json:"prev_trans_lt"`
	PrevTransHash string `json:"prev_trans_hash"`

	StateUpdateNewHash string `json:"state_update_new_hash"`
	StateUpdateOldHash string `json:"state_update_old_hash"`

	Aborted     bool   `json:"aborted"`
	Destroyed   bool   `json:"destroyed"`
	IsTock      bool   `json:"is_tock"`
	CreditFirst bool   `json:"credit_first"`

	ActionPhase  *ActionPhase  `json:"action_phase,omitempty"`
	ComputePhase *ComputePhase `json:"compute_phase,omitempty"`
	StoragePhase *StoragePhase `json:"storage_phase,omitempty"`
	CreditPhase  *CreditPhase  `json:"credit_phase,omitempty"`
	BouncePhase  *BouncePhase  `json:"bounce_phase,omitempty"`

	InMsg   *TransactionMessage   `json:"in_msg,omitempty"`
	OutMsgs []*TransactionMessage `json:"out_msgs,omitempty"`
}
