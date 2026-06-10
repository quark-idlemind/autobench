# autobench - memory benchmark for LSL scripts

**autobench** does benchmarks for LSL scripts.  You provide it the
LSL code you want to benchmark the size of and it reports how many
bytes that code will occupy.

See the *Configuration* section below.

## Usage

```
Usage: autobench [-1v] [--chatlog PATH] [--code CODE] [--globals NAME,...]
                 [--ipad PADDING] [--locals NAME,...] [--lsl SCRIPT] [--pad PAD]
                 [--params NAME,...] [--postamble POSTAMBLE]
                 [--preamble PREAMBLE] [--probe] [--statement CODE]
                 [--title NAME] [FILE]
 -1                Using padding and a single copy of CODE
     --chatlog=PATH
                   path to chat.txt [/Users/prb/Library/Application
                   Support/Firestorm/quark_idlemind/chat.txt]
     --code=CODE   code to test
     --globals=NAME,...
                   declare globals
     --ipad=N      initial guess at padding for -1
     --locals=NAME,...
                   locals used with --statement
     --lsl=SCRIPT  path to temporary LSL script to write to
     --pad=PAD     padding instructions
     --params=NAME,...
                   parameters used with --statement
     --postamble=POSTAMBLE
                   Make STR the test's postamble
     --preamble=PREAMBLE
                   Make STR the test's preamble
     --probe       send a simple script to LSL as a probe
     --statement=CODE
                   statement(s) to test
     --title=NAME  name of benchmark
 -v                show the source before each execution
```

Exactly one of `--code`, `--statement`  or `FILE` must be specified.

## Two modes of operation

**autobench** has two modes.  By default it uses *copy* mode.  Using `-1`
causes it to use *padding* mode.  *copy* mode is much faster than *padding*
mode.

Since LSL allocates memory in 512-byte chunks, code can only see changes in
size in multiples of 512 bytes.  Each mode mitigates this in a different way.

### Copy Mode

In *copy* mode **autobench** determines the size of CODE by inserting multiple
copies of CODE (up to 512 copies) and measuring the total.  This will run 3 LSL
scripts:

*  Determine the base line
*  Rough estimate by using 4 copies
*  Best estimate using the greatest power of 2 that fits

While fast, it will generally undercount the size of the CODE.  For example,
if the code has a literal string, only a single copy of it will exist and its
cost is then amortized across N copies.

### Padding Mode

In *padding* mode **autobench** determines the size of CODE with a single
copy of CODE.  It does this as follows:

*  Determine the base line
*  Add padding until the base script tips over to the next memory allocation
*  Run CODE with the determined padding
*  Add more padding until the CODE size tips over to the next memory allocation
*  Subtract the additional padding from the CODE size increase

This produces a more accurate value for a single copy of CODE.  The down side
is this may end up running 20 versions of the script as it figures out padding.

Padding is determined by adding padding to the base until adding one
more byte will cause the measured memory to increase.  This is initially done
against the boilierplate plus preamble and postamble.  Once that is determined
it does the same thing, except this time with the CODE added.  It then
subtracts how much more padding was needed to tip into the next memory block
and subtracts that amount from the difference between the two values.  The two
values are always a multiple of 512 bytes apart.

Note that padding can be off by one if less than 5 bytes of padding as 5 is
the smallest odd value of bytes that **autobench** knows how to pad by.  Its
default padding is a power of 2.

## Examples:

**Determine the size of a function**

```
$ autobench --code "foo_CNT(){llDie();}"
Size: 340 ±3
```

```
$ autobench -1 --code "foo_CNT(){llDie();}"
Padding: 458
Size: 368
```

LSL adds a fixed number of bytes for each builting function that is called
someplace in the script.  Using --preamble can mitigate this:

```
$ autobench -1 --preamble "f(){llDie();}" --code "foo_CNT(){llDie();}"
Padding: 102
Size: 340
```
And now the two modes agree on the size of the function.  In *copy* mode
this fixed cost (about 20 bytes) is amortized across up to 512 calls or less
that 0.04 per builtin function called so it has no impact.

The result from *padding* mode can undercount by up to 3 bytes.  The correct
value for the function above is 343 bytes.

You can use `--ipad` with displayed padding when in *padding* mode.
**autobench** will check the a pad of N and a pad of N+1 result in a
different amount of memory being allocated.  If you don't know the
right padding don't use `--ipad`.

** Determine the size of a single statement**
```
$ autobench --preamble "integer g; foo() {" --code "g++;" --postamble "}"   
Size: 28 ±1
```

## Options:

### --chatlog

Specifies the chatlog to monitor.  This normally does not need to be specified.

### --code

Specifies the actual code to benchmark if FILE is not specified

### --globals

Specifies global variables to declare.  See the section below about variable
names and types.

### --lsl

Specifies where to write the LSL script.  This normally does not need to be specified.

### --locals

Specifies local variables to declare when used with `--statement`.  See the
section below about variable names and types.

### --pad

LSL instructions placed in the code to pad the base size.  These instructions are never executed.

### --params

Specifies parameters for the generated function by `--statement`.  See the
section below about variable names and types.

### --preamble/--postamble

A single copy of the preamble is written to the top of the script.
A single copy of the postamble is written after the test code has been written.

This is normally used when global variables are needed or when wanting to benchmark individual lines of LSL rather than an entire function.

### --probe

`--probe` sends a small script to SL and replies with `SL Live` if the script
successfully ran or prints an error and exits with a non-zero status.  IF
`--probe` is specified other parameters are ignored and no benchmark is run.

### --statement

