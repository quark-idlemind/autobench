package main

import (
	"errors"
	"fmt"
	"math/bits"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/quark-idlemind/autobench/bench"

	"github.com/pborman/options"
)

var flags = struct {
	Preamble  string   `getopt:"--preamble=PREAMBLE Make STR the test's preamble"`
	Postamble string   `getopt:"--postamble=POSTAMBLE Make STR the test's postamble"`
	Code      string   `getopt:"--code=CODE code to test"`
	Statement string   `getopt:"--statement=CODE statement(s) to test"`
	Pad       string   `getopt:"--pad=PAD padding instructions"`
	ChatLog   string   `getopt:"--chatlog=PATH path to chat.txt"`
	Source    string   `getopt:"--lsl=SCRIPT path to temporary LSL script to write to"`
	Show      bool     `getopt:"-v show the source before each execution"`
	Title     string   `getopt:"--title=NAME name of benchmark"`
	Params    []string `getopt:"--params=NAME,... parameters used with --statement"`
	Locals    []string `getopt:"--locals=NAME,... locals used with --statement"`
	Globals   []string `getopt:"--globals=NAME,... declare globals"`
	One       bool     `getopt:"-1 Using padding and a single copy of CODE"`
	IPad      int      `getopt:"--ipad=N initial guess at padding for -1"`
	Fast      bool     `getopt:"--fast skip looking for pad"`
	Debug     bool     `getopt:"--debug enable debugging"`
	Probe     bool     `getopt:"--probe send a simple script to LSL as a probe"`
}{}

func init() {
	flags.ChatLog, _ = bench.GetChatLog()
	flags.Source, _ = bench.GetLSLPath()
}

func errf(format string, v ...any) {
	fmt.Fprintf(os.Stderr, format, v...)
	os.Exit(1)
}

func debugf(format string, v ...any) {
	if flags.Debug {
		fmt.Fprintf(os.Stderr, format, v...)
	}
}

const plusminus = "±"
const blockSize = int(512)

func findPadding(b *bench.Bench, cnt, pad int) int {
	var r Results
	getBase := func() int { return r.Base }
	if cnt > 0 {
		getBase = func() int { return r.Test }
	}
	runScript(b, cnt, pad, &r)
	debugf("Base[%d] %d : %d\n", cnt, pad, getBase())

	var mid int
	low := 0
	high := blockSize
	base := getBase()
	for high-low > 1 {
		mid = low + (high-low)/2
		runScript(b, cnt, mid+pad, &r)
		debugf("Check[%d] %d < %d < %d + %d : %d\n", cnt, low, mid, high, pad, getBase())
		if base == getBase() {
			low = mid
		} else {
			high = mid
		}
	}
	for {
		runScript(b, cnt, low+pad, &r)
		debugf("Got[%d] %d + %d: %d\n", cnt, low, pad, getBase())
		if base != getBase() {
			return low
		}
		low++
	}
}

const probeScript = `
default {
	state_entry() {
		llOwnerSay("RESULT:Hello");
		llOwnerSay("DONE");
	}
}
`

