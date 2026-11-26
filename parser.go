package adaparser

import (
	"fmt"
	"io"

	"github.com/DrStephenCW/ada-parser-1981-go/lex"
	"github.com/DrStephenCW/ada-parser-1981-go/lisp"
)

// PlaceHolder is the marker PARSE1 returns when an optional part (/[ ... /])
// or a repetition (/{ ... }) matches nothing. TPRINT skips it.
const PlaceHolder = lisp.Atom("PLACE-HOLDER")

// Atoms that mean something inside a rule body. The MLISP source writes
// them with its own escape character: '?^, '?[, '?{ and '!.
const (
	atomOr       = "!" // alternation: X ! Y ! Z
	atomReserved = "^" // (^ WORD), a reserved word
	atomOptional = "[" // (/[ ... /]), zero or one
	atomRepeat   = "{" // (/{ ... }), zero or more
)

// Parser is one run of the 1981 parser over a grammar.
//
// The zero value is not usable; call New.
type Parser struct {
	g *Grammar

	// Fixed turns on corrections for the 1981 quirks. When false (the
	// default) the parser behaves exactly as PARSER.MLI did:
	//
	//   - Alternation steps through a rule body two elements at a time
	//     (CDDR), so an alternative written without parentheses, like
	//     NAME RANGE-CONSTRAINT in CHOICE, puts the rest out of step.
	//     Fixed splits the body on "!" instead, and an alternative of
	//     several elements is parsed as a sequence.
	//   - An atom is punctuation when its first character sorts below "A"
	//     (ASCII 65), so "|", "^", "{" and "}" are looked up as rule names.
	//     Fixed treats any atom that does not start with a letter as
	//     punctuation.
	Fixed bool

	// Diag receives the messages the parser printed with MSG, such as
	// "Can't find production for IDENTIFIR". Nil discards them.
	Diag io.Writer

	// input is INPUT-STREAM: the remaining tokens are input[pos:].
	// Saving and restoring pos is the whole backtracking mechanism, just
	// as SAVED-INPUT-STREAM _ INPUT-STREAM was in 1981.
	input []lisp.Atom
	pos   int

	// toks is set by ParseTokens: the scanned source behind input, used
	// to match the lexical rules by kind and to report positions.
	toks []lex.Token

	// furthest is the most tokens any attempt managed to consume.
	furthest int
}

// lexicalRules are the grammar rules that GRAMMA.LSP only stubs out
// (IDENTIFIER is A ! B ! C). When the input comes from the scanner, each
// one matches any single token of its kind instead.
var lexicalRules = map[lisp.Atom]lex.Kind{
	"IDENTIFIER":        lex.Identifier,
	"NUMERIC-LITERAL":   lex.Number,
	"CHARACTER-LITERAL": lex.Character,
	"CHARACTER-STRING":  lex.String,
}

// New returns a parser for g that behaves as the 1981 original did.
func New(g *Grammar) *Parser { return &Parser{g: g} }

// Parse parses tokens as the named rule and returns the parse tree, or nil
// if the rule does not match. Like the original, it does not require the
// rule to use up all of the input; Remaining reports what is left.
//
// The tree has the shape PARSE built: (RULENAME (child child ...)), where
// each child is either a matched token or another (RULENAME (...)) pair
// spliced into its parent's list. Print it with TPrint.
func (p *Parser) Parse(rule string, tokens []string) lisp.List {
	p.input = make([]lisp.Atom, len(tokens))
	for i, t := range tokens {
		p.input[i] = lisp.Atom(t)
	}
	p.pos = 0
	p.furthest = 0
	p.toks = nil
	return p.parse(lisp.Atom(rule))
}

// ParseTokens parses scanned Ada source as the named rule. The stubbed
// lexical rules (IDENTIFIER, NUMERIC-LITERAL, CHARACTER-LITERAL and
// CHARACTER-STRING) match real tokens of the right kind; everything else
// works exactly as in Parse.
func (p *Parser) ParseTokens(rule string, toks []lex.Token) lisp.List {
	p.input = make([]lisp.Atom, len(toks))
	for i, t := range toks {
		p.input[i] = lisp.Atom(t.Text)
	}
	p.pos, p.furthest, p.toks = 0, 0, toks
	return p.parse(lisp.Atom(rule))
}

