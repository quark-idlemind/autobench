// Program gen is used to generate benchmark tests for the automate program.
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
	Preamble  string   `getopt:"--preamble=PREMBLE Make STR the tests's preamble"`
	Postamble string   `getopt:"--postamble=POSTAMBLE Make STR the tests's postamble"`
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
	// Get the base result
	if err := runScript(b, 0, &r); err != nil {
		errf("%v\n", err)
	}

	if err := runScript(b, 4, &r); err != nil {
		errf("%v\n", err)
	}
	var cnt int
	if r.Size == 0 {
		cnt = 512
	} else {
		maxMem := 60*1024 - r.Base
		cnt = maxMem / (int(r.Size) + 64)
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
	if err := runScript(b, cnt, &r); err != nil {
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

func runScript(b *bench.Bench, cnt int, r *Results) error {
	var buf strings.Builder
	if flags.Preamble != "" {
		buf.WriteString(flags.Preamble)
		buf.WriteString("\n")
	}
	for i := 0; i < cnt; i++ {
		fmt.Fprintln(&buf, strings.Replace(flags.Code, "COUNT", fmt.Sprintf("%03d", i), -1))
	}
	if flags.Postamble != "" {
		buf.WriteString(flags.Postamble)
		buf.WriteString("\n")
	}
	var title string
	if flags.Title != "" {
		title = `llOwnerSay("\nRESULT: TITLE=` + flags.Title + `");`
	}
	fmt.Fprintf(&buf, code, title, cnt, flags.Pad)
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
