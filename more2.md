# Additional autobench testing

Below are additional runs of autobench on differing functions.

## Boilerplate

The following boilerplate is used to run autobench with.  It is the same
boilerplate for all runs.  The only difference is in the call to `result`:
the number of times the code was repeated replaces the 0, and the amount
of padding inserted varies.

```
// Repeated code goes here

report(string msg) {
    llOwnerSay("\nRESULT:" + msg);
}

result(integer mem, integer count) {
    
    if (count) {
        integer old = (integer)llLinksetDataRead("mem");
        string s = llLinksetDataRead("name");
    report("TEST_MEM=" + (string)mem);
    report("SIZE=" + (string)((float)(mem - old)/(float)count));
    } else {
        llLinksetDataWrite("mem", (string)mem);
        report("BASE_MEM=" + (string)mem);
    }
    llOwnerSay("DONE");
}

default {
    state_entry() {
        result(llGetUsedMemory(), 0);
        return;
        integer i;
        // i+i+... padding goes here
    }
}
```

Note there are two functions defined in the boilerplate so the assumption
that the first function adds an extra 24 bytes is unsubstantiated.

## Test Results

### `fNN_CNT(integer i) {llDie();}`

| N | Copy | Copy/N | Padding | Padding/N | ΔPadding |
|---|---|---|---|---|---|
| 1 |  384 ±7 | 384 |  416 | 416 | - |
| 2 |  776 ±7 | 388 |  808 | 404 | 392 |
| 3 |  1168 ±15 | 389 |  1196 | 398 | 388 |
| 4 |  1552 ±15 | 388 |  1588 | 397 | 392 |
| 5 |  1920 ±31 | 384 |  1972 | 394 | 384 |
| 6 |  2336 ±31 | 389 |  2364 | 394 | 392 |
| 7 |  2720 ±31 | 388 |  2752 | 393 | 388 |
| 8 |  3104 ±31 | 388 |  3144 | 393 | 392 |

### `fNN_CNT(integer i) {llDie();jump x; @x;}`

| N | Copy | Copy/N | Padding | Padding/N | ΔPadding |
|---|---|---|---|---|---|
| 1 |  392 ±7 | 392 |  424 | 424 | - |
| 2 |  792 ±7 | 396 |  824 | 412 | 400 |
| 3 |  1184 ±15 | 394 |  1220 | 406 | 396 |
| 4 |  1584 ±15 | 396 |  1620 | 405 | 400 |
| 5 |  1984 ±31 | 396 |  2012 | 402 | 392 |
| 6 |  2368 ±31 | 394 |  2412 | 402 | 400 |
| 7 |  2784 ±31 | 397 |  2808 | 401 | 396 |
| 8 |  3168 ±31 | 396 |  3208 | 401 | 400 |
| 9 |  3520 ±63 | 391 |  3600 | 400 | 392 |
| 10 |  3968 ±63 | 396 |  4000 | 400 | 400 |
| 11 |  4352 ±63 | 395 |  4396 | 399 | 396 |
| 12 |  4736 ±63 | 394 |  4796 | 399 | 400 |
| 13 |  5120 ±63 | 393 |  5188 | 399 | 392 |
| 14 |  5568 ±63 | 397 |  5588 | 399 | 400 |
| 15 |  5952 ±63 | 396 |  5984 | 398 | 396 |
| 16 |  6336 ±63 | 396 |  6384 | 399 | 400 |

### `fNN_CNT(integer i){}`

| N | Copy | Copy/N | Padding | Padding/N | ΔPadding |
|---|---|---|---|---|---|
| 1 |  45 ±0 | 45 |  52 | 52 | - |
| 2 |  90 ±0 | 45 |  100 | 50 | 48 |
| 3 |  134 ±1 | 44 |  144 | 48 | 44 |
| 4 |  180 ±1 | 45 |  188 | 47 | 44 |
| 5 |  224 ±1 | 44 |  232 | 46 | 44 |
| 6 |  268 ±3 | 44 |  280 | 46 | 48 |
| 7 |  312 ±3 | 44 |  324 | 46 | 44 |
| 8 |  360 ±3 | 45 |  368 | 46 | 44 |
| 9 |  400 ±7 | 44 |  412 | 45 | 44 |
| 10 |  448 ±7 | 44 |  460 | 46 | 48 |
| 11 |  488 ±7 | 44 |  504 | 45 | 44 |
| 12 |  536 ±7 | 44 |  548 | 45 | 44 |
| 13 |  584 ±7 | 44 |  592 | 45 | 44 |
| 14 |  624 ±7 | 44 |  640 | 45 | 48 |
| 15 |  672 ±7 | 44 |  684 | 45 | 44 |
| 16 |  720 ±7 | 45 |  728 | 45 | 44 |