// Furthest returns how many tokens the deepest attempt of the last parse
// consumed before it backtracked. When a parse fails or stops early,
// the token at that index is usually where the trouble is.
func (p *Parser) Furthest() int { return p.furthest }

// isLexical reports whether the next token is a scanned identifier or
// literal, which only the lexical rules may match.
func (p *Parser) isLexical() bool {
	if p.toks == nil || p.pos >= len(p.toks) {
		return false
	}
	k := p.toks[p.pos].Kind
	return k != lex.Delimiter && k != lex.Reserved
}

// consume is POP(INPUT-STREAM).
func (p *Parser) consume() {
	p.pos++
	if p.pos > p.furthest {
		p.furthest = p.pos
	}
}

// Remaining returns the tokens the last Parse did not consume.
func (p *Parser) Remaining() []string {
	out := make([]string, 0, len(p.input)-p.pos)
	for _, a := range p.input[p.pos:] {
		out = append(out, string(a))
	}
	return out
}

// peek returns INPUT-STREAM[1], or "" (standing in for NIL) at the end.
func (p *Parser) peek() lisp.Atom {
	if p.pos < len(p.input) {
		return p.input[p.pos]
	}
	return ""
}

// isPunct is the test CAR(AEXPLODEC(RULENAME)) LT 65.
func (p *Parser) isPunct(a lisp.Atom) bool {
	if a == "" {
		return false
	}
	c := a[0]
	if p.Fixed {
		return !('A' <= c && c <= 'Z' || 'a' <= c && c <= 'z')
	}
	return c < 'A'
}

func (p *Parser) msg(format string, args ...any) {
	if p.Diag != nil {
		fmt.Fprintf(p.Diag, format, args...)
	}
}

// parse is EXPR PARSE(RULENAME). It is called with a rule name, not a rule.
func (p *Parser) parse(ruleName lisp.Atom) lisp.List {
	saved := p.pos // SAVED-INPUT-STREAM _ INPUT-STREAM

	// IF RULENAME EQ NIL OR INPUT-STREAM EQ NIL THEN RETURN NIL
	if ruleName == "" || p.pos >= len(p.input) {
		return nil
	}

	// Not in 1981: with scanned input, a lexical rule matches by token kind.
	if kind, ok := lexicalRules[ruleName]; ok && p.toks != nil {
		if p.toks[p.pos].Kind == kind {
			tok := p.input[p.pos]
			p.consume()
			return lisp.List{ruleName, lisp.List{tok}}
		}
		return nil
	}

	// IF CAR(AEXPLODEC(RULENAME)) LT 65 THEN (IF RULENAME EQ INPUT-STREAM[1] ...)
	if p.isPunct(ruleName) {
		if ruleName == p.peek() {
			p.consume() // POP(INPUT-STREAM)
			return lisp.List{ruleName}
		}
		return nil
	}

	// RULE _ GET(RULENAME, 'GRAMMAR)
	rule, ok := p.g.rules[ruleName]
	if !ok {
		p.msg("Can't find production for %s\n", ruleName)
		return nil
	}

	// IF RULE[2] EQ '! THEN ... handle alternation
	if alts, ok := rule.(lisp.List); ok && len(alts) >= 2 && lisp.IsAtom(alts[1], atomOr) {
		if p.Fixed {
			return p.parseAlternatives(ruleName, alts, saved)
		}
		for {
			// 1; IF ATOM(RULE[1]) THEN X _ PARSE(RULE[1]) ELSE X _ PARSE1(RULE[1])
			x := p.parseItem(alts[0])
			if x != nil {
				// RETURN <RULENAME, IF ATOM X THEN <X> ELSE X>
				// (X is always a list here, so the ATOM test never fires.)
				return lisp.List{ruleName, x}
			}
			// ELSE IF (RULE _ CDDR RULE) EQ NIL THEN restore and RETURN NIL
			if len(alts) <= 2 {
				p.pos = saved
				return nil
			}
			alts = alts[2:] // ELSE GO 1
		}
	}

	// ELSE RETURN IF X _ PARSE1(RULE) THEN <RULENAME, X> ELSE restore, NIL
	if x := p.parse1(rule); x != nil {
		return lisp.List{ruleName, x}
	}
	p.pos = saved
	return nil
}

