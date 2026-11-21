// Command adaparse runs the 1981 Ada parser.
//
// With -rule, it parses the remaining arguments as one token list:
//
//	adaparse -rule SIMPLE-EXPRESSION A + B '*' C
//
// Without -rule, it starts the little command loop from EXPR TOP() in
// PARSER.MLI, prompting with "-> ":
//
//	S tok tok ...   set the input (OLD _ CDR(X))
//	RULE-NAME       reset the input and print its parse as RULE-NAME
//	E               exit (in 1981, back to Lisp)
//
// Tokens are read with UCI Lisp syntax, so write a comma as "/," and a
// parenthesis as "/(" just as the 1981 session did. Input is upper-cased.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	adaparser "github.com/Above-the-Garage/ada-parser-1981-go"
	"github.com/Above-the-Garage/ada-parser-1981-go/lisp"
)

// defaultInput is the test program BUILD.MIC left in OLD:
// (SETQ OLD @(PROCEDURE A IS B : C ; BEGIN NULL ; END ; ))
var defaultInput = []string{"PROCEDURE", "A", "IS", "B", ":", "C", ";", "BEGIN", "NULL", ";", "END", ";"}

func main() {
	grammarPath := flag.String("grammar", "", "grammar file of DEFPROP forms (default: the embedded 1981 GRAMMA.LSP)")
	fixed := flag.Bool("fix", false, "correct the 1981 alternation and punctuation quirks")
	rule := flag.String("rule", "", "parse the arguments as this rule and exit")
	quiet := flag.Bool("q", false, "hide \"Can't find production\" messages")
	flag.Parse()

	g, err := loadGrammar(*grammarPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "adaparse:", err)
		os.Exit(1)
	}
	p := adaparser.New(g)
	p.Fixed = *fixed
	if !*quiet {
		p.Diag = os.Stderr
	}
	if *fixed && len(g.Duplicates) > 0 {
		fmt.Fprintf(os.Stderr, "adaparse: rules defined more than once (the last one wins): %s\n",
			strings.Join(g.Duplicates, ", "))
	}

	if *rule != "" {
		tokens, err := readTokens(strings.Join(flag.Args(), " "))
		if err != nil {
			fmt.Fprintln(os.Stderr, "adaparse:", err)
			os.Exit(1)
		}
		if !parseAndPrint(p, strings.ToUpper(*rule), tokens) {
			os.Exit(1)
		}
		return
	}
	top(p, os.Stdin, os.Stdout)
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