func main() {
	defer func() {
		if p := recover(); p != nil {
			errf("%v\n", p)
		}
	}()
	args := options.RegisterAndParse(&flags)
	if flags.ChatLog == "" {
		errf("Unable to find chat.txt\n")
	}
	if flags.Source == "" {
		errf("Unable to find temporary .lsl file to write to\n")
	}
	b, err := bench.New(flags.Source, flags.ChatLog)
	if err != nil {
		errf("%v\n", err)
	}
	if flags.Probe {
		ch := make(chan struct{})
		var results []string
		go func() {
			results, err = b.Send(probeScript)
			close(ch)
		}()
		select {
		case <-ch:
			if err != nil {
				errf("%v\n", err)
			}
			if len(results) == 0 {
				errf("Script returned no results")
			}
			if results[0] != "Hello" {
				errf("Got %q, want %q\n", results[0], "Hello")
			}
			fmt.Printf("SL Live\n")
			return
		case <-time.After(time.Minute):
			errf("timed out waiting for Second Life\n")
		}
		return
	}
	switch {
	case len(args) > 1:
		errf("At most 1 test file may be specified\n")
	case len(args) < 1 && flags.Code == "" && flags.Statement == "":
		errf("Either --code, --statement or a file must be specified\n")
	case len(args) == 1 && flags.Code != "":
		errf("Only one of --code or a file may be specified\n")
	case len(args) == 1 && flags.Statement != "":
		errf("Only one of --statement or a file may be specified\n")
	case flags.Code != "" && flags.Statement != "":
		errf("Only one of --code or --statement may be specified\n")
	case flags.Code == "" && flags.Statement == "":
		data, err := os.ReadFile(args[0])
		if err != nil {
			errf("%v\n", err)
		}
		flags.Code = string(data)
	case flags.Statement != "":
		flags.Code = flags.Statement
	}
	var globals string
	for _, g := range flags.Globals {
		g, err := mkVar(g)
		if err != nil {
			errf("%v\n", err)
		}
		globals += g + ";\n"
	}
	flags.Code = strings.TrimSpace(flags.Code)
	flags.Preamble = globals + strings.TrimSpace(flags.Preamble)
	flags.Postamble = strings.TrimSpace(flags.Postamble)
	if flags.Statement != "" {
		flags.Preamble += "\n_("
		for i, p := range flags.Params {
			p, err := mkVar(p)
			if err != nil {
				errf("%v\n", err)
			}
			if strings.Contains(p, "=") {
				errf("%s: parameters must not be initialized", p)
			}
			if i > 0 {
				flags.Preamble += ", " + p
			} else {
				flags.Preamble += p
			}
		}
		flags.Preamble += ") {\n"
		for _, v := range flags.Locals {
			v, err := mkVar(v)
			if err != nil {
				errf("%v\n", err)
			}
			flags.Preamble += v + ";\n"
		}
		flags.Postamble = "\n}\n" + flags.Postamble
	}

	var r Results

	if flags.One {
		// Find the smallest pad above 5 that
		// that does not tip us into the next bucket.
		pad := findPadding(b, 0, 5) + 5
		pad = findPadding(b, 1, pad)

		if r.Title != "" {
			fmt.Printf("Title: %s\n", r.Title)
		}

		fmt.Printf("Size: %d\n", 512+r.Test-r.Base-pad)
		return
	}

	// We should really always pad the base
	pad := 0
	if !flags.Fast {
		pad = findPadding(b, 0, 5) + 5
	}
	// Get the base result
	runScript(b, 0, pad, &r)

	D := 8
	runScript(b, D, pad, &r)
	for r.Size == 0 {
		D *= 2
		runScript(b, D, pad, &r)
	}
	var cnt int
	if r.Size == 0 {
		cnt = blockSize
	} else {
		maxMem := 62*1024 - r.Base
		cnt = maxMem / (int(r.Size) + blockSize/(2*D))
		cnt = int(1 << (bits.Len(uint(cnt)) - 1))
		if cnt == 0 {
			fmt.Print("Unable to benchmark\n")
		}
	}
	runScript(b, cnt, pad, &r)
	if r.Title != "" {
		fmt.Printf("Title: %s\n", r.Title)
	}
	fmt.Printf("Size: %d %s%d\n", int(r.Size), plusminus, 511/cnt)
}

type Results struct {
	Title string
	Base  int
	Test  int
	Size  float64
}

type Cache struct {
	Count   int
	Padding int
}

// cache prevents us from running the exact same script twice.
var cache = map[Cache]Results{}

