package lex

import (
	"strings"
	"testing"
)

func TestScan(t *testing.T) {
	src := `with Text_IO; -- a comment
X := Text_IO.Put ("say""hi""") ** 2#1010# + 1_000.5E-3 + C'Last;
if A /= B then Q := 'x'; for I in 1..N loop << L >> <> => end if;`
	toks, err := Scan(src)
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, tk := range toks {
		got = append(got, tk.Kind.String()[:3]+":"+tk.Text)
	}
	want := strings.Fields(`res:WITH ide:Text_IO del:;
		ide:X del::= ide:Text_IO del:. ide:Put del:( str:"say""hi""" del:) del:** num:2#1010# del:+ num:1_000.5E-3 del:+ ide:C del:' ide:Last del:;
		res:IF ide:A del:/= ide:B res:THEN ide:Q del::= cha:'x' del:; res:FOR ide:I res:IN num:1 del:.. ide:N res:LOOP
		del:<< ide:L del:>> del:<> del:=> res:END res:IF del:;`)
	if strings.Join(got, " ") != strings.Join(want, " ") {
		t.Errorf("got\n%v\nwant\n%v", got, want)
	}
	if toks[3].Line != 2 || toks[3].Col != 1 {
		t.Errorf("X at %d:%d, want 2:1", toks[3].Line, toks[3].Col)
	}
}

func TestScanErrors(t *testing.T) {
	for _, src := range []string{`"unterminated`, "X := $;"} {
		if _, err := Scan(src); err == nil {
			t.Errorf("Scan(%q) succeeded, want an error", src)
		}
	}
}
