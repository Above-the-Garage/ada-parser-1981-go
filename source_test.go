package adaparser

import (
	"os"
	"strings"
	"testing"

	"github.com/Above-the-Garage/ada-parser-1981-go/lex"
)

func parseFile(t *testing.T, name string, fixed bool) (*Parser, []lex.Token, bool) {
	t.Helper()
	src, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	toks, err := lex.Scan(string(src))
	if err != nil {
		t.Fatal(err)
	}
	p := New(DefaultGrammar())
	p.Fixed = fixed
	tree := p.ParseTokens("COMPILATION", toks)
	return p, toks, tree != nil && len(p.Remaining()) == 0
}

// TestRealAda parses real Ada source scanned by the lexer.
func TestRealAda(t *testing.T) {
	for _, name := range []string{"examples/sum_squares.adb", "examples/gcd.adb"} {
		for _, fixed := range []bool{false, true} {
			if _, _, ok := parseFile(t, name, fixed); !ok {
				t.Errorf("%s (fixed=%v) did not parse completely", name, fixed)
			}
		}
	}
}

// TestGrammarGap checks that a failure points at the right place. The 1981
// DECLARATIVE-PART is (/{ UNIVERSAL-DECLARATION ; }), but most declaration
// rules already end in ";", so "Max : constant := 100;" wants a second ";"
// and the parser gets stuck on the TYPE that follows it.
func TestGrammarGap(t *testing.T) {
	p, toks, ok := parseFile(t, "examples/stack.ada", false)
	if ok {
		t.Fatal("stack.ada parsed; the 1981 grammar should not accept it")
	}
	tk := toks[p.Furthest()]
	if tk.Line != 9 || tk.Text != "TYPE" {
		t.Errorf("stuck at %d:%d %q, want line 9 TYPE", tk.Line, tk.Col, tk.Text)
	}
}

// TestIdentifierKeepsSpelling checks that a lexical rule matches by kind
// and keeps the identifier as written.
func TestIdentifierKeepsSpelling(t *testing.T) {
	toks, _ := lex.Scan("Total + i * I")
	p := New(DefaultGrammar())
	got := Format(p.ParseTokens("SIMPLE-EXPRESSION", toks))
	for _, id := range []string{"Total", "i", "I"} {
		if !strings.Contains(got, " "+id+"\n") {
			t.Errorf("tree lacks identifier %q:\n%s", id, got)
		}
	}
}
