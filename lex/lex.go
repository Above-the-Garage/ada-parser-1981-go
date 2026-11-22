// Package lex is an Ada 83 scanner. The 1981 parser had none: its input was
// a Lisp list of atoms typed by hand. This package turns real Ada source
// into tokens the parser can match against GRAMMA.LSP.
//
// Reserved words are upper-cased, since the grammar spells them that way
// in (^ WORD) forms. Identifiers keep the spelling they had in the source;
// Ada ignores case, and the parser matches them by kind, not by text.
package lex

import (
	"fmt"
	"strings"
)

// Kind is the lexical class of a token.
type Kind int

const (
	Delimiter Kind = iota // ; := ( ) => .. and the rest
	Reserved              // BEGIN, PROCEDURE, ...
	Identifier
	Number    // 42, 3.14, 16#FF#, 1.0E-6
	String    // "text", kept with its quotes
	Character // 'x', kept with its quotes
)

func (k Kind) String() string {
	switch k {
	case Delimiter:
		return "delimiter"
	case Reserved:
		return "reserved word"
	case Identifier:
		return "identifier"
	case Number:
		return "number"
	case String:
		return "string"
	case Character:
		return "character literal"
	}
	return fmt.Sprintf("Kind(%d)", int(k))
}

// Token is one lexical element and where it starts (1-based).
type Token struct {
	Kind Kind
	Text string
	Line int
	Col  int
}

// reserved holds the 63 reserved words of Ada 83 (ANSI/MIL-STD-1815A).
var reserved = map[string]bool{}

func init() {
	for _, w := range strings.Fields(`
		ABORT ABS ACCEPT ACCESS ALL AND ARRAY AT BEGIN BODY CASE CONSTANT
		DECLARE DELAY DELTA DIGITS DO ELSE ELSIF END ENTRY EXCEPTION EXIT
		FOR FUNCTION GENERIC GOTO IF IN IS LIMITED LOOP MOD NEW NOT NULL
		OF OR OTHERS OUT PACKAGE PRAGMA PRIVATE PROCEDURE RAISE RANGE
		RECORD REM RENAMES RETURN REVERSE SELECT SEPARATE SUBTYPE TASK
		TERMINATE THEN TYPE USE WHEN WHILE WITH XOR`) {
		reserved[w] = true
	}
}

// IsReserved reports whether word (in any case) is an Ada 83 reserved word.
func IsReserved(word string) bool { return reserved[strings.ToUpper(word)] }

var compound = []string{"=>", "..", "**", ":=", "/=", ">=", "<=", "<<", ">>", "<>"}

const simple = "&'()*+,-./:;<=>|"

// Error is a scanning error at a position in the source.
type Error struct {
	Line, Col int
	Msg       string
}

func (e *Error) Error() string { return fmt.Sprintf("%d:%d: %s", e.Line, e.Col, e.Msg) }

// Scan splits Ada source into tokens. Comments (-- to end of line) and
// white space are dropped.
func Scan(src string) ([]Token, error) {
	s := &scanner{src: src, line: 1, col: 1}
	var toks []Token
	for {
		s.skipSpaceAndComments()
		if s.pos >= len(s.src) {
			return toks, nil
		}
		t, err := s.next(toks)
		if err != nil {
			return toks, err
		}
		toks = append(toks, t)
	}
}

type scanner struct {
	src       string
	pos       int
	line, col int
}

func (s *scanner) peekAt(off int) byte {
	if s.pos+off < len(s.src) {
		return s.src[s.pos+off]
	}
	return 0
}

func (s *scanner) advance(n int) {
	for i := 0; i < n && s.pos < len(s.src); i++ {
		if s.src[s.pos] == '\n' {
			s.line++
			s.col = 1
		} else {
			s.col++
		}
		s.pos++
	}
}

func (s *scanner) errorf(format string, args ...any) error {
	return &Error{Line: s.line, Col: s.col, Msg: fmt.Sprintf(format, args...)}
}

func isLetter(c byte) bool { return 'A' <= c && c <= 'Z' || 'a' <= c && c <= 'z' }
func isDigit(c byte) bool  { return '0' <= c && c <= '9' }
func isExtDigit(c byte) bool {
	return isDigit(c) || 'A' <= c && c <= 'F' || 'a' <= c && c <= 'f'
}

