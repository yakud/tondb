package filter

import "testing"

func TestAnd_And(t *testing.T) {
	and := NewAnd(
		NewKV("a", 1),
		NewKV("b", 2),
		NewKV("c", 3),
	)

	q, a, err := and.Build()
	if err != nil {
		t.Fatal(err)
	}

	if q != "(a=? AND b=? AND c=?)" {
		t.Fatal("bad AND filter sql query")
	}
	if len(a) != 3 || a[0] != 1 || a[1] != 2 || a[2] != 3 {
		t.Fatal("bad AND filter args query:", a)
	}
}
