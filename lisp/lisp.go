// Package lisp is a tiny reader and printer for the subset of UCI Lisp
// syntax that appears in GRAMMA.LSP and in the token lists typed at the
// 1981 parser: atoms, parenthesized lists, and the slash escape.
//
// UCI Lisp uses "/" as its escape character, so "/(" is the atom "(" and
// not the start of a list, "/," is the atom ",", and "//=" is the atom "/=".
// Nothing else is special: ";" and ":=" and "}" are ordinary atoms.
package lisp

import (
	"errors"
	"fmt"
	"io"
	"strings"
)

// Value is a Lisp value: either an Atom or a List.
type Value interface{ isValue() }

// Atom is a Lisp atom, stored by its print name with escapes removed.
type Atom string

// List is a Lisp list. A nil List plays the part of NIL.
type List []Value

func (Atom) isValue() {}
func (List) isValue() {}

// IsAtom reports whether v is the atom named name.
func IsAtom(v Value, name string) bool {
	a, ok := v.(Atom)
	return ok && string(a) == name
}

// Reader reads Lisp values one at a time from a string.
type Reader struct {
	src string
	pos int
}

// NewReader returns a Reader over src.
func NewReader(src string) *Reader { return &Reader{src: src} }

// ErrUnbalanced is returned when a ")" has no matching "(" or the input
// ends inside a list.
var ErrUnbalanced = errors.New("lisp: unbalanced parentheses")

// Next returns the next top-level value, or io.EOF when none remain.
func (r *Reader) Next() (Value, error) {
	r.skipSpace()
	if r.pos >= len(r.src) {
		return nil, io.EOF
	}
	if r.src[r.pos] == ')' {
		return nil, fmt.Errorf("%w: unexpected ) at offset %d", ErrUnbalanced, r.pos)
	}
	return r.read()
}

// ReadAll reads every value in src.
func ReadAll(src string) ([]Value, error) {
	r := NewReader(src)
	var out []Value
	for {
		v, err := r.Next()
		if err == io.EOF {
			return out, nil
		}
		if err != nil {
			return out, err
		}
		out = append(out, v)
	}
}

func isSpace(c byte) bool {
	switch c {
	case ' ', '\t', '\r', '\n', '\f', '\v', 0:
		return true
	}
	return false
}

func (r *Reader) skipSpace() {
	for r.pos < len(r.src) && isSpace(r.src[r.pos]) {
		r.pos++
	}
}

func (r *Reader) read() (Value, error) {
	r.skipSpace()
	if r.pos >= len(r.src) {
		return nil, fmt.Errorf("%w: input ends inside a list", ErrUnbalanced)
	}
	if r.src[r.pos] == '(' {
		r.pos++
		list := List{}
		for {
			r.skipSpace()
			if r.pos >= len(r.src) {
				return nil, fmt.Errorf("%w: input ends inside a list", ErrUnbalanced)
			}
			if r.src[r.pos] == ')' {
				r.pos++
				return list, nil
			}
			v, err := r.read()
			if err != nil {
				return nil, err
			}
			list = append(list, v)
		}
	}
	return r.readAtom(), nil
}

func (r *Reader) readAtom() Atom {
	var b strings.Builder
	for r.pos < len(r.src) {
		c := r.src[r.pos]
		if c == '/' && r.pos+1 < len(r.src) {
			b.WriteByte(r.src[r.pos+1])
			r.pos += 2
			continue
		}
		if isSpace(c) || c == '(' || c == ')' {
			break
		}
		b.WriteByte(c)
		r.pos++
	}
	return Atom(b.String())
}

// needsEscape reports whether c must be written with a slash so that the
// atom reads back the same way. The set follows what the UCI Lisp printer
// wrote into GRAMMA.LSP and DSKLOG.LOG ("/(", "/,", "/.", "/'", "/[").
func needsEscape(c byte) bool {
	switch c {
	case '/', '(', ')', '[', ']', '{', ',', '.', '\'', ' ', '\t':
		return true
	}
	return false
}

// String returns v in UCI Lisp print syntax.
func String(v Value) string {
	var b strings.Builder
	write(&b, v)
	return b.String()
}

func write(b *strings.Builder, v Value) {
	switch x := v.(type) {
	case Atom:
		if x == "-" { // a lone "-" would read as a number in UCI Lisp
			b.WriteString("/-")
			return
		}
		for i := 0; i < len(x); i++ {
			if needsEscape(x[i]) {
				b.WriteByte('/')
			}
			b.WriteByte(x[i])
		}
	case List:
		b.WriteByte('(')
		for i, e := range x {
			if i > 0 {
				b.WriteByte(' ')
			}
			write(b, e)
		}
		b.WriteByte(')')
	case nil:
		b.WriteString("NIL")
	}
}
