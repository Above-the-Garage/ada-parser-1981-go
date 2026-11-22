package adaparser

import (
	"io"
	"strings"

	"github.com/Above-the-Garage/ada-parser-1981-go/lisp"
)

// TPrint writes the parse tree x the way EXPR TPRINT(X,COL) did: one atom
// per line, indented two more columns for each level of list nesting, with
// PLACE-HOLDER markers left out. Atoms are printed with UCI Lisp escapes,
// so a comma shows as "/,", as it does in DSKLOG.LOG.
//
// SPRINT(I, COL+2) printed at column COL+2, counting from 1, so the top
// level starts after one space. DSKLOG.LOG shows the same layout, with the
// TOPS-10 terminal driver having folded runs of eight spaces into tabs.
func TPrint(w io.Writer, x lisp.List) error {
	_, err := io.WriteString(w, Format(x))
	return err
}

// Format returns the TPrint output as a string.
func Format(x lisp.List) string {
	var b strings.Builder
	tprint(&b, x, 0, func(a lisp.Atom) string { return lisp.String(a) })
	return b.String()
}

// TPrintText is TPrint without the UCI Lisp escapes: each token is shown
// as it was written, so a string literal reads "Hello, world" rather than
// "Hello/,/ world". Use it for scanned Ada source.
func TPrintText(w io.Writer, x lisp.List) error {
	var b strings.Builder
	tprint(&b, x, 0, func(a lisp.Atom) string { return string(a) })
	_, err := io.WriteString(w, b.String())
	return err
}

func tprint(b *strings.Builder, x lisp.List, col int, show func(lisp.Atom) string) {
	for _, i := range x { // FOR NEW I IN X DO:
		switch v := i.(type) {
		case lisp.Atom:
			if v == PlaceHolder { // IF I EQ 'PLACE-HOLDER THEN NIL
				continue
			}
			b.WriteString(strings.Repeat(" ", col+1)) // SPRINT(I, COL+2)
			b.WriteString(show(v))
			b.WriteByte('\n')
		case lisp.List:
			tprint(b, v, col+2, show) // UNLESS IT'S A LIST, THEN INDENT
		}
	}
}
