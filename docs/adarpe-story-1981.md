# From a Weekend Parser to ADARPE: the UCI Ada Environment, 1981

*Pieced together from the tapes: the TAS 145 files (Jan 1981), project mail saved in Stephen's
TOPS-20 accounts (May to Sept 1981), and Stephen's later thesis. The source code from the middle of
the story did not survive. What follows is what the surviving files show, with the gaps marked.*

## The people

| Who | Account | Role in the story |
|---|---|---|
| Stephen (Steve) Willson | TS.WILLSON | Wrote the original parser; maintained and rebuilt the shared system; answered everyone's questions |
| Tim Standish | STANDISH, `<ts>` | Taught TAS 145 (Stephen's recollection); faculty lead ("our leader," "loyal leader" in the logicals file); found the disk space and the money |
| Jim Meehan | MEEHAN | Faculty owner of MLISP and its editor (LISPF/MLISPF); added the `ASTMT` extension the system ran on |
| Ray Klefstad | TS.KLEFSTAD | Terminal I/O with command completion and line editing; later co-wrote the Lisp Arcturus and Adash |
| Stephen Whitehill | TS.WHITEHILL | Ada hacker; later co-wrote the Lisp Arcturus and the C rewrite |
| Eilers | TS.EILERS | Lexical and string work (special characters); asked for a top-level `RUN` command |
| Dave Smith | TS.SMITH | Documentation, cross-reference tools, internal-form tree grammar, unparser, Ada macro facility |
| Bannister, Velick, Snider, Long, Taylor | TS.* | Listed project members; their parts aren't described in the surviving mail |
| Phil Ritter | RITTER | Systems support: the tools EXEC, MM mail, the Z editor, mail nicknames |

## January 1981: the parser

It started in TAS 145. As Stephen recalls it, this was Programming Languages 145, a compiler course in which students wrote a compiler, and "TAS" stood for its teacher, Thomas A. ("Tim") Standish. Over the first three weeks of January 1981, Stephen typed
the preliminary Ada grammar into Lisp. His first try (A.MLI, 1 Jan) was a generator that printed
MLISP source, one procedure per production. A hand-made sample (ADALSP.LSP, 2 Jan) showed Ada
written as Lisp `PROG` code with generated labels, a mock-up of what an Ada-to-Lisp translator
might emit. Over the weekend of 10 and 11 January he switched approaches: the grammar became Lisp
data on property lists (READP.MLI, GRAMMA.LSP), and PARSER.MLI interpreted it by backtracking
recursive descent. By 22 January it parsed real Ada fragments. (This is the code in
[ada-parser-1981](https://github.com/DrStephenCW/ada-parser-1981), and the code this repository ports to Go.)

On 21 January he wrote a class paper, *Comparison of ADA Environment Requirements with UCI Lisp
Environment Facilities*. He went through the environment the class wanted (run at top level,
suspend, inspect variables, edit, break, trace) and showed that UCI Lisp already did most of it. He
proposed replacing Lisp's read-eval-print loop with an Ada reader, so a user would never fall out
into Lisp. The paper already talks about "the ADA interpreter." In outline, that was the plan for
the system the group then built.

## February to May 1981: the gap

None of the code from the next four months survives on the tapes. The last MLISP on tape2 is the
22 January PARSER.MLI, and the later save sets hold only papers. Stephen remembers that he made the
parser generate Lisp code, and that the team then piled on: the result was slow, but it showed the
idea worked. By May the mail shows that it had happened.

## May 1981: a running environment

By late May the group had a working system called **ADARPE**, the **Ada Rapid Prototyping
Environment**, built on Meehan's MLISPF. It lived in a set of shared project directories under
`<ts.aa>`: interpreter, parser, lexical analyzer, **code generation**, pretty printer, break
package, editor, test programs, current system, and documentation-and-news.

At the top level, users typed Ada. On 23 May Stephen explained it to Dave Smith, who had just joined:
`compile("fact.ada");` was an ordinary procedure call, with a short form that dropped the
punctuation and an <esc> key that completed commands. Underneath was the Lisp code-generation
trick: Meehan had added an `ASTMT` (Ada statement) form to MLISP. MLISP compiled it into a Lisp
lambda and stored it on the function's `ASTMT` property, where the environment could fetch and
apply it. Built-ins that couldn't be written in Ada were Lisp functions in the interpreter's
`AFUNS.MLI`, all named starting with "a". There was a symbol table of loaded top-level functions, a
pretty printer, and a `QUIT;` command.

Smith's questions in late May trace the rest of it: `AEXECUTE`, `aparse`, `LINE-PARSE-2`,
`CALL.MLI`. On 28 May he noticed a **new parser**. It was still recursive descent, but now
written as code (`PAR.MLI`, which he counted at 166 EXPRs) rather than an interpreted grammar. He
asked whether it was "going to be *it* for a while," and what had happened to "the LR(k) parser that
was in the wind."

The team was close-knit. At a party on 16 May, Hans Long and Tami gave Standish a gift on the
group's behalf, at Stephen's suggestion.

## June to July 1981: the team piles on

- **1 June.** Stephen rebuilt the system so it printed a build number at every top-level prompt, and
  announced it to the ADA-HACKERS list (Whitehill, Eilers, Klefstad, Smith). Smith had suggested it.
- **2 June.** Standish asked the computing facility for 5,000 more disk blocks. He called it "an
  extremely important research project that has made fantastic progress in the past few months."
- **2 June.** Smith published the first documentation of the **UCI Internal Ada Format**: a tree
  grammar for the trees PAR.MLI produced, plus a checker for undefined and duplicate nonterminals.
  He meant it as the contract between the parser, pretty printer and interpreter.
- **Early July.** Smith built his own copy of ADARPE so he could add a structure editor and a
  modified lexer. Stephen gave him the disk quota and worried a little that the two versions
  would drift apart. Smith's lexer change, made so that Ada commands could use English phrases
  like `FOR ALL (X) IN (S) ASSERT (P(X))`, recognized reserved words only in lower case.
- **9 July.** Stephen explained the terminal handling: Klefstad had written terminal I/O that did
  command completion and took over line editing.
- **13 and 16 July.** Smith modified PAR.MLI to accept nonterminals in its input, so macros could
  generate code fragments. He then released an **Ada unparser** that turned internal-form trees
  back into lexemes, built for the Ada macro facility he was writing.

## September 1981: the bigger idea

On 4 September Stephen wrote down where he thought this should go (GRANT.RNO, "My thoughts as of
September 4th"). He had just read Tony Hoare's Turing lecture warning that Ada would be another
Algol 68. His answer was that Ada was a poor language for writing a sort routine but a great
opportunity for **environments** research. Ada had enough in it to *be* the operating system,
command language, debugger, editor and file manager, so the separate edit/compile/debug tools
could go away. His example was finding a bug in the task dispatcher, then starting a new
operating system from inside the break to report it. He laid out what a prototype on the
TOPS-20 would need: a compiler they could fake, a programming style that kept Ada's features
apart, a selective-viewing tool for designs and code, and an Ada-shaped operating system ready for
local networks.

The same month he and Standish talked about a book for beginners ("Ada for Beginners" or "Ada Made
Easy"). Stephen dropped it so he could concentrate on his degree.

## Afterward: Arcturus

ADARPE's main lesson went into the next system. According to Stephen's thesis, the first version of
**Arcturus** was written in UCI Lisp on a DEC-2020 by Whitehill and Klefstad, both ADA-HACKERS from
1981. Whitehill and Stephen then rewrote it in C on a VAX (the C rewrite is a separate story). The
thesis gives the reason for one of its main design choices: "Experience with the ADA Rapid
Prototyping Environment (ADARPE) [Willson] indicated that representing the program as
s-expressions was too space-expensive." So Arcturus used records for its syntax trees. The other
ideas carried over from 1981: an interactive top level where you type Ada, Klefstad's completing
command shell (now **Adash**), a shared internal form, and generic tree walkers in the spirit of
Smith's unparser.

## What's missing

The ADARPE source (`<ts.aa.*>`) wasn't on any of the six tapes. If it survives, it would be on a UCI
backup of the TS account or in someone's listings. Smith, Klefstad, Whitehill and Standish would be
the people to ask. The supporting detail and file paths are in Stephen's notes on the tape recovery.