### `fNN_CNT(){}`

| N | Copy | Copy/N | Padding | Padding/N | ΔPadding |
|---|---|---|---|---|---|
| 1 |  39 ±0 | 39 |  44 | 44 | 44 |
| 2 |  78 ±0 | 39 |  80 | 40 | 36 |
| 3 |  117 ±0 | 39 |  120 | 40 | 40 |
| 4 |  156 ±1 | 39 |  160 | 40 | 40 |
| 5 |  194 ±1 | 38 |  200 | 40 | 40 |
| 6 |  234 ±1 | 39 |  236 | 39 | 36 |
| 7 |  272 ±3 | 38 |  276 | 39 | 40 |
| 8 |  312 ±3 | 39 |  316 | 39 | 40 |
| 9 |  348 ±3 | 38 |  356 | 39 | 40 |
| 10 |  384 ±7 | 38 |  392 | 39 | 36 |
| 11 |  424 ±7 | 38 |  432 | 39 | 40 |
| 12 |  464 ±7 | 38 |  472 | 39 | 40 |
| 13 |  504 ±7 | 38 |  512 | 39 | 40 |
| 14 |  544 ±7 | 38 |  548 | 39 | 36 |
| 15 |  584 ±7 | 38 |  588 | 39 | 40 |
| 16 |  624 ±7 | 39 |  628 | 39 | 40 |

### `fNN_CNT(){jump x; @x;}`

| N | Copy | Copy/N | Padding | Padding/N | ΔPadding |
|---|---|---|---|---|---|
| 1 |  43 ±0 | 43 |  48 | 48 | - |
| 2 |  86 ±0 | 43 |  88 | 44 | 40 |
| 3 |  128 ±1 | 42 |  132 | 44 | 44 |
| 4 |  172 ±1 | 43 |  176 | 44 | 44 |
| 5 |  214 ±1 | 42 |  220 | 44 | 44 |
| 6 |  256 ±3 | 42 |  260 | 43 | 40 |
| 7 |  300 ±3 | 42 |  304 | 43 | 44 |
| 8 |  344 ±3 | 43 |  348 | 43 | 44 |
| 9 |  384 ±7 | 42 |  392 | 43 | 44 |
| 10 |  424 ±7 | 42 |  432 | 43 | 40 |

### `ffNN_CNT(){}`

*Note it is `ff` and not `f` in the name*

| N | Copy | Copy/N | Padding | Padding/N | ΔPadding |
|---|---|---|---|---|---|
| 1 |  40 ±0 | 40 |  44 | 44 | - |
| 2 |  80 ±1 | 40 |  84 | 42 | 40 |
| 3 |  120 ±1 | 40 |  124 | 41 | 40 |
| 4 |  160 ±1 | 40 |  164 | 41 | 40 |
| 5 |  200 ±3 | 40 |  204 | 40 | 40 |
| 6 |  240 ±3 | 40 |  244 | 40 | 40 |
| 7 |  280 ±3 | 40 |  284 | 40 | 40 |
| 8 |  320 ±3 | 40 |  324 | 40 | 40 |

### `ffffNN_CNT(){}`

*Note it is `ffff` and not `f` in the name*

| N | Copy | Copy/N | Padding | Padding/N | ΔPadding |
|---|---|---|---|---|---|
| 1 |  42 ±0 | 42 |  48 | 48 | 48 |
| 2 |  84 ±1 | 42 |  88 | 44 | 40 |
| 3 |  126 ±1 | 42 |  132 | 44 | 44 |
| 4 |  168 ±1 | 42 |  172 | 43 | 40 |
| 5 |  208 ±3 | 41 |  216 | 43 | 44 |
| 6 |  252 ±3 | 42 |  256 | 42 | 40 |
| 7 |  292 ±3 | 41 |  300 | 42 | 44 |
| 8 |  336 ±3 | 42 |  340 | 42 | 40 |

###  `fNN_CNT(){llDie();}` (preamble `fff(){llDie();}`)

| N | Copy | Copy/N | Padding | Padding/N | ΔPadding |
|---|---|---|---|---|---|
| 1 |  344 ±7 | 344 |  340 | 340 | - |
| 2 |  688 ±7 | 344 |  684 | 342 | 344 |
| 3 |  1040 ±15 | 346 |  1024 | 341 | 340 |
| 4 |  1376 ±15 | 344 |  1372 | 343 | 348 |
| 5 |  1728 ±31 | 345 |  1712 | 342 | 340 |
| 6 |  2080 ±31 | 346 |  2056 | 342 | 344 |
| 7 |  2400 ±31 | 342 |  2396 | 342 | 340 |
| 8 |  2752 ±31 | 344 |  2744 | 343 | 348 |

