package filter

type KV struct {
	k string
	v interface{}
}

func (f *KV) Build() (string, []interface{}, error) {
	return f.k + "=?", []interface{}{f.v}, nil
}

func NewKV(k string, v interface{}) *KV {
	return &KV{
		k: k,
		v: v,
	}
}
