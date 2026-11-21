// Package adaparser is a Go transliteration of the Ada parser that Steve
// Willson wrote in MLISP on the UC Irvine PDP-10 in January 1981.
//
// The grammar is not compiled into Go. As in 1981, it is read at run time
// from the DEFPROP forms in GRAMMA.LSP and walked by a backtracking
// recursive-descent interpreter (PARSE and PARSE1 in PARSER.MLI).
//
// The original sources are in the original/ directory, byte for byte as
// they came off the tape, and in github.com/DrStephenCW/ada-parser-1981.
package adaparser

import (
	_ "embed"
	"fmt"

	"github.com/Above-the-Garage/ada-parser-1981-go/lisp"
)

// Gramma1981 is GRAMMA.LSP exactly as it was recovered from tape.
//
//go:embed original/GRAMMA.LSP
var Gramma1981 string

// Grammar holds the productions, keyed by rule name. In UCI Lisp each one
// lived on the property list of its rule name under the indicator GRAMMAR,
// and PARSER.MLI fetched it with GET(RULENAME, 'GRAMMAR).
type Grammar struct {
	rules map[lisp.Atom]lisp.Value

	// Names lists the rule names in the order they were first defined.
	Names []string

	// Duplicates lists rule names that were DEFPROPed more than once.
	// As with DEFPROP in Lisp, the later definition replaces the earlier.
	Duplicates []string
}

// LoadGrammar reads (DEFPROP name body GRAMMAR) forms from src. Forms with
// any other indicator are ignored, the way GET would never see them.
func LoadGrammar(src string) (*Grammar, error) {
	forms, err := lisp.ReadAll(src)
	if err != nil {
		return nil, err
	}
	g := &Grammar{rules: make(map[lisp.Atom]lisp.Value)}
	for i, f := range forms {
		form, ok := f.(lisp.List)
		if !ok || len(form) != 4 || !lisp.IsAtom(form[0], "DEFPROP") {
			return nil, fmt.Errorf("form %d is not (DEFPROP name body indicator): %s", i+1, lisp.String(f))
		}
		if !lisp.IsAtom(form[3], "GRAMMAR") {
			continue
		}
		name, ok := form[1].(lisp.Atom)
		if !ok {
			return nil, fmt.Errorf("form %d: rule name is not an atom: %s", i+1, lisp.String(form[1]))
		}
		if _, seen := g.rules[name]; seen {
			g.Duplicates = append(g.Duplicates, string(name))
		} else {
			g.Names = append(g.Names, string(name))
		}
		g.rules[name] = form[2]
	}
	return g, nil
}

// DefaultGrammar returns the 1981 grammar from the embedded GRAMMA.LSP.
func DefaultGrammar() *Grammar {
	g, err := LoadGrammar(Gramma1981)
	if err != nil {
		panic("adaparser: embedded GRAMMA.LSP does not load: " + err.Error())
	}
	return g
}

// Rule returns the body of the named rule: GET(name, 'GRAMMAR).
func (g *Grammar) Rule(name string) (lisp.Value, bool) {
	v, ok := g.rules[lisp.Atom(name)]
	return v, ok
}