###  `fNN_CNT(){}` (preamble `fff(){llDie();}`)

| N | Copy | Copy/N | Padding | Padding/N | ΔPadding |
|---|---|---|---|---|---|
| 1 |  40 ±1 | 40 |  44 | 44 | - |
| 2 |  78 ±1 | 39 |  80 | 40 | 36 |
| 3 |  118 ±1 | 39 |  124 | 41 | 44 |
| 4 |  156 ±3 | 39 |  160 | 40 | 36 |
| 5 |  196 ±3 | 39 |  200 | 40 | 40 |
| 6 |  236 ±3 | 39 |  236 | 39 | 36 |
| 7 |  272 ±3 | 38 |  280 | 40 | 44 |
| 8 |  312 ±3 | 39 |  316 | 39 | 36 |

###  `ffNN_CNT(){}` (preamble `fff(){llDie();}`)

*Note it is `ff` and not `f` in the name*

| N | Copy | Copy/N | Padding | Padding/N | ΔPadding |
|---|---|---|---|---|---|
| 1 |  40 ±1 | 40 |  44 | 44 | - |
| 2 |  80 ±1 | 40 |  84 | 42 | 40 |
| 3 |  120 ±1 | 40 |  124 | 41 | 40 |
| 4 |  160 ±3 | 40 |  164 | 41 | 40 |
| 5 |  200 ±3 | 40 |  204 | 40 | 40 |
| 6 |  240 ±3 | 40 |  244 | 40 | 40 |
| 7 |  280 ±3 | 40 |  284 | 40 | 40 |
| 8 |  320 ±3 | 40 |  324 | 40 | 40 |

###  `ffffNN_CNT(){}` (preamble `fff(){llDie();}`)

| N | Copy | Copy/N | Padding | Padding/N | ΔPadding |
|---|---|---|---|---|---|
| 1 |  42 ±1 | 42 |  48 | 48 | - |
| 2 |  84 ±1 | 42 |  88 | 44 | 40 |
| 3 |  126 ±1 | 42 |  132 | 44 | 44 |
| 4 |  168 ±3 | 42 |  172 | 43 | 40 |
| 5 |  212 ±3 | 42 |  216 | 43 | 44 |
| 6 |  252 ±3 | 42 |  256 | 42 | 40 |
| 7 |  296 ±3 | 42 |  300 | 42 | 44 |
| 8 |  336 ±3 | 42 |  340 | 42 | 40 |

###  `fNN_CNT(integer i){}` (preamble `fff(){llDie();}`)

| N | Copy | Copy/N | Padding | Padding/N | ΔPadding |
|---|---|---|---|---|---|
| 1 |  46 ±1 | 46 |  56 | 56 | - |
| 2 |  90 ±1 | 45 |  100 | 50 | 44 |
| 3 |  136 ±1 | 45 |  144 | 48 | 44 |
| 4 |  180 ±3 | 45 |  188 | 47 | 44 |
| 5 |  224 ±3 | 44 |  236 | 47 | 48 |
| 6 |  272 ±3 | 45 |  280 | 46 | 44 |
| 7 |  316 ±3 | 45 |  324 | 46 | 44 |
| 8 |  360 ±7 | 45 |  368 | 46 | 44 |

###  `fNN_CNT(integer i){llDie();}` (preamble `fff(){llDie();}`)

| N | Copy | Copy/N | Padding | Padding/N | ΔPadding |
|---|---|---|---|---|---|
| 1 |  392 ±7 | 392 |  404 | 404 | - |
| 2 |  784 ±7 | 392 |  792 | 396 | 388 |
| 3 |  1168 ±15 | 389 |  1184 | 394 | 392 |
| 4 |  1568 ±15 | 392 |  1568 | 392 | 384 |
| 5 |  1952 ±31 | 390 |  1960 | 392 | 392 |
| 6 |  2336 ±31 | 389 |  2348 | 391 | 388 |
| 7 |  2720 ±31 | 388 |  2740 | 391 | 392 |
| 8 |  3136 ±31 | 392 |  3124 | 390 | 384 |

###  `fNN_CNT(integer i){llDie();jump x;@x;}` (preamble `fff(){llDie();}`)