// parseAlternatives is the Fixed-mode replacement for the CDDR loop in
// PARSE: it splits the body on "!" so that an alternative may be written
// as several elements without parentheses.
func (p *Parser) parseAlternatives(ruleName lisp.Atom, body lisp.List, saved int) lisp.List {
	start := 0
	for i := 0; i <= len(body); i++ {
		if i < len(body) && !lisp.IsAtom(body[i], atomOr) {
			continue
		}
		alt := body[start:i]
		start = i + 1
		var x lisp.List
		switch len(alt) {
		case 0:
			continue
		case 1:
			x = p.parseItem(alt[0])
		default:
			x = p.parse1(alt)
		}
		if x != nil {
			return lisp.List{ruleName, x}
		}
	}
	p.pos = saved
	return nil
}

// parseItem is the idiom IF ATOM I THEN PARSE(I) ELSE PARSE1(I), used for
// each alternative in PARSE and each element of a sequence in PARSE1.
func (p *Parser) parseItem(v lisp.Value) lisp.List {
	if a, ok := v.(lisp.Atom); ok {
		return p.parse(a)
	}
	return p.parse1(v)
}

// inner is PRELIST(CDR RULE, LENGTH(RULE)-2): the body of an optional or
// repeated part without its opening and closing markers. The closing "/]"
// or "}" is dropped without being checked, as in 1981.
func inner(rule lisp.List) lisp.List {
	if len(rule) < 2 {
		return nil
	}
	return rule[1 : len(rule)-1]
}

// parse1 is EXPR PARSE1(RULE). It is passed the actual rule, or a piece of
// one, rather than a rule name.
func (p *Parser) parse1(rule lisp.Value) lisp.List {
	saved := p.pos // SAVED-INPUT-STREAM _ INPUT-STREAM

	switch r := rule.(type) {
	case lisp.Atom:
		// Either another rule or punctuation.
		// IF RULE EQ INPUT-STREAM[1]: exact match. (With scanned input,
		// an identifier spelled like a rule name must not count.)
		if r == p.peek() && !p.isLexical() {
			p.consume()
			return lisp.List{r}
		}
		if p.isPunct(r) {
			return nil
		}
		return p.parse(r) // or rule name

	case lisp.List:
		if len(r) == 0 { // IF RULE EQ NIL THEN RETURN NIL
			return nil
		}
		switch {
		case lisp.IsAtom(r[0], atomReserved): // IS IT A RESERVED WORD?
			// IF RULE[2] EQ INPUT-STREAM[1] THEN POP(INPUT-STREAM) ALSO RETURN <RULE[2]>
			if tok := p.peek(); tok != "" && len(r) > 1 && lisp.IsAtom(r[1], string(tok)) {
				p.consume()
				return lisp.List{r[1]}
			}
			return nil

		case lisp.IsAtom(r[0], atomOptional): // OPTIONAL PART
			// RETURN IF PARSE1(...) THENTHAT ELSE <'PLACE-HOLDER>
			if x := p.parse1(inner(r)); x != nil {
				return x
			}
			return lisp.List{PlaceHolder}

		case lisp.IsAtom(r[0], atomRepeat): // ZERO OR MORE PART
			// WHILE X _ PARSE1(...) DO: RESULT _ RESULT @ X
			var result lisp.List
			for {
				before := p.pos
				x := p.parse1(inner(r))
				if x == nil {
					break
				}
				result = append(result, x...)
				// The one guard the original lacked: a repetition whose
				// body matches without consuming input would loop forever.
				if p.pos == before {
					break
				}
			}
			if result != nil { // RETURN IF RESULT THENTHAT ELSE <'PLACE-HOLDER>
				return result
			}
			return lisp.List{PlaceHolder}

		default: // MUST BE A LIST OF RULES, ALL MUST BE TRUE
			var result lisp.List
			for _, item := range r {
				y := p.parseItem(item)
				if y == nil { // IF Y = NIL THEN restore and RETURN NIL
					p.pos = saved
					return nil
				}
				result = append(result, y...) // RESULT _ RESULT @ Y
			}
			return result
		}
	}
	return nil
}
