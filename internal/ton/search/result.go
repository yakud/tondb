package search

type ResultType string

const (
	ResultTypeAccount     ResultType = "account"
	ResultTypeBlock       ResultType = "block"
	ResultTypeTransaction ResultType = "transaction"
	ResultTypeMessage     ResultType = "message"
)

type Result struct {
	Type ResultType `json:"type"`
	Hint string     `json:"hint"`
	Link string     `json:"link"`
}
