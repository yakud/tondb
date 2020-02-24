package search

type ResultType string

const (
	ResultTypeAccount     ResultType = "account"
	ResultTypeBlock       ResultType = "block"
	ResultTypeTransaction ResultType = "transaction"
	ResultTypeMessage     ResultType = "message"
)

type Result struct {
	Type ResultType
	Link string
}
