package ton

type TransactionMessage struct {
	Type    string `json:"type"`
	Init    string `json:"init,omitempty"`
	Bounce  bool   `json:"bounce"`
	Bounced bool   `json:"bounced"`

	CreatedAt uint64 `json:"created_at"`
	CreatedLt uint64 `json:"created_lt"`

	ValueNanograms    uint64 `json:"value_nanograms,omitempty"`
	ValueNanogramsLen uint8  `json:"value_nanograms_len,omitempty"`

	// Fees
	FwdFeeNanograms       uint64 `json:"fwd_fee_nanograms,omitempty"`
	FwdFeeNanogramsLen    uint8  `json:"fwd_fee_nanograms_len,omitempty"`
	IhrDisabled           bool   `json:"ihr_disabled,omitempty"`
	IhrFeeNanograms       uint64 `json:"ihr_fee_nanograms,omitempty"`
	IhrFeeNanogramsLen    uint8  `json:"ihr_fee_nanograms_len,omitempty"`
	ImportFeeNanograms    uint64 `json:"import_fee_nanograms,omitempty"`
	ImportFeeNanogramsLen uint8  `json:"import_fee_nanograms_len,omitempty"`

	// Dest
	Dest AddrStd `json:"dest"`
	Src  AddrStd `json:"src"`

	// Body
	BodyType  string `json:"body_type"`
	BodyValue string `json:"body_value"`
}

type AddrStd struct {
	IsEmpty     bool   `json:"is_empty"`
	WorkchainId int32  `json:"workchain_id,omitempty"`
	Addr        string `json:"addr,omitempty"`
	Anycast     string `json:"anycast,omitempty"`
}
