package main

import (
	"bufio"
	"bytes"
	_ "embed"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime"
	"sort"
	"strings"
	"unicode"

	"github.com/fatih/color"
	"github.com/mattn/go-jsd"
	"github.com/mattn/go-unicodeclass"
)

const name = "ckeck"

const version = "0.0.1"

var revision = "HEAD"

var words []string

//go:embed words.txt
var wordlist string

func isSomeWords(s string) bool {
	for _, w := range words {
		if s == w {
			return true
		}
		if len(w) < len(s) && strings.HasPrefix(s, w) {
			if isSomeWords(s[len(w):]) {
				return true
			}
		}
	}
	return false
}

func maybeTypo(s string) float64 {
	s = strings.ToLower(s)
	m := 0.0
	for _, w := range words {
		mm := jsd.StringDistance(w, s)
		if mm > m {
			m = mm
		}
	}
	return m
}

func tokenize(s string) []string {
	ts := []string{}
	prev := 0
	rs := []rune(s)
	for i := 1; i < len(rs); i++ {
		if unicode.IsSpace(rs[prev]) != unicode.IsSpace(rs[i]) || unicode.IsSymbol(rs[prev]) != unicode.IsSymbol(rs[i]) {
			ts = append(ts, string(rs[prev:i]))
			prev = i
		}
	}
	ts = append(ts, string(rs[prev:]))
	return ts
}

func loadWords(files []string) error {
	// built-in
	words = strings.Split(wordlist, "\n")

	for _, file := range files {
		f, err := os.Open(file)
		if err != nil {
			return err
		}
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			words = append(words, scanner.Text())
		}
		f.Close()
	}
	for i := 0; i < len(words); i++ {
		if len(words[i]) == 0 || words[i][0] == '#' {
			words = words[:i+copy(words[i:], words[i+1:])]
		}
	}
	sort.Slice(words, func(i, j int) bool {
		return len(words[i]) > len(words[j])
	})
	for i := 1; i < len(words); i++ {
		if words[i-1] == words[i] {
			words = words[:i+copy(words[i:], words[i+1:])]
		}
	}
	return nil
}

type wordFiles []string

func (i *wordFiles) String() string {
	return "word file"
}

func (i *wordFiles) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func run() int {
	var showVersion bool
	var errlist bool
	var min int
	var files wordFiles
	flag.Var(&files, "d", "word file")
	flag.IntVar(&min, "min", 4, "minimum length for words")
	flag.BoolVar(&errlist, "u", false, "show error list")
	flag.BoolVar(&showVersion, "V", false, "Print the version")
	flag.Parse()

	if showVersion {
		fmt.Printf("%s %s (rev: %s/%s)\n", name, version, revision, runtime.Version())
		return 0
	}

	var in io.Reader = os.Stdin
	var out io.Writer = color.Output
	var filename string = "stdin"

	if flag.NArg() == 1 {
		filename = flag.Arg(0)
		f, err := os.Open(filename)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		defer f.Close()
		in = f
	} else if flag.NArg() > 1 {
		flag.Usage()
		return 1
	}

	if err := loadWords(files); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	scanner := bufio.NewScanner(in)

	code := 0
	nl := 1
	for scanner.Scan() {
		var buf bytes.Buffer
		wrong := -1
		col := 1

		s := scanner.Text()
		for _, tok := range unicodeclass.Split(s) {
			for _, token := range tokenize(tok) {
				if len(token) > 0 && !unicode.IsLetter(rune(token[0])) {
					fmt.Fprint(&buf, token)
				} else if len(token) < min || isSomeWords(strings.ToLower(token)) {
					fmt.Fprint(&buf, token)
				} else if v := maybeTypo(token); v != 0 && v < 0.8 {
					fmt.Fprint(&buf, token)
				} else {
					fmt.Fprint(&buf, color.CyanString(token))
					if wrong == -1 {
						wrong = col
					}
				}
				col += len(token)
			}
		}
		if errlist {
			if wrong != -1 {
				fmt.Fprintf(out, "%s:%d:%d:%s\n", filename, nl, wrong, buf.String())
				code = 1
			}
		} else {
			fmt.Fprintln(out, buf.String())
		}

		nl += 1
	}

	return code
}

func main() {
	os.Exit(run())
}
