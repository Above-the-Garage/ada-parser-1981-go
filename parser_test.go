package adaparser

import (
	"bufio"
	"os"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/Above-the-Garage/ada-parser-1981-go/lisp"
)

// session is one parse recorded in DSKLOG.LOG: the input that (RESET)
// echoed, the rule given to PARSE, and the tree TPRINT printed.
type session struct {
	input []string
	rule  string
	want  string
}

var tprintCmd = regexp.MustCompile(`^\*\(TPRINT \(PARSE '([A-Z-]+)\)\)$`)

// expandTabs undoes the TOPS-10 terminal's tab compression (stops every 8).
func expandTabs(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r == '\t' {
			b.WriteString(strings.Repeat(" ", 8-b.Len()%8))
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

// readLog pulls the three parses out of the 1981 transcript.
func readLog(t *testing.T) []session {
	t.Helper()
	f, err := os.Open("original/DSKLOG.LOG")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	var lines []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		lines = append(lines, strings.TrimRight(sc.Text(), "\r"))
	}
	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}

	var out []session
	var input []string
	for i := 0; i < len(lines); i++ {
		switch line := lines[i]; {
		case line == "*(RESET)":
			// (RESET) returns, and the REPL echoes, the new INPUT-STREAM.
			i++
			vals, err := lisp.ReadAll(lines[i])
			if err != nil || len(vals) != 1 {
				t.Fatalf("line %d: cannot read input %q: %v", i+1, lines[i], err)
			}
			input = nil
			for _, v := range vals[0].(lisp.List) {
				input = append(input, string(v.(lisp.Atom)))
			}
		case tprintCmd.MatchString(line):
			s := session{input: input, rule: tprintCmd.FindStringSubmatch(line)[1]}
			var want strings.Builder
			for i++; lines[i] != "NIL"; i++ {
				want.WriteString(expandTabs(lines[i]))
				want.WriteByte('\n')
			}
			s.want = want.String()
			out = append(out, s)
		}
	}
	return out
}

// TestDSKLOG replays the three parses from the January 19, 1981 session
// and checks that the Go port prints the very same trees.
func TestDSKLOG(t *testing.T) {
	sessions := readLog(t)
	if len(sessions) != 3 {
		t.Fatalf("found %d parses in DSKLOG.LOG, want 3", len(sessions))
	}
	for _, fixed := range []bool{false, true} {
		for _, s := range sessions {
			name := s.rule + " " + strings.Join(s.input, " ")
			if fixed {
				name = "fixed/" + name
			}
			t.Run(name, func(t *testing.T) {
				p := New(DefaultGrammar())
				p.Fixed = fixed
				tree := p.Parse(s.rule, s.input)
				if tree == nil {
					t.Fatal("parse failed")
				}
				if got := Format(tree); got != s.want {
					t.Errorf("tree differs from DSKLOG.LOG\n--- got\n%s--- want\n%s", got, s.want)
				}
				if rest := p.Remaining(); len(rest) != 0 {
					t.Errorf("unconsumed input: %v", rest)
				}
			})
		}
	}
}

func TestGrammarLoads(t *testing.T) {
	g := DefaultGrammar()
	if n := len(g.Names); n < 130 {
		t.Errorf("loaded %d rules, expected about 140", n)
	}
	slices.Sort(g.Duplicates)
	if want := []string{"BODY", "CHOICE", "PROGRAM-COMPONENT"}; !slices.Equal(g.Duplicates, want) {
		t.Errorf("duplicates = %v, want %v", g.Duplicates, want)
	}
}

// TestChoiceAlternation shows the off-by-two alternation in CHOICE:
// SIMPLE-EXPRESSION ! NAME RANGE-CONSTRAINT ! RANGE ! (^ OTHERS)
// In 1981 the CDDR stepping lands on "!" and never reaches (^ OTHERS).
func TestChoiceAlternation(t *testing.T) {
	g := DefaultGrammar()
	in := []string{"OTHERS"}

	if tree := New(g).Parse("CHOICE", in); tree != nil {
		t.Errorf("1981 mode parsed OTHERS as a CHOICE: %s", lisp.String(tree))
	}

	p := New(g)
	p.Fixed = true
	tree := p.Parse("CHOICE", in)
	if got, want := lisp.String(tree), "(CHOICE (OTHERS))"; got != want {
		t.Errorf("fixed mode: got %s, want %s", got, want)
	}
}

// TestBarIsRuleName shows the punctuation test: "|" is ASCII 124, above
// "A", so the 1981 parser looks it up as a production.
func TestBarIsRuleName(t *testing.T) {
	g, err := LoadGrammar(`
		(DEFPROP CHOICES (CHOICE (/{ | CHOICE })) GRAMMAR)
		(DEFPROP CHOICE ((^ X) ! (^ Y)) GRAMMAR)`)
	if err != nil {
		t.Fatal(err)
	}
	in := []string{"X", "|", "Y"}

	var diag strings.Builder
	p := New(g)
	p.Diag = &diag
	p.Parse("CHOICES", in)
	if rest := p.Remaining(); !slices.Equal(rest, []string{"|", "Y"}) {
		t.Errorf("1981 mode left %v, want [| Y]", rest)
	}
	if !strings.Contains(diag.String(), "Can't find production for |") {
		t.Errorf("1981 mode diagnostics = %q", diag.String())
	}

	p = New(g)
	p.Fixed = true
	tree := p.Parse("CHOICES", in)
	if got, want := lisp.String(tree), "(CHOICES (CHOICE (X) | CHOICE (Y)))"; got != want {
		t.Errorf("fixed mode: got %s, want %s", got, want)
	}
}

// TestBacktracking checks that a failed alternative gives its tokens back.
func TestBacktracking(t *testing.T) {
	g := DefaultGrammar()
	// SUBTYPE-INDICATION tries (NAME RANGE-CONSTRAINT) first; that consumes
	// NAME, fails on RANGE-CONSTRAINT, and must rewind before plain NAME.
	tree := New(g).Parse("SUBTYPE-INDICATION", []string{"C"})
	want := " SUBTYPE-INDICATION\n   NAME\n     IDENTIFIER\n       C\n"
	if got := Format(tree); got != want {
		t.Errorf("got\n%swant\n%s", got, want)
	}
}