Specifies the statement to benchmark.  The statements will be repeated inside
of a function named `_`.  For example `--statement "llDie();"` results in

```
_() {
llDie();
...
}
```

## Variables

The `--globals`, `--locals`, and `--params` are used to declare variables to
automatically insert into the benchmark.  `--locals` and `--params` are only
used with `--statement`.  These variables are not part of the measurements.

### Names and types

The type of the variable is dependent on the first letter of its name (either case):

| Letter | Type |
|---|---|
| a | list |
| f | float |
| g | float |
| i | integer |
| j | integer |
| k | key |
| l | list |
| q | quaternion |
| r | quaternion |
| s | string |
| v | vector |
| x | integer |

Locals and globals may have values assigned (`--local i=1,f=3.14`).

The globals are inserted before any preamble.
The parameters are inserted in the function generated by `--statement`.
The locals are inserted as the first statements in the function generated by `--statement`.

Example: `autobench --globals f --params s --locals i,s1=\"foo\" --statement "llOwnerSay((string)i + s + s1);"`

Produces the code

```
float f;
_(string s) {
	integer i;
	string s1 = "foo";

	llOwnerSay((string)i + s + s1);
	...
}
```

## Configuration

**autobench** works by writing LSL scripts to a temporary file
created when using the *External Editor* feature from
within Second Life.  It reads results from the `chat.txt` file of
a user logged into Second Life.

The chat.txt file is located in different places for different
operating systems.  On macOS it normally is in one of:

- `/Users/`*myaccount*`/Library/Application Support/Firestorm/`*my_login*`/chat.txt`
- `/Users/`*myaccount*`/Library/Application Support/SecondLife/`*my_login*`/chat.txt`

Write this into the file `$HOME/.automate/who`.

Next you will need to log into Second Life with that account and edit a script that is in an object.
You can just create a new object and new script in it.  Edit that script and use the **Edit** button
to open it in an external editor.  On macOS it will open a file with a name something like:

`/var/folders/_8/x72_19wj04b3a1_ff8vk_3mr0000gn/T/sl_script_New Script_8f3c7b2a10dfed43ba9c041eaf3deb52.lsl`

Make a symbolic link named `$HOME/.automate/source` point to this file.

You just keep the edit window open inside of Second Life but, at
least on macOS, you can close the external editor.

**autobench** will write scripts to `$HOME/.automate/source`.  At the same time it is watching the
file named in `$HOME/.automate/who` for lines of output that either contain "RESULT:" or "DONE".

When it sees "RESULT:" it reports that back.  When it sees "DONE" it moves on to the next script.

## Details

**autobench** works by first producing a NUL version of the test code.  It consists of:

*  Globals from `--globals`
*  The preamble (if any)
*  `_(parameters) { locals; }` (with `--statement`)
*  The postamble (if any)
*  The test harness

This is run to determine the base memory use.

It then creates a script with multiple copies of the test code.  In the test code the word CNT is replaced with the iteration count.  These scripts look like:

*  Globals from `--globals`
*  The preamble (if any)
*  `_(parameters) { locals`; (with `--statement`)
*  N copies of CODE with CNT replaced by an index between [1..N)
*  `}` (with `--statement`)
*  The postamble (if any)
*  The test harness

It first benchmarks the CODE with 4 copies, using this to determine the maximum number of copies to test with.
The size is reported with how many bytes the size may be off.

## Questions (added during review)

These are points that were unclear, under-documented, or possibly wrong while
reading the README against the code.  My best current understanding is noted so
you can confirm or correct.


4. **`--ipad` semantics could be spelled out.**  From the code, `--ipad N` is a
   guess at the *base* padding (phase 1) only: if `N` already tips the base it is
   taken as exact, otherwise it seeds the binary-search midpoint; phase 2 (the
   CODE search) is always redone.  Is the intended workflow to take the
   `Padding:` value printed by a prior `-1` run and feed it straight back as
   `--ipad` for subsequent runs that change only CODE?  Stating that explicitly
   (with the printed `Padding:` line as the source) would help.

5. **`llDie` (and other control-terminating builtins) disagree between modes.**
   `llDie()` measures 340 by copies but 368 by padding — and by padding it reads
   *larger* than `llOwnerSay("")` despite taking no argument, which inverts the
   copy-mode ordering.  **Resolved during review:** I measured `llResetTime()`
   (another harness-unused void builtin) -> the same +28 padding offset, while
   harness-USED builtins show a small offset (`llGetUsedMemory` +4, `llOwnerSay`
   +12).  So this is a **padding-mode import confound**, not special-casing: a
   builtin the measurement harness does not itself use pays a one-time import
   entry (~+16) in the single padded copy, which copy mode amortizes across
   copies.  `llDie` is not special.  A short caveat seems warranted that
   padding-mode absolute size includes one-time per-distinct-builtin import cost
   (confounded by which builtins the harness already imports), so it over-states
   a function calling a builtin that is used elsewhere in a real script.

6. **Padding is dead code after `return;`.**  The injected `integer i; i+i+...;`
   and `jump Z; @Z;` sit after the harness `return;` and never execute.  The
   method relies on unexecuted CIL costing the same per byte as live CIL (so the
   pad faithfully stands in for real code size).  Worth stating that assumption
   explicitly, since it underpins the whole padding approach.

7. **`--statement` and `-1` together.**  Does padding mode compose with
   `--statement` (single copy inside the generated `_()` function), or is `-1`
   intended for `--code` / `FILE` only?  The two-modes section describes `-1`
   only in terms of "a single copy of CODE."
