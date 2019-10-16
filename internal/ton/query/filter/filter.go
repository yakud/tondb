package filter

type Builder interface {
	Build() (sqlFilter string, args []interface{}, err error)
}