| N | Copy | Copy/N | Padding | Padding/N | ΔPadding |
|---|---|---|---|---|---|
| 1 |  400 ±7 | 400 |  412 | 412 | - |
| 2 |  800 ±15 | 400 |  808 | 404 | 396 |
| 3 |  1200 ±15 | 400 |  1208 | 402 | 400 |
| 4 |  1600 ±15 | 400 |  1600 | 400 | 392 |
| 5 |  1984 ±31 | 396 |  2000 | 400 | 400 |
| 6 |  2400 ±31 | 400 |  2396 | 399 | 396 |
| 7 |  2784 ±31 | 397 |  2796 | 399 | 400 |
| 8 |  3200 ±31 | 400 |  3188 | 398 | 392 |

###  `fNN_CNT(){jump x; @x;}` (preamble `fff(){llDie();}`)

| N | Copy | Copy/N | Padding | Padding/N | ΔPadding |
|---|---|---|---|---|---|
| 1 |  44 ±1 | 44 |  48 | 48 | - |
| 2 |  86 ±1 | 43 |  88 | 44 | 40 |
| 3 |  130 ±1 | 43 |  136 | 45 | 48 |
| 4 |  172 ±3 | 43 |  176 | 44 | 40 |
| 5 |  216 ±3 | 43 |  220 | 44 | 44 |
| 6 |  260 ±3 | 43 |  260 | 43 | 40 |
| 7 |  300 ±3 | 42 |  308 | 44 | 48 |
| 8 |  344 ±7 | 43 |  348 | 43 | 40 |

## Analysis: is there extra space allocated per builtin called?

**Yes -- a one-time ~20-byte import allocation per DISTINCT builtin, not per call
site, and separate from the per-function call cost.**

The two robust quantities are the **slope** (per-function marginal cost; every
four-function span adds exactly `4*slope`, so it is alignment-free) and the
**intercept** (the N-independent one-time cost) from fitting
`Padding(N) = slope*N + intercept`.

### Per-function slopes

| body | slope (bytes/fn) | vs `f(){}` |
|------|-----------------:|-----------|
| `f(){}`              |  39 | baseline (7-char name `f00_000`) |
| `ff(){}`             |  40 | +1 per extra name char |
| `ffff(){}`           |  42 | +1 per extra name char |
| `f(int i){}`         |  45 | +6  (int param, no call) |
| `f(){jump x;@x;}`    |  43 | +4  (the jump block) |
| `f(){llDie();}`      | 343 | +304 (the call: migration + dispatch) |
| `f(int i){llDie();}` | 389 | +46  (param *under* a call = 6 + ~40 migration) |

Names cost +1 byte/char in the marginal (linear), even though a *single isolated*
function's total rounds to a 4-byte boundary.  A param costs +6 with no call but
+46 once the function calls something (the extra ~40 is its migration slot).

### The import (one-time, per distinct builtin)

Compare a function to the same function *without* the call.  Same signature means
the slopes share `mod 4`, so the period-4 alignment wobble cancels and the
intercept difference is clean:

| signature pair | intercept no-call | intercept w/ `llDie` | import |
|----------------|------------------:|---------------------:|-------:|
| `f(){}`  vs `f(){llDie();}`          | 3.7 | 24.0 | **20.3** |
| `f(int i){}` vs `f(int i){llDie();}` | 8.3 | 28.0 | **19.7** |

Both give ~20 bytes: the `llDie` import-table entry, allocated once regardless of
how many functions call it (the slope shows each *additional* llDie function adds
only 343 -- no second import).

### The preamble experiment confirms "per distinct builtin"

Pre-calling `llDie` in `fff(){llDie();}` puts the import in the base, so test
functions that also call `llDie` stop paying it -- their cost drops (e.g.
`f(){llDie();}` first reading 368 -> 340).  Control bodies that do NOT call
`llDie` (`f(){}`, `ff(){}`, `ffff(){}`, `f(){jump}`, `f(int i){}`) are unchanged
by the preamble (intercept delta ~0, within ~2 bytes).  The drop *measured from
the preamble* is alignment-smeared (15-25), because adding `fff()` to the base
shifts the alignment phase; the clean ~20 is the same-signature intercept
difference above.

### On the boilerplate note

The earlier "the first function adds ~24 bytes" framing was indeed
unsubstantiated.  The boilerplate already defines two functions (`report`,
`result`), so a test function is never the script's first.  The ~20 is not a
"first function" cost -- it is specifically the `llDie` import, which is exactly
why it disappears once an earlier function (`fff`) already imports `llDie`.
