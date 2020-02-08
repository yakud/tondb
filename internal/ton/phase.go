package ton

type ActionPhase struct {
	Success           bool   `json:"success"`
	Valid             bool   `json:"valid"`
	NoFunds           bool   `json:"no_funds"`
	CodeChanged       bool   `json:"code_changed,omitempty"`
	ActionListInvalid bool   `json:"action_list_invalid,omitempty"`
	AccDeleteReq      bool   `json:"acc_delete_req,omitempty"`
	AccStatusChange   string `json:"acc_status_change"`
	TotalFwdFees      uint64 `json:"total_fwd_fees"`
	TotalActionFees   uint64 `json:"total_action_fees"`
	ResultCode        int32  `json:"result_code"`
	ResultArg         int32  `json:"result_arg"`
	TotActions        uint32 `json:"tot_actions"`
	SpecActions       uint32 `json:"spec_actions"`
	SkippedActions    uint32 `json:"skipped_actions"`
	MsgsCreated       uint32 `json:"msgs_created"`

	RemainingBalance uint64 `json:"remaining_balance,omitempty"`
	ReservedBalance  uint64 `json:"reserved_balance,omitempty"`
	EndLt            uint64 `json:"end_lt,omitempty"`
	TotMsgBits       uint64 `json:"tot_msg_bits"`
	TotMsgCells      uint64 `json:"tot_msg_cells"`
}

type ComputePhase struct {
	Skipped          bool   `json:"skipped"`
	SkippedReason    string `json:"skipped_reason,omitempty"`
	AccountActivated bool   `json:"account_activated"`
	Success          bool   `json:"success"`
	MsgStateUsed     bool   `json:"msg_state_used"`
	OutOfGas         bool   `json:"out_of_gas"`
	Accepted         bool   `json:"accepted"`
	ExitArg          int32  `json:"exit_arg"`
	ExitCode         int32  `json:"exit_code"`
	Mode             int32  `json:"mode"`
	VmSteps          uint32 `json:"vm_steps"`

	GasUsed   uint64 `json:"gas_used"`
	GasMax    uint64 `json:"gas_max"`
	GasCredit uint64 `json:"gas_credit"`
	GasLimit  uint64 `json:"gas_limit"`
	GasFees   uint64 `json:"gas_fees"`
}

type StoragePhase struct {
	Status        string `json:"status"`
	FeesCollected uint64 `json:"fees_collected"`
	FeesDue       uint64 `json:"fees_due"`
}

type CreditPhase struct {
	DueFeesCollected uint64 `json:"due_fees_collected"`
	CreditNanograms  uint64 `json:"credit_nanograms"`
}

// TODO; Could't find examples of bounce phase in blocks. If there were some they were empty.
type BouncePhase struct {

}
