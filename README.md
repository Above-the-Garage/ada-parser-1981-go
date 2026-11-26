# The 1981 Ada Parser, in Go

_This entire project was done with Claude Opus 5.5 Medium Effort guided by Stephen Clarke-Willson_

This is a Go transliteration of a small Ada parser that I wrote in MLISP on the UC Irvine PDP-10 in January 1981. The original, recovered from backup tape in 2026, is at **[DrStephenCW/ada-parser-1981](https://github.com/DrStephenCW/ada-parser-1981)**, along with the story of how it was written and what came of it.

The port keeps the 1981 design. The grammar is not translated into Go: the program reads the same `GRAMMA.LSP` file of `(DEFPROP rule body GRAMMAR)` forms with a tiny Lisp reader and walks it at run time, and the parser functions follow `PARSER.MLI` one for one. By default it also keeps the 1981 bugs.

— Stephen Clarke-Willson

Two companion documents tell the rest of the story:

- [From a Weekend Parser to ADARPE](docs/adarpe-story-1981.md): what the 1981 parser led to, the UCI Ada Rapid Prototyping Environment, pieced together from the tapes.
- [From MLISP to a Go Demonstration](docs/go-port.md): how Stephen and Claude ported it to Go in 2026 and took it to real Ada.

On Windows, the `BuildAndRun*.cmd` scripts build `adaparse` and run it on an example.

## Try it

```
go install github.com/DrStephenCW/ada-parser-1981-go/cmd/adaparse@latest

adaparse examples/sum_squares.ada
```

```
 COMPILATION
   COMPILATION-UNIT
     CONTEXT-SPECIFICATION
     SUBPROGRAM-BODY
       SUBPROGRAM-SPECIFICATION
         PROCEDURE
         IDENTIFIER
           Sum_Squares
       IS
       DECLARATIVE-PART
       ...
```

There are three ways to feed it:

- **Ada source files:** `adaparse FILE...`, or `adaparse -e 'source text'`. A real Ada 83 scanner ([`lex/`](lex/lex.go)) turns the source into tokens. The parse starts at `COMPILATION` unless you give `-rule`, for example `-rule STATEMENT -e 'X := Y + 1;'`. If the parse fails or stops early, it reports the line and column of the furthest token any attempt reached, which is usually where the trouble is.
- **A 1981-style token list:** `adaparse -tokens -rule SIMPLE-EXPRESSION A + B '*' C`. The tokens use UCI Lisp syntax, so a comma is `/,` and a parenthesis is `/(`. With no scanner, the grammar's stub lexical rules apply: the only identifiers are `A`, `B` and `C`, and the only numbers are `1` and `2`.
- **The 1981 command loop:** run `adaparse` with no arguments for `TOP()`. `S tokens...` sets the input, a rule name parses it, and `E` exits.

Add `-source` to interleave the program with the tree. Each source line is printed, flush left with its line number, just before the outermost node that starts on it:

```
   3|    N     : Integer := 10;
         DECLARATIVE-PART
           DECLARATIVE-ITEM
             BASIC-DECLARATION
               OBJECT-DECLARATION
                 ...
   4|    Total : Integer := 0;
           DECLARATIVE-ITEM
```

Other flags: `-ada83` uses the corrected grammar (see below), `-fix` turns on the corrections described below, `-grammar FILE` loads a different grammar (any of the GRAMMA.* snapshots, for instance), and `-q` hides the "Can't find production" messages. Each distinct message is printed only once.

## Layout

| Path | What it is | 1981 counterpart |
|---|---|---|
| [`lisp/`](lisp/lisp.go) | Reader and printer for atoms, lists and the `/` escape | the UCI Lisp reader |
| [`lex/`](lex/lex.go) | Ada 83 scanner | none: in 1981 the token list was typed by hand |
| [`grammar.go`](grammar.go) | Loads `DEFPROP` forms into a map | `DSKIN GRAMMA` and `GET(name,'GRAMMAR)` |
| [`parser.go`](parser.go) | `parse` and `parse1` | `EXPR PARSE` and `EXPR PARSE1` |
| [`tprint.go`](tprint.go) | Tree printer | `EXPR TPRINT` |
| [`cmd/adaparse`](cmd/adaparse/main.go) | Command line and command loop | `EXPR TOP` and `BUILD.MIC` |
| [`grammar/ada83.lsp`](grammar/ada83.lsp) | Corrected 2026 grammar, used with `-ada83` | GRAMMA.LSP |
| [`examples/`](examples/) | Small Ada programs, and larger ones in `examples/ada83/` | |
| [`original/`](original/) | `GRAMMA.LSP`, `PARSER.MLI`, `DSKLOG.LOG`, `BUILD.MIC`, byte-exact from tape | |

The comments in `parser.go` quote the MLISP line each piece of Go comes from.

## How the MLISP maps to Go

- **`INPUT-STREAM` is a slice index.** In 1981 the remaining input was a pointer into an immutable Lisp list, and each function saved it on entry and put it back on failure. Here the tokens are a `[]lisp.Atom` and the position is an `int`. Backtracking is still one assignment.
- **NIL is a nil `lisp.List`.** `parse` and `parse1` return the parse tree, or nil for failure, so the Go reads like the Lisp: `if x := p.parse1(rule); x != nil { ... }`.
- **The tree has the 1981 shape.** `PARSE` returned `(RULENAME X)` and sequences were joined with `@` (append), which splices each child's name and subtree into its parent's list. The Go builds exactly the same nested lists, so `TPrint` can follow `TPRINT` line for line.
- **`MSG` goes to `Parser.Diag`**, an `io.Writer` that is nil (silent) by default.

## Real Ada source

GRAMMA.LSP stubs out the lexical rules. `IDENTIFIER` is `(^ A) ! (^ B) ! (^ C)`, `NUMERIC-LITERAL` is `1 ! 2`, `CHARACTER-LITERAL` is `C-LIT`, and `CHARACTER-STRING` is never defined. When the input comes from the scanner, these four rules match any one token of the right kind instead, and the tree keeps the identifier as it was spelled in the source. Everything else is still the 1981 grammar, interpreted by the 1981 parser.

That is enough for small programs like [`examples/sum_squares.ada`](examples/sum_squares.ada) and [`examples/gcd.ada`](examples/gcd.ada), which parse completely. Larger ones run into gaps in the grammar itself:

- **A doubled `;` in declarative parts.** `DECLARATIVE-PART` is `(/{ UNIVERSAL-DECLARATION ; })`, but most declaration rules already end in `;`. Only object declarations (`X : T := V`) leave the `;` to `DECLARATIVE-PART`, so `Max : constant := 100;` or a nested procedure body needs a second `;`. [`examples/stack.ada`](examples/stack.ada) stops at line 9 for this reason.
- **No string literals in expressions.** `PRIMARY` has no string alternative, so `Put_Line ("Hello, world")` fails at the string ([`examples/hello.ada`](examples/hello.ada)).
- **Procedure calls lack their `;`.** `PROCEDURE-CALL` is `(NAME)`, so `Put_Line (X);` leaves the `;` for a following statement that never matches it.
- **Constrained arrays.** `array (1 .. Max) of Integer` needs `( ... )` around its constraint, but `CONSTRAINT` uses `UNIVERSAL-EXPRESSION-PACK` without the parentheses.

These are mistakes in the data, not in the parser, so `-fix` leaves them alone. GRAMMA.LSP also refers to seven rules that are never defined: `TASK-DECLARACTION`, `IDENTIFIR`, `ARRAY-TYPE-INDICATION`, `UNVERSIAL-EXPRESSION-PACK`, `ACTUAL-PARAMETER-PART`, `TASK-NAME` and a bare `OTHERS`.

## The corrected grammar (`-ada83`)

[`grammar/ada83.lsp`](grammar/ada83.lsp) is a new grammar, written in 2026 in the same `DEFPROP` notation and read by the same parser. GRAMMA.LSP stays untouched as the 1981 record. Select it with `-ada83`, which also turns on `-fix`:

```
adaparse -ada83 examples/ada83/queues.ada
```

It keeps the 1981 rule names where they still fit, and it covers most of Ada 83:

- context clauses, subprograms, packages, generics and their instantiations, tasks, subunits and body stubs
- object, number, exception, type, subtype and renaming declarations
- enumeration, integer, real, array (constrained and unconstrained), record (with variants), access, derived and private types
- every statement, including labels, named loops, blocks, `select`, `accept` and exception handlers
- full expressions: short-circuit forms, membership tests, string literals, aggregates, attributes, qualified expressions and allocators
- representation clauses and pragmas

Every rule it refers to is defined, and a test checks that. The grammar is still a PEG, so the order of alternatives matters: longer forms come first. Where Ada is genuinely ambiguous without semantics, the grammar picks one reading. For example, `A (I)` is always a `NAME` with a call-or-index tail, and an index constraint such as `String (1 .. 10)` parses the same way. The examples in [`examples/ada83/`](examples/ada83/) (a generic queue, variant records with aggregates and allocators, and a task with `select`) all parse completely, as do the four programs in `examples/`. On a syntax error, `adaparse` gives the line and column of the token it got stuck on.

## The 1981 quirks, and `-fix`

`Parser.Fixed` (the `-fix` flag) corrects the two parser bugs. It is off by default.

| Quirk | 1981 behaviour (default) | With `-fix` |
|---|---|---|
| Alternation steps by two (`CDDR`) | `CHOICE` is `SIMPLE-EXPRESSION ! NAME RANGE-CONSTRAINT ! RANGE ! (^ OTHERS)`. The unparenthesized `NAME RANGE-CONSTRAINT` throws the stepping off, so the loop lands on `!` and never tries `RANGE` or `OTHERS`. | The body is split on `!`, and an alternative of several elements is parsed as a sequence. |
| Punctuation is "first character below `A`" | `\|` (ASCII 124), `^`, `{` and `}` are looked up as rule names and fail with "Can't find production for \|". | Anything that does not start with a letter is punctuation. |
| Duplicate rules | `BODY`, `PROGRAM-COMPONENT` and `CHOICE` are each defined twice. The later one wins. | Same, but `adaparse -fix` lists them. (The copies are identical, so nothing changes.) |

The grammar's own typos (`TASK-DECLARACTION`, `IDENTIFIR`, `ARRAY-TYPE-INDICATION`, `UNVERSIAL-EXPRESSION-PACK`, and `CHARACTER-STRING`, which is never defined) are data, not parser bugs, so `-fix` leaves them alone. Each one fails its alternative and prints a "Can't find production" message.

One change applies in both modes: a repetition `(/{ ... })` whose body matches without consuming any input stops instead of looping forever. No rule in GRAMMA.LSP triggers it.

## Tests

```
go test ./...
```

[`parser_test.go`](parser_test.go) reads `original/DSKLOG.LOG`, the transcript of the session on January 19, 1981, finds the three parses in it (two subprogram bodies and `A + B * C`), runs each one through the Go parser, and checks that the printed tree matches the 1981 output character for character. The only adjustment is expanding the tabs the TOPS-10 terminal driver inserted. All three match in both modes. Other tests pin down the alternation and `|` quirks and the backtracking, check the scanner, parse the example programs, and check that `stack.ada` stops at line 9 with the 1981 grammar. With the corrected grammar they parse every example completely, check where some broken programs fail, and check that no rule is left undefined.

One small puzzle: run against GRAMMA.LSP, the first two parses print "Can't find production for TASK-DECLARACTION" (and one prints the same for `CHARACTER-STRING`), but the 1981 log shows no such lines. Every grammar snapshot on the tape has those typos, so the likeliest answer is that the January 19 build of `PARSER.MLI` (the surviving copy is dated January 22) did not yet print that message. That is a guess.

## License

MIT. See [LICENSE](LICENSE).
