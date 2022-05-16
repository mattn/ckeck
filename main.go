package main

import (
	"bufio"
	_ "embed"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"sort"
	"strings"
	"unicode"

	"github.com/fatih/color"
	"github.com/mattn/go-lsd"
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
	l := float64(len([]rune(s)))
	m := l
	for _, w := range words {
		mm := float64(lsd.StringDistance(w, s))
		if mm < m {
			m = mm
		}
	}
	return m / l
}

func tokenize(s string) []string {
	ts := []string{}
	prev := 0
	rs := []rune(s)
	for i := 1; i < len(rs); i++ {
		if unicode.IsSymbol(rs[prev]) != unicode.IsSymbol(rs[i]) {
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

func main() {
	var showVersion bool
	var min int
	var files wordFiles
	flag.Var(&files, "d", "word file")
	flag.IntVar(&min, "min", 4, "minimum length for words")
	flag.BoolVar(&showVersion, "V", false, "Print the version")
	flag.Parse()

	if showVersion {
		fmt.Printf("%s %s (rev: %s/%s)\n", name, version, revision, runtime.Version())
		return
	}

	if err := loadWords(files); err != nil {
		log.Fatal(err)
	}

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Split(unicodeclass.SplitClass)
	for scanner.Scan() {
		s := scanner.Text()
		for _, token := range tokenize(s) {
			if !unicode.IsLetter(rune(token[0])) {
				fmt.Fprint(color.Output, token)
			} else if len(token) < min || isSomeWords(strings.ToLower(token)) {
				fmt.Fprint(color.Output, token)
			} else if v := maybeTypo(token); v != 0 && v > 0.4 {
				fmt.Fprint(color.Output, token)
			} else {
				fmt.Fprint(color.Output, color.CyanString(token))
			}
		}
	}
}
