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
	"github.com/mattn/go-unicodeclass"
)

var words []string

//go:embed words.txt
var wordlist string

func init() {
	words = strings.Split(wordlist, "\n")
	sort.Slice(words, func(i, j int) bool {
		return len(words[i]) > len(words[j])
	})
}

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

func main() {
	var min int
	flag.IntVar(&min, "min", 4, "minimum length for words")
	flag.Parse()
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Split(unicodeclass.SplitClass)
	for scanner.Scan() {
		s := scanner.Text()
		if !unicode.IsLetter(rune(s[0])) {
			fmt.Fprint(color.Output, s)
		} else {
			if len(s) < min || isSomeWords(s, min) {
				fmt.Fprint(color.Output, s)
			} else {
				fmt.Fprint(color.Output, color.CyanString(s))
			}
		}
	}
}
