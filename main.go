package main

import (
	"bufio"
	_ "embed"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
	"unicode"

	"github.com/fatih/color"
	"github.com/mattn/go-lsd"
	"github.com/mattn/go-unicodeclass"
)

var words []string

//go:embed words.txt
var wordlist string

func isSomeWords(s string, m int) bool {
	for _, w := range words {
		if len(w) < m {
			continue
		}
		if s == w {
			return true
		}
		if strings.HasPrefix(s, w) {
			if isSomeWords(s[len(w):], m) {
				return true
			}
		}
	}
	return false
}

func maybeTypo(s string) float64 {
	s = strings.ToLower(s)
	l := float64(len(s))
	m := l
	for _, w := range words {
		mm := float64(lsd.StringDistance(w, s))
		if mm < m {
			m = mm
		}
	}
	return m / l
}

func main() {
	var min int
	flag.IntVar(&min, "min", 4, "minimum length for words")
	flag.Parse()

	words = strings.Split(wordlist, "\n")
	sort.Slice(words, func(i, j int) bool {
		return len(words[i]) > len(words[j])
	})

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Split(unicodeclass.SplitClass)
	for scanner.Scan() {
		s := scanner.Text()
		if !unicode.IsLetter(rune(s[0])) {
			fmt.Fprint(color.Output, s)
		} else if len(s) < min || isSomeWords(strings.ToLower(s), min) {
			fmt.Fprint(color.Output, s)
		} else if v := maybeTypo(s); v != 0 && v > 0.4 {
			fmt.Fprint(color.Output, s)
		} else {
			fmt.Fprint(color.Output, color.CyanString(s))
		}
	}
}
