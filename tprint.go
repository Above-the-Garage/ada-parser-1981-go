package adaparser

import (
	"io"
	"strings"

	"github.com/Above-the-Garage/ada-parser-1981-go/lisp"
)

// TPrint writes the parse tree x the way EXPR TPRINT(X,COL) did: one atom
// per line, indented two more columns for each level of list nesting, with
// PLACE-HOLDER markers left out.
//
// SPRINT(I, COL+2) printed at column COL+2, counting from 1, so the top
// level starts after one space. DSKLOG.LOG shows the same layout, with the
// TOPS-10 terminal driver having folded runs of eight spaces into tabs.
func TPrint(w io.Writer, x lisp.List) error {
	var b strings.Builder
	tprint(&b, x, 0)
	_, err := io.WriteString(w, b.String())
	return err
}

// Format returns the TPrint output as a string.
func Format(x lisp.List) string {
	var b strings.Builder
	tprint(&b, x, 0)
	return b.String()
}

func tprint(b *strings.Builder, x lisp.List, col int) {
	for _, i := range x { // FOR NEW I IN X DO:
		switch v := i.(type) {
		case lisp.Atom:
			if v == PlaceHolder { // IF I EQ 'PLACE-HOLDER THEN NIL
				continue
			}
			b.WriteString(strings.Repeat(" ", col+1)) // SPRINT(I, COL+2)
			b.WriteString(lisp.String(v))
			b.WriteByte('\n')
		case lisp.List:
			tprint(b, v, col+2) // UNLESS IT'S A LIST, THEN INDENT
		}
	}
}
