// Command adaparse runs the 1981 Ada parser.
//
// Parse Ada source files, scanned by a real Ada lexer:
//
//	adaparse hello.ada               # as a COMPILATION
//	adaparse -rule STATEMENT -e 'Total := Total + I * I;'
//
// Parse a 1981-style token list, written in UCI Lisp syntax and matched
// against the grammar's stub lexical rules (identifiers A, B and C):
//
//	adaparse -tokens -rule SIMPLE-EXPRESSION A + B '*' C
//
// With no arguments, it starts the little command loop from EXPR TOP() in
// PARSER.MLI, prompting with "-> ":
//
//	S tok tok ...   set the input (OLD _ CDR(X))
//	RULE-NAME       reset the input and print its parse as RULE-NAME
//	E               exit (in 1981, back to Lisp)
//
// The command loop reads tokens with UCI Lisp syntax, so write a comma as
// "/," and a parenthesis as "/(" just as the 1981 session did. Its input is
// upper-cased.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	adaparser "github.com/Above-the-Garage/ada-parser-1981-go"
	"github.com/Above-the-Garage/ada-parser-1981-go/lex"
	"github.com/Above-the-Garage/ada-parser-1981-go/lisp"
)

// defaultInput is the test program BUILD.MIC left in OLD:
// (SETQ OLD @(PROCEDURE A IS B : C ; BEGIN NULL ; END ; ))
var defaultInput = []string{"PROCEDURE", "A", "IS", "B", ":", "C", ";", "BEGIN", "NULL", ";", "END", ";"}

func usage() {
	fmt.Fprintf(os.Stderr, `usage:
  adaparse [flags] FILE...          parse Ada source files
  adaparse [flags] -e SOURCE        parse Ada source given as an argument
  adaparse [flags] -tokens TOKEN... parse a 1981-style Lisp token list
  adaparse [flags]                  the 1981 command loop

flags:
`)
	flag.PrintDefaults()
}

func main() {
	grammarPath := flag.String("grammar", "", "grammar file of DEFPROP forms (default: the embedded 1981 GRAMMA.LSP)")
	fixed := flag.Bool("fix", false, "correct the 1981 alternation and punctuation quirks")
	rule := flag.String("rule", "", "rule to parse as (default COMPILATION for source, required with -tokens)")
	source := flag.String("e", "", "parse this Ada source text")
	tokens := flag.Bool("tokens", false, "the arguments are a Lisp token list, not file names")
	ada83 := flag.Bool("ada83", false, "use the corrected 2026 grammar (implies -fix)")
	withSource := flag.Bool("source", false, "interleave the source lines with the tree")
	quiet := flag.Bool("q", false, "hide \"Can't find production\" messages")
	flag.Usage = usage
	flag.Parse()
	showSource = *withSource

	var g *adaparser.Grammar
	var err error
	if *ada83 {
		if *grammarPath != "" {
			fatal(fmt.Errorf("use -ada83 or -grammar, not both"))
		}
		g, *fixed = adaparser.Ada83Grammar(), true
	} else if g, err = loadGrammar(*grammarPath); err != nil {
		fatal(err)
	}
	p := adaparser.New(g)
	p.Fixed = *fixed
	if !*quiet {
		p.Diag = &onceWriter{w: os.Stderr, seen: map[string]bool{}}
	}
	if *fixed && len(g.Duplicates) > 0 {
		fmt.Fprintf(os.Stderr, "adaparse: rules defined more than once (the last one wins): %s\n",
			strings.Join(g.Duplicates, ", "))
	}

	switch {
	case *tokens:
		if *rule == "" {
			fatal(fmt.Errorf("-tokens needs -rule"))
		}
		toks, err := readTokens(strings.Join(flag.Args(), " "))
		if err != nil {
			fatal(err)
		}
		if !parseAndPrint(p, strings.ToUpper(*rule), toks) {
			os.Exit(1)
		}
	case *source != "":
		if !parseSource(p, ruleOr(*rule), "-e", *source) {
			os.Exit(1)
		}
	case flag.NArg() > 0:
		ok := true
		for _, name := range flag.Args() {
			src, err := os.ReadFile(name)
			if err != nil {
				fatal(err)
			}
			if flag.NArg() > 1 {
				fmt.Printf("==> %s <==\n", name)
			}
			ok = parseSource(p, ruleOr(*rule), name, string(src)) && ok
		}
		if !ok {
			os.Exit(1)
		}
	default:
		top(p, os.Stdin, os.Stdout)
	}
}

