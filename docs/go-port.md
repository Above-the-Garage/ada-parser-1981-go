# From MLISP to a Go Demonstration, 2026

*How Stephen and Claude took the 1981 parser from tape to a working Go program over one afternoon, November 21, 2026.*

The 1981 parser had been recovered from a PDP-10 backup tape and published, byte for byte, in [ada-parser-1981](https://github.com/DrStephenCW/ada-parser-1981). The goal here was to make it run again, on a modern machine, without losing what made it interesting: a grammar held as Lisp data, interpreted at run time by about 130 lines of backtracking recursive descent. (For what the parser led to in 1981, see [the ADARPE story](adarpe-story-1981.md).)

The work went in five steps, each driven by a question or request from Stephen.

## 1. A faithful transliteration

The plan was agreed before any code was written:

- **Read GRAMMA.LSP directly.** The Go program does not translate the grammar into Go. A small Lisp reader ([`lisp/`](../lisp/lisp.go)) handles atoms, lists and UCI Lisp's `/` escape, and [`grammar.go`](../grammar.go) loads the `(DEFPROP rule body GRAMMAR)` forms into a map, the way `GET(name,'GRAMMAR)` found them on property lists.
- **Port PARSER.MLI function by function.** `PARSE`, `PARSE1`, `TPRINT` and `TOP` became `parse`, `parse1`, `TPrint` and the command loop in `cmd/adaparse`. The comments quote the MLISP line each piece comes from.
- **Backtracking is a saved slice index.** In 1981 the input was a pointer into an immutable list, saved on entry and restored on failure. In Go it is an `int` into a `[]lisp.Atom`, and NIL is a nil `lisp.List`, so the Go reads much like the Lisp.
- **Keep the 1981 quirks by default.** The off-by-two `CDDR` alternation, the "below `A` means punctuation" test that makes `|` a rule name, and the duplicate rules all behave as they did. `-fix` corrects the two parser bugs.
- **Test against the 1981 session.** The tests read `original/DSKLOG.LOG`, the transcript from January 19, 1981, replay its three parses, and compare the output with the 1981 trees character for character. They matched on the first run, in both modes.

One puzzle came out of the tests. Run against GRAMMA.LSP, those parses print "Can't find production for TASK-DECLARACTION", but the 1981 log shows no such message. Every grammar snapshot on the tape has the typo, so the January 19 build probably didn't yet print it. That is a guess.

## 2. A lexer and file input

Stephen asked for a real Ada program to try. The 1981 parser couldn't read one: it had no scanner (the input was a Lisp list of atoms typed by hand), and its lexical rules were stubs (`IDENTIFIER` is `A ! B ! C`).

So [`lex/`](../lex/lex.go) is an Ada 83 scanner: reserved words, identifiers, decimal and based numbers, strings, character literals (told apart from the attribute tick), compound delimiters and comments. With scanned input, the four stub rules match any token of the right kind. `adaparse FILE` parses a file as a `COMPILATION`, `-e` parses a string, and the old token-list mode moved behind `-tokens`. On a failure, the parser reports the line and column of the furthest token any attempt reached.

That was enough for small programs. Bigger ones showed that the 1981 grammar itself had gaps: a doubled `;` in declarative parts, no string literals in expressions, no `;` after a procedure call, and no constrained arrays.

## 3. A corrected grammar

[`grammar/ada83.lsp`](../grammar/ada83.lsp) is a new grammar, written in the same `DEFPROP` notation and read by the same parser, with `-ada83`. GRAMMA.LSP stays untouched as the 1981 record. The new grammar keeps the 1981 rule names where they still fit, and covers most of Ada 83: generics, tasks and `select`, variant records, aggregates, attributes, allocators, exception handlers and representation clauses. It is still a PEG, so longer alternatives come first.

The example programs in [`examples/ada83/`](../examples/ada83/) (a generic queue with a client, variant records with allocators and aggregates, and a task with `select` and `accept`) all parse completely. A test checks that the new grammar refers to no undefined rule. Run on GRAMMA.LSP, the same check found seven, three more than were known.

## 4. Small things

- **`.ada`, not `.adb`.** Stephen asked why the examples used `.adb`. That's GNAT's convention from the 1990s. Ada 83 compilers used `.ada` (DEC Ada, Alsys) or `.a` (Verdix), so the examples were renamed to `.ada`.
- **Source interleaved with the tree.** Stephen wondered whether the tokens could be shown alongside the parse tree. `-source` prints each source line just before the outermost node that starts on it. This works because the tree keeps the 1981 shape: an atom followed by a list is a rule name, and every other atom is a token, in input order.

## 5. Windows build scripts

Stephen added three scripts for building and running on Windows:

| Script | What it runs |
|---|---|
| [`BuildAndRunFiles.cmd`](../BuildAndRunFiles.cmd) | `adaparse -ada83 examples\gcd.ada` |
| [`BuildAndRunSource.cmd`](../BuildAndRunSource.cmd) | `adaparse -ada83 -source examples/ada83/shapes.ada` |
| [`BuildAndRunTokenTest.cmd`](../BuildAndRunTokenTest.cmd) | `adaparse -tokens -rule SIMPLE-EXPRESSION A + B * C`, the 1981 style |

## What it is now

A demonstration, not a compiler. It parses and prints trees: there is no semantic analysis and no code generation. Its interest is historical. It runs the 1981 design, nearly unchanged, on real Ada, and it shows that a weekend's grammar-as-data idea can be scaled to most of the language by changing only the data.
