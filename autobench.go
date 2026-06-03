package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

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
}{}

func init() {
	flags.ChatLog, _ = bench.GetChatLog()
	flags.Source, _ = bench.GetLSLPath()
}

func errf(format string, v ...any) {
	fmt.Fprintf(os.Stderr, format, v...)
	os.Exit(1)
}

const plusminus = "±"

func main() {
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
		low := 0
		high := 512
		mid := (high - low) / 2
		pad := 0
		if flags.IPad != 0 {
			if err := runScript(b, 0, flags.IPad-1, &r); err != nil {
				errf("%v\n", err)
			}
			base := r.Base
			if err := runScript(b, 0, flags.IPad, &r); err != nil {
				errf("%v\n", err)
			}
			if base != r.Base {
				pad = flags.IPad
			} else {
				mid = flags.IPad
			}
		} else {
			if err := runScript(b, 0, 0, &r); err != nil {
				errf("%v\n", err)
			}
		}
		if pad == 0 {
			base := r.Base
			for high-low > 1 {
				if err := runScript(b, 0, mid, &r); err != nil {
					errf("%v\n", err)
				}
				if base == r.Base {
					low = mid
				} else {
					high = mid
				}
				mid = low + (high-low)/2
			}
			if low > 4 {
				if err := runScript(b, 0, low+1, &r); err != nil {
					errf("%v\n", err)
				}
				if base == r.Base {
					low++
				}
			}
			pad = low + 1 // bump us over
		}

		if err := runScript(b, 1, pad, &r); err != nil {
			errf("%v\n", err)
		}
		low = 0
		high = 512
		mid = (high - low) / 2
		test := r.Test
		for high-low > 1 {
			if err := runScript(b, 1, mid+pad, &r); err != nil {
				errf("%v\n", err)
			}
			if test == r.Test {
				low = mid
			} else {
				high = mid
			}
			mid = low + (high-low)/2
		}
		if low > 4 {
			if err := runScript(b, 1, low+1+pad, &r); err != nil {
				errf("%v\n", err)
			}
			if test == r.Test {
				low++
			}
		}
		low += 1 // bumps us over
		if err := runScript(b, 1, low+pad, &r); err != nil {
			errf("%v\n", err)
		}

		if r.Title != "" {
			fmt.Printf("Title: %s\n", r.Title)
		}

		fmt.Printf("Padding: %d\n", pad)
		fmt.Printf("Size: %d\n", r.Test - r.Base -low)
		return
	}

	// Get the base result
	if err := runScript(b, 0, 0, &r); err != nil {
		errf("%v\n", err)
	}

	D := 4
	if err := runScript(b, D, 0, &r); err != nil {
		errf("%v\n", err)
	}
/*
	if r.Size == 0 {
		D = 8
		if err := runScript(b, D, 0, &r); err != nil {
			errf("%v\n", err)
		}
	}
	var cnt int
*/
	if r.Size == 0 {
		cnt = 512
	} else {
		maxMem := 60*1024 - r.Base
		cnt = maxMem / (int(r.Size) + 256 / D)
	}
	switch {
	case cnt == 512:
	case cnt >= 256:
		cnt = 256
	case cnt >= 128:
		cnt = 128
	case cnt >= 64:
		cnt = 64
	case cnt >= 32:
		cnt = 32
	case cnt >= 16:
		cnt = 16
	case cnt >= 8:
		cnt = 8
	case cnt >= 4:
		cnt = 4
	case cnt >= 2:
		cnt = 2
	case cnt >= 1:
		cnt = 1
	default:
		fmt.Print("Unable to benchmark\n")
		return
	}
	if err := runScript(b, cnt, 0, &r); err != nil {
		errf("%v\n", err)
	}
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

func runScript(b *bench.Bench, cnt, pad int, r *Results) error {
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

	fmt.Fprintf(&buf, code, title, cnt, padding)
	script := buf.String()
	if flags.Show {
		fmt.Println(script)
	}
	results, err := b.Send(script)
	if err != nil {
		return err
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
	return nil
}

var code = `
report(string msg) {
	llOwnerSay("\nRESULT:" + msg);
}

result(integer mem, integer count) {
    %s
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
        result(llGetUsedMemory(), %d);
	return;
	%s
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