// showSource is the -source flag.
var showSource bool

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "adaparse:", err)
	os.Exit(1)
}

func ruleOr(rule string) string {
	if rule == "" {
		return "COMPILATION"
	}
	return strings.ToUpper(rule)
}

// onceWriter passes each distinct line through once, so a parse that
// backtracks into the same missing production a thousand times says so
// only once.
type onceWriter struct {
	w    io.Writer
	seen map[string]bool
}

func (o *onceWriter) Write(b []byte) (int, error) {
	if s := string(b); !o.seen[s] {
		o.seen[s] = true
		if _, err := o.w.Write(b); err != nil {
			return 0, err
		}
	}
	return len(b), nil
}

// parseSource scans and parses Ada source, prints the tree, and reports
// where the parse stopped if it did not reach the end.
func parseSource(p *adaparser.Parser, rule, name, src string) bool {
	toks, err := lex.Scan(src)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s:%v\n", name, err)
		return false
	}
	tree := p.ParseTokens(rule, toks)
	rest := len(p.Remaining())
	if tree != nil && rest < len(toks) {
		if showSource {
			adaparser.TPrintSource(os.Stdout, tree, toks, src)
		} else {
			adaparser.TPrintText(os.Stdout, tree)
		}
	}
	if tree != nil && rest == 0 {
		return true
	}
	parsed := 0
	if tree != nil {
		parsed = len(toks) - rest
	}
	// Point at the token just past the deepest progress any attempt made.
	if i := p.Furthest(); i < len(toks) {
		t := toks[i]
		fmt.Fprintf(os.Stderr, "%s:%d:%d: syntax error at %s %q (parsed %d of %d tokens as %s)\n",
			name, t.Line, t.Col, t.Kind, t.Text, parsed, len(toks), rule)
	} else {
		fmt.Fprintf(os.Stderr, "%s: unexpected end of input (parsed %d of %d tokens as %s)\n",
			name, parsed, len(toks), rule)
	}
	return false
}

func loadGrammar(path string) (*adaparser.Grammar, error) {
	if path == "" {
		return adaparser.DefaultGrammar(), nil
	}
	src, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return adaparser.LoadGrammar(string(src))
}

// readTokens reads a line of input the way LINEREAD did, as Lisp atoms.
func readTokens(line string) ([]string, error) {
	vals, err := lisp.ReadAll(strings.ToUpper(line))
	if err != nil {
		return nil, err
	}
	var out []string
	for _, v := range vals {
		a, ok := v.(lisp.Atom)
		if !ok {
			return nil, fmt.Errorf("token %s is a list, not an atom", lisp.String(v))
		}
		out = append(out, string(a))
	}
	return out, nil
}

func parseAndPrint(p *adaparser.Parser, rule string, tokens []string) bool {
	tree := p.Parse(rule, tokens)
	if tree == nil {
		fmt.Println("NIL")
		return false
	}
	adaparser.TPrint(os.Stdout, tree)
	if rest := p.Remaining(); len(rest) > 0 {
		fmt.Printf("(unparsed: %s)\n", strings.Join(rest, " "))
	}
	return true
}

// top is EXPR TOP().
func top(p *adaparser.Parser, in io.Reader, out io.Writer) {
	old := defaultInput
	sc := bufio.NewScanner(in)
	for {
		fmt.Fprint(out, "-> ") // 1; MSG(T,"-> ")
		if !sc.Scan() {        // X _ LINEREAD()
			fmt.Fprintln(out)
			return
		}
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "(") {
			// ELSE SPRINT(EVAL(X[1])): there is no Lisp underneath this port.
			fmt.Fprintln(out, "Lisp expressions are not evaluated in the Go port.")
			continue
		}
		x, err := readTokens(line)
		if err != nil {
			fmt.Fprintln(out, err)
			continue
		}
		switch x[0] {
		case "S": // OLD _ CDR(X), set the input
			old = x[1:]
		case "E": // PROMPT(42) ALSO RETURN NIL, exit to Lisp
			return
		default: // RESET() ALSO TPRINT(PARSE(X[1]))
			parseAndPrint(p, x[0], old)
		}
	}
}
