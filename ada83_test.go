package adaparser

import (
	"path/filepath"
	"sort"
	"testing"

	"github.com/Above-the-Garage/ada-parser-1981-go/lex"
	"github.com/Above-the-Garage/ada-parser-1981-go/lisp"
)

// TestAda83Examples parses every example program with the corrected
// grammar, including the ones the 1981 grammar cannot handle.
func TestAda83Examples(t *testing.T) {
	files, _ := filepath.Glob("examples/*.ad[abs]")
	more, _ := filepath.Glob("examples/ada83/*.ad[abs]")
	files = append(files, more...)
	if len(files) < 7 {
		t.Fatalf("found only %d example files", len(files))
	}
	for _, name := range files {
		t.Run(name, func(t *testing.T) {
			p := parseWith(t, Ada83Grammar(), name)
			if rest := p.Remaining(); len(rest) > 0 {
				t.Errorf("stopped with %d tokens left; furthest token %d", len(rest), p.Furthest())
			}
		})
	}
}

// TestAda83Errors checks that bad programs fail at the right token.
func TestAda83Errors(t *testing.T) {
	for _, tc := range []struct {
		src        string
		line, col  int
		wantTokens string
	}{
		{"procedure P is begin X := 1 end P;", 1, 29, "END"},
		{"procedure P is begin if X then null; end; end P;", 1, 41, ";"},
		{"package Q is\n  X : Integer := ;\nend Q;", 2, 18, ";"},
	} {
		toks, err := lex.Scan(tc.src)
		if err != nil {
			t.Fatal(err)
		}
		p := New(Ada83Grammar())
		p.Fixed = true
		p.ParseTokens("COMPILATION", toks)
		if len(p.Remaining()) == 0 {
			t.Errorf("%q parsed, want an error", tc.src)
			continue
		}
		tk := toks[p.Furthest()]
		if tk.Line != tc.line || tk.Col != tc.col || tk.Text != tc.wantTokens {
			t.Errorf("%q: stuck at %d:%d %q, want %d:%d %q",
				tc.src, tk.Line, tk.Col, tk.Text, tc.line, tc.col, tc.wantTokens)
		}
	}
}

// TestAda83Undefined checks that every rule the corrected grammar refers
// to is defined (unlike GRAMMA.LSP's TASK-DECLARACTION and IDENTIFIR).
func TestAda83Undefined(t *testing.T) {
	if missing := undefinedRules(Ada83Grammar()); len(missing) > 0 {
		t.Errorf("undefined rules: %v", missing)
	}
	got := undefinedRules(DefaultGrammar())
	t.Logf("GRAMMA.LSP refers to undefined rules: %v", got)
}

func undefinedRules(g *Grammar) []string {
	seen := map[string]bool{}
	var walk func(v lisp.Value)
	walk = func(v lisp.Value) {
		switch x := v.(type) {
		case lisp.Atom:
			s := string(x)
			if s[0] >= 'A' && s[0] <= 'Z' && s != string(PlaceHolder) {
				if _, ok := g.Rule(s); !ok {
					if _, lexical := lexicalRules[x]; !lexical {
						seen[s] = true
					}
				}
			}
		case lisp.List:
			if len(x) > 0 && lisp.IsAtom(x[0], atomReserved) {
				return
			}
			for _, e := range x {
				walk(e)
			}
		}
	}
	for _, n := range g.Names {
		r, _ := g.Rule(n)
		walk(r)
	}
	var out []string
	for s := range seen {
		out = append(out, s)
	}
	sort.Strings(out)
	return out
}
