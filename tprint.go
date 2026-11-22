package adaparser

import (
	"fmt"
	"io"
	"strings"

	"github.com/Above-the-Garage/ada-parser-1981-go/lex"
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

// TPrintSource is TPrintText with the source interleaved. Each source line
// is printed, flush left and marked with its line number, just before the
// outermost tree node whose first token comes from that line, so the tree
// reads against the program. toks must be the tokens the tree was parsed
// from, and src the text they were scanned from.
//
// In the tree, an atom that is immediately followed by a list is a rule
// name, as PARSE built (RULENAME (...)). Every other atom except
// PLACE-HOLDER is a token, and the tokens appear in input order.
func TPrintSource(w io.Writer, x lisp.List, toks []lex.Token, src string) error {
	sp := &sourcePrinter{toks: toks, lines: strings.Split(src, "\n")}
	sp.walk(x, 0)
	_, err := io.WriteString(w, sp.b.String())
	return err
}

type sourcePrinter struct {
	b        strings.Builder
	toks     []lex.Token
	lines    []string
	next     int // index of the next token to appear in the tree
	lastLine int // last source line printed
}

// countTokens returns how many tokens the tree x holds.
func countTokens(x lisp.List) int {
	n := 0
	for i, e := range x {
		switch v := e.(type) {
		case lisp.Atom:
			if v != PlaceHolder && !isRuleName(x, i) {
				n++
			}
		case lisp.List:
			n += countTokens(v)
		}
	}
	return n
}

func isRuleName(x lisp.List, i int) bool {
	if i+1 >= len(x) {
		return false
	}
	_, ok := x[i+1].(lisp.List)
	return ok
}

// showLine prints the source line of the next token if it is new.
func (sp *sourcePrinter) showLine() {
	if sp.next >= len(sp.toks) {
		return
	}
	if ln := sp.toks[sp.next].Line; ln > sp.lastLine {
		sp.lastLine = ln
		fmt.Fprintf(&sp.b, "%4d| %s\n", ln, strings.TrimRight(sp.lines[ln-1], "\r"))
	}
}

func (sp *sourcePrinter) walk(x lisp.List, col int) {
	for i, e := range x {
		switch v := e.(type) {
		case lisp.Atom:
			if v == PlaceHolder {
				continue
			}
			if isRuleName(x, i) {
				if countTokens(x[i+1].(lisp.List)) > 0 {
					sp.showLine()
				}
			} else {
				sp.showLine()
				sp.next++
			}
			sp.b.WriteString(strings.Repeat(" ", col+1))
			sp.b.WriteString(string(v))
			sp.b.WriteByte('\n')
		case lisp.List:
			sp.walk(v, col+2)
		}
	}
}
