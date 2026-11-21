package lisp

import (
	"errors"
	"testing"
)

func TestReadEscapes(t *testing.T) {
	vals, err := ReadAll(`(PROCEDURE A IS A /, B : C ; //= /( /./. (/[ X /]) })`)
	if err != nil {
		t.Fatal(err)
	}
	got := vals[0].(List)
	want := []string{"PROCEDURE", "A", "IS", "A", ",", "B", ":", "C", ";", "/=", "(", ".."}
	for i, w := range want {
		if !IsAtom(got[i], w) {
			t.Errorf("element %d = %v, want atom %q", i, got[i], w)
		}
	}
	sub := got[len(want)].(List)
	if !IsAtom(sub[0], "[") || !IsAtom(sub[2], "]") {
		t.Errorf("sublist = %s", String(sub))
	}
	if !IsAtom(got[len(want)+1], "}") {
		t.Errorf("last = %v, want }", got[len(want)+1])
	}
}

func TestRoundTrip(t *testing.T) {
	for _, src := range []string{
		`(A /, B)`,
		`((^ PRAGMA) IDENTIFIER (/[ /( ARGUMENT (/{ /, ARGUMENT }) /) /]) ;)`,
		`(+ ! /- ! &)`,
	} {
		v, err := ReadAll(src)
		if err != nil {
			t.Fatal(err)
		}
		if got := String(v[0]); got != src {
			t.Errorf("String(Read(%q)) = %q", src, got)
		}
	}
}

func TestUnbalanced(t *testing.T) {
	for _, src := range []string{"(A B", "A )"} {
		if _, err := ReadAll(src); !errors.Is(err, ErrUnbalanced) {
			t.Errorf("ReadAll(%q) error = %v, want ErrUnbalanced", src, err)
		}
	}
}