func runScript(b *bench.Bench, cnt, pad int, r *Results) {
	key := Cache{Count: cnt, Padding: pad}
	if or, ok := cache[key]; ok {
		debugf("Using cache for %v\n", key)
		*r = or
		return
	}
	defer func() {
		cache[key] = *r
	}()
	var buf strings.Builder
	if flags.Preamble != "" {
		buf.WriteString(flags.Preamble)
		buf.WriteString("\n")
	}
	for i := 0; i < cnt; i++ {
		fmt.Fprintln(&buf, strings.Replace(flags.Code, "CNT", fmt.Sprintf("%03d", i), -1))
	}
	if flags.Postamble != "" {
		buf.WriteString(flags.Postamble)
		buf.WriteString("\n")
	}
	var title string
	if flags.Title != "" {
		title = `llOwnerSay("\nRESULT: TITLE=` + flags.Title + `");`
	}
	padding := flags.Pad
	if pad > 0 {
		padding += "integer i;\n"
	}
	if (pad&1) != 0 && pad >= 5 {
		padding += "\njump Z; @Z;\n"
		pad -= 5
	}
	if pad > 1 {
		padding += "i"
		pad -= 2
		for pad > 1 {
			padding += "+i"
			pad -= 2
		}
		padding += ";\n"
	}

	fmt.Fprintf(&buf, code, title, cnt, key.Padding, padding)
	script := buf.String()
	if flags.Show {
		fmt.Println(script)
	}
	results, err := b.Send(script)
	if err != nil {
		panic(err)
	}
	const (
		BM    = "BASE_MEM="
		TM    = "TEST_MEM="
		SZ    = "SIZE="
		TITLE = "TITLE="
	)
	for _, s := range results {
		switch {
		case strings.HasPrefix(s, BM):
			r.Base, _ = strconv.Atoi(s[len(BM):])
		case strings.HasPrefix(s, TM):
			r.Test, _ = strconv.Atoi(s[len(TM):])
		case strings.HasPrefix(s, SZ):
			r.Size, _ = strconv.ParseFloat(s[len(SZ):], 64)
		case strings.HasPrefix(s, TITLE):
			r.Title = s[len(TITLE):]
		}
	}
}

// code is the boilerplate for autobench.  It is printed with 4 positional
// parameters:
//  1. statement to print title (if any)
//  2. the number of times the CODE was repeated
//  3. the amount of padding added
//  4. instructions to pad the code size
var code = `
result(integer mem, integer count, integer padding) {
	llOwnerSay("\n");
    %s                      // Title
    if (count) {
        integer old = (integer)llLinksetDataRead("mem");
        string s = llLinksetDataRead("name");
		llOwnerSay("INFO  :COUNT=" + (string)count);
		llOwnerSay("RESULT:SIZE=" + (string)((float)(mem - old)/(float)count));
		llOwnerSay("RESULT:TEST_MEM=" + (string)mem);
		llOwnerSay("INFO  :BASE_MEM=" + (string)old);
    } else {
        llLinksetDataWrite("mem", (string)mem);
        llOwnerSay("RESULT:BASE_MEM=" + (string)mem);
    }
    llOwnerSay("INFO  :LAST_MEM=" + (string)llGetUsedMemory());
	llOwnerSay("INFO  :PADDING=" + (string)padding);
    llOwnerSay("DONE");
	return;
}

default {
    state_entry() {
		result(llGetUsedMemory(), %d, %d);
        %s                  // Padding
    }
}`

func mkVar(s string) (string, error) {
	if s == "" {
		return "", errors.New("missing variable name")
	}
	switch s[0] | ' ' {
	case 'a', 'l':
		return "list " + s, nil
	case 'f', 'g':
		return "float " + s, nil
	case 'i', 'j', 'x':
		return "integer " + s, nil
	case 'k':
		return "key " + s, nil
	case 'q', 'r':
		return "quaternion " + s, nil
	case 's':
		return "string " + s, nil
	case 'v':
		return "vector " + s, nil
	default:
		return "", fmt.Errorf("%s: invalid variable name", s)
	}
}