func (s *scanner) skipSpaceAndComments() {
	for s.pos < len(s.src) {
		c := s.src[s.pos]
		switch {
		case c == ' ' || c == '\t' || c == '\r' || c == '\n' || c == '\f' || c == '\v':
			s.advance(1)
		case c == '-' && s.peekAt(1) == '-':
			for s.pos < len(s.src) && s.src[s.pos] != '\n' {
				s.advance(1)
			}
		default:
			return
		}
	}
}

func (s *scanner) next(prev []Token) (Token, error) {
	t := Token{Line: s.line, Col: s.col}
	start := s.pos
	c := s.src[s.pos]

	switch {
	case isLetter(c):
		for isLetter(s.peekAt(0)) || isDigit(s.peekAt(0)) || s.peekAt(0) == '_' {
			s.advance(1)
		}
		t.Text = s.src[start:s.pos]
		if IsReserved(t.Text) {
			t.Kind = Reserved
			t.Text = strings.ToUpper(t.Text)
		} else {
			t.Kind = Identifier
		}
		return t, nil

	case isDigit(c):
		t.Kind = Number
		t.Text = s.number()
		return t, nil

	case c == '"':
		s.advance(1)
		for {
			if s.pos >= len(s.src) || s.src[s.pos] == '\n' {
				return t, &Error{Line: t.Line, Col: t.Col, Msg: "unterminated string"}
			}
			if s.src[s.pos] == '"' {
				if s.peekAt(1) == '"' { // "" is a quote inside a string
					s.advance(2)
					continue
				}
				s.advance(1)
				break
			}
			s.advance(1)
		}
		t.Kind = String
		t.Text = s.src[start:s.pos]
		return t, nil

	case c == '\'' && s.peekAt(2) == '\'' && !tickIsAttribute(prev):
		s.advance(3)
		t.Kind = Character
		t.Text = s.src[start:s.pos]
		return t, nil
	}

	for _, d := range compound {
		if strings.HasPrefix(s.src[s.pos:], d) {
			s.advance(len(d))
			t.Kind = Delimiter
			t.Text = d
			return t, nil
		}
	}
	if strings.IndexByte(simple, c) >= 0 {
		s.advance(1)
		t.Kind = Delimiter
		t.Text = string(c)
		return t, nil
	}
	return t, s.errorf("unexpected character %q", c)
}

// tickIsAttribute reports whether an apostrophe following prev is the
// attribute or qualification tick (X'FIRST, T'(...)) rather than the start
// of a character literal. It is, after an identifier, ")" or ALL.
func tickIsAttribute(prev []Token) bool {
	if len(prev) == 0 {
		return false
	}
	p := prev[len(prev)-1]
	return p.Kind == Identifier || p.Text == ")" || p.Text == "ALL"
}

// number scans a decimal or based literal: 42, 1_000, 3.14, 1.0E-6, 16#FF#,
// 2#1010_1010#E2.
func (s *scanner) number() string {
	start := s.pos
	digits := func(ok func(byte) bool) {
		for ok(s.peekAt(0)) || s.peekAt(0) == '_' && ok(s.peekAt(1)) {
			s.advance(1)
		}
	}
	digits(isDigit)
	if s.peekAt(0) == '#' {
		s.advance(1)
		digits(isExtDigit)
		if s.peekAt(0) == '.' && isExtDigit(s.peekAt(1)) {
			s.advance(1)
			digits(isExtDigit)
		}
		if s.peekAt(0) == '#' {
			s.advance(1)
		}
	} else if s.peekAt(0) == '.' && isDigit(s.peekAt(1)) { // not the ".." of a range
		s.advance(1)
		digits(isDigit)
	}
	if c := s.peekAt(0); c == 'E' || c == 'e' {
		off := 1
		if s.peekAt(1) == '+' || s.peekAt(1) == '-' {
			off = 2
		}
		if isDigit(s.peekAt(off)) {
			s.advance(off)
			digits(isDigit)
		}
	}
	return s.src[start:s.pos]
}
