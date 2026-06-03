# Additional findings

A test was performed by measuring the cost of multiple functions starting with
a single function up to twenty functions.  It was benchmarked by both the
*copy* and *padding* (`-1`) modes of **autobench**.

The functions all looked like:

```
FNN_CNT(){llDie();}
```

Where NN is a two-digit number (00, 01, 02, ...).  For example, the code for the
test case of 3 functions was:

```
F02_CNT(){llDie();} F01_CNT(){llDie();} F00_CNT(){llDie();}
```

| N | Copy | Copy/N | Padding | Padding/N | ΔPadding |
|---|---|---|---|---|---|
| 1 |  340 ±3 | 340 |  368 | 368 | - |
| 2 |  680 ±7 | 340 |  708 | 354 | 340 |
| 3 |  1024 ±15 | 341 |  1052 | 350 | 344 |
| 4 |  1360 ±15 | 340 |  1392 | 348 | 340 |
| 5 |  1712 ±15 | 342 |  1740 | 348 | 348 |
| 6 |  2048 ±31 | 341 |  2080 | 346 | 340 |
| 7 |  2400 ±31 | 342 |  2424 | 346 | 344 |
| 8 |  2720 ±31 | 340 |  2764 | 345 | 340 |
| 9 |  3072 ±31 | 341 |  3112 | 345 | 348 |
| 10 |  3424 ±31 | 342 |  3452 | 345 | 340 |
| 11 |  3776 ±63 | 343 |  3796 | 345 | 344 |
| 12 |  4096 ±63 | 341 |  4136 | 344 | 340 |
| 13 |  4416 ±63 | 339 |  4484 | 344 | 348 |
| 14 |  4800 ±63 | 342 |  4824 | 344 | 340 |
| 15 |  5120 ±63 | 341 |  5168 | 344 | 344 |
| 16 |  5440 ±63 | 340 |  5508 | 344 | 340 |
| 17 |  5824 ±63 | 342 |  5856 | 344 | 348 |
| 18 |  6144 ±63 | 341 |  6196 | 344 | 340 |
| 19 |  6528 ±63 | 343 |  6540 | 344 | 344 |
| 20 |  6848 ±63 | 342 |  6880 | 344 | 340 |

- **N** - Number of functions in CODE
- **Copy** - Cost calculated in Copy Mode
- **Copy/N** - Cost calculated in Copy Mode / N
- **Padding** - Cost calculated in Padding Mode
- **Padding/N** - Cost calculated in Padding Mode / N
- **ΔPadding** - Padding<sub>N</sub> - Padding<sub>N-1</sub>

## A formula for the pattern

Working from the byte-exact **Padding** column (the deltas are derived from it),
the totals fit exactly, for all twenty rows:

```
Padding(N) = 343*N + base(N mod 4)

    base(0) = 20      base(1) = 25
    base(2) = 22      base(3) = 23
```

Equivalently, letting `ceil4(x)` be `x` rounded up to the next multiple of 4:

```
Padding(N) = ceil4(343*N) + 20 + (N mod 4 == 1 ? 4 : 0)
```

The user's guessed `(X + N*Y) % Z` form does NOT fit: a single modulus produces a
monotone sawtooth (0, Y, 2Y, ... wrapping), but the marginal sequence is the
non-monotone period-4 cycle

```
ΔPadding by N mod 4:   {1: 348, 2: 340, 3: 344, 0: 340}
ΔPadding - 340:        {1:   8, 2:   0, 3:   4, 0:   0}
```

That is a 4-entry lookup, not one modular expression; the only "modulus" is
`N mod 4`.

### What it tells us

- **Per-function cost = 343 bytes.**  This is the slope: every span of four
  functions adds exactly `4 * 343 = 1372` bytes (verified at every aligned
  point, e.g. `Padding(5)-Padding(1) = Padding(9)-Padding(5) = ... = 1372`).  So
  one more `llDie` function costs 343 bytes -- the true marginal.  (Copy mode's
  340-343 readings are this value within +-noise; the lslparse cilcost model
  predicts 336 for this function, -7, within margin.)

- **One-time base cost ~= 20-22 bytes.**  The constant term -- the `+20`, or the
  average of the four base values (22.5) -- is the fixed overhead that does NOT
  scale with N.  It is the once-only cost of introducing these functions over
  the bare harness: essentially the `llDie` import-table entry, shared by all the
  functions, so only the first one pays it.  This matches the ~24 estimate: the
  first function reads `368 = 343 + 25`, where the extra 25 is this base plus its
  alignment rounding.  (Independently, a harness-USED builtin -- `llGetUsedMemory`
  -- showed only a +4 first-function bump, i.e. no new import, so ~20 of the ~24
  is the import entry itself.)

- **4-byte alignment, and why the period is 4.**  Function code packs on 4-byte
  boundaries, so every `Padding` value is a multiple of 4.  Because
  `343 = 3 (mod 4) = -1 (mod 4)`, each added function shifts the running total's
  position mod 4 by -1, so the pad needed to reach the next 4-boundary walks
  `0 -> 1 -> 2 -> 3 -> 0` with period 4.  That walk is what makes the marginal
  cost cycle 340 / 344 / 348 around the true 343 while the total stays 4-aligned.
  There is one extra `+4` at `N mod 4 == 1` (the `+4` term above): this looks
  like a *second* 4-aligned structure -- e.g. a per-method table growing ~1
  byte/function and crossing a boundary on that phase -- but a single 20-row
  dataset can't pin its exact split from the body, so the empirical `base(N mod
  4)` lookup is the safest exact statement.
