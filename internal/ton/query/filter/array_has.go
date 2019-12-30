package filter

type ArrayHas struct {
	arr string
	v   interface{}
}

func (f *ArrayHas) Build() (string, []interface{}, error) {
	return "has(" + f.arr + ",?)", []interface{}{f.v}, nil
}

func NewArrayHas(arr string, v interface{}) *ArrayHas {
	return &ArrayHas{
		arr: arr,
		v:   v,
	}
}
