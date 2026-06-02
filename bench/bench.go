package bench

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/quark-idlemind/lslcomp"
)

var (
	ErrFailed = errors.New("reading chatlog failed")
	ErrNoHome = errors.New("$HOME not set")
)

var (
	home = os.Getenv("HOME")
	dirs = []string{
		"/Library/Application Support/Firestorm",
		"/Library/Application Support/SecondLife",
	}

	adir    = home + "/.automate"
	who     = adir + "/who"
	loghome = adir + "/home"
	source  = adir + "/source"
)

func GetChatLog() (string, error) {
	if home == "" {
		return "", ErrNoHome
	}
	lh := readstring(loghome)
	if lh == "" {
		lh = home
	}
	user := readstring(who)
	if user == "" {
		return "", fmt.Errorf("%s: missing or empty", who)
	}
	if strings.Contains(user, "/") {
		return user, nil
	}
	for _, dir := range dirs {
		path := lh + dir + "/" + user + "/chat.txt"
		if isFile(path) {
			return path, nil
		}
		path = lh + dir + "/" + user + "_resident/chat.txt"
		if isFile(path) {
			return path, nil
		}
	}
	return "", fmt.Errorf("could not find chatlog for %s", user)
}

func GetLSLPath() (string, error) {
	if home == "" {
		return "", ErrNoHome
	}
	file, err := os.Readlink(source)
	if err != nil {
		return "", err
	}
	if fi, err := os.Stat(file); err != nil {
		return "", err
	} else if !fi.Mode().IsRegular() {
		return "", fmt.Errorf("%s: not a regular file", file)
	} else if fp, err := os.OpenFile(file, os.O_RDWR, 0644); err != nil {
		return "", err
	} else {
		fp.Close()
	}
	return file, nil
}

func readstring(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func isFile(file string) bool {
	fi, err := os.Stat(file)
	if err != nil {
		return false
	}
	if !fi.Mode().IsRegular() {
		return false
	}
	return true
}

type Bench struct {
	LSLPath string
	ch      chan result
}

type result struct {
	done bool
	msg  string
}

func New(path, chatlog string) (*Bench, error) {
	fp, err := os.Open(chatlog)
	if err != nil {
		return nil, err
	}
	fp.Seek(0, 2)

	b := &Bench{
		LSLPath: path,
		ch:      make(chan result, 1),
	}
	go b.monitor(bufio.NewReader(fp))
	return b, nil
}

// Send sends the script to Second Life.  Send returns an error if either the
// script will not compile or it cannot send the script to LSL
func (b *Bench) Send(script string) ([]string, error) {
	if err := lslcomp.Parse(script); err != nil {
		return nil, fmt.Errorf("%s", strings.TrimSpace(err.Error()))
	}
	fp, err := os.OpenFile(b.LSLPath, os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}
	fp.Truncate(0)
	fp.Write([]byte(script))
	if err := fp.Close(); err != nil {
		return nil, err
	}
	var results []string
	for r := range b.ch {
		if r.done {
			return results, nil
		}
		results = append(results, r.msg)
	}
	return results, ErrFailed
}

func (b *Bench) monitor(r *bufio.Reader) {
	defer close(b.ch)
	for {
		line, err := r.ReadBytes('\n')
		if len(line) > 0 {
			sline := string(line)
			switch {
			case strings.Contains(sline, "DONE"):
				b.ch <- result{done: true}
			case strings.Contains(sline, "RESULT:"):
				b.ch <- result{msg: strings.TrimSpace(sline[strings.Index(sline, "RESULT:")+7:])}
			}
		}
		if err == io.EOF {
			time.Sleep(500 * time.Millisecond)
		} else if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return
		}
	}
}
