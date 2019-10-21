package filter

import "testing"

func TestOr_Simple(t *testing.T) {
	and := NewOr(
		NewKV("a", 1),
		NewKV("b", 2),
		NewKV("c", 3),
	)

	q, a, err := and.Build()
	if err != nil {
		t.Fatal(err)
	}

	if q != "(a=? OR b=? OR c=?)" {
		t.Fatal("bad OR filter sql query")
	}
	if len(a) != 3 || a[0] != 1 || a[1] != 2 || a[2] != 3 {
		t.Fatal("bad OR filter args query:", a)
	}
}

func TestOr_OrOr(t *testing.T) {
	and := NewOr(
		NewOr(
			NewKV("a", 1),
			NewKV("b", 2),
		),
		NewKV("c", 3),
	)

	q, a, err := and.Build()
	if err != nil {
		t.Fatal(err)
	}

	if q != "((a=? OR b=?) OR c=?)" {
		t.Fatal("bad OR filter sql query")
	}
	if len(a) != 3 || a[0] != 1 || a[1] != 2 || a[2] != 3 {
		t.Fatal("bad OR filter args query:", a)
	}
}

func TestOr_OrAnd(t *testing.T) {
	and := NewOr(
		NewAnd(
			NewKV("a", 1),
			NewKV("b", 2),
		),
		NewKV("c", 3),
	)

	q, a, err := and.Build()
	if err != nil {
		t.Fatal(err)
	}

	if q != "((a=? AND b=?) OR c=?)" {
		t.Fatal("bad OR filter sql query")
	}
	if len(a) != 3 || a[0] != 1 || a[1] != 2 || a[2] != 3 {
		t.Fatal("bad OR filter args query:", a)
	}
}
