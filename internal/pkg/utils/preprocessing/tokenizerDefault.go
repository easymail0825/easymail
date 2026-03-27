package preprocessing

import (
	"github.com/go-ego/gse"
	"strings"
	"sync"
)

// TokenizerDefault singled
type TokenizerDefault struct {
	useHmm       bool
	segmenter    gse.Segmenter
	stopWords    map[string]struct{}
	excludeChars map[string]struct{}
}

var onceDefault sync.Once

func NewTokenizerDefault(dictList []string, stopWords map[string]struct{}, excludeChars map[string]struct{}) *TokenizerDefault {
	var tokenizerDefault *TokenizerDefault
	onceDefault.Do(func() {
		var segmenter gse.Segmenter
		err := segmenter.LoadDict(dictList...)
		if err != nil {
			panic(err)
		}
		tokenizerDefault = &TokenizerDefault{
			useHmm:       true,
			segmenter:    segmenter,
			stopWords:    stopWords,
			excludeChars: excludeChars,
		}
	})
	return tokenizerDefault
}

func (t *TokenizerDefault) String() string {
	return "default tokenizer"
}

func (t *TokenizerDefault) Cut(text string, unitary bool, binary bool, ternary bool, fuzzy bool) (tokens []string) {
	text = strings.ReplaceAll(text, string('\uFEFF'), "")
	data := make([]string, 0)
	for _, s := range t.segmenter.Cut(text, t.useHmm) {
		s := strings.TrimSpace(s)
		s = strings.ReplaceAll(s, "\n", "")
		s = strings.ReplaceAll(s, "\r", "")

		r := []rune(s)
		if len(r) < 1 {
			continue
		}
		if r[0] < 255 && !((r[0] >= 65 && r[0] <= 90) || (r[0] >= 97 && r[0] <= 122)) {
			continue
		}

		matched := false
		for c := range t.excludeChars {
			if strings.Contains(s, c) {
				matched = true
				break
			}
		}
		if matched {
			continue
		}
		data = append(data, s)

	}

	if binary {
		for _, token := range CombineToken(data, 2, true, fuzzy) {
			if _, ok := t.stopWords[token]; !ok {
				tokens = append(tokens, token)
			}
		}
	}
	if ternary {
		for _, token := range CombineToken(data, 3, true, fuzzy) {
			if _, ok := t.stopWords[token]; !ok {
				tokens = append(tokens, token)
			}
		}
	}

	if unitary {
		for _, token := range data {
			if _, ok := t.stopWords[token]; !ok {
				tokens = append(tokens, token)
			}
		}
	}
	return
}

// CombineToken This function takes in a slice of strings, a step size, a boolean value indicating whether to add a spacer,
// and a boolean value indicating whether  to chain tokens together.
// It then returns a slice of strings.
func CombineToken(src []string, step int, addSpacer bool, chain bool) []string {
	var buf strings.Builder
	l := len(src)
	for i := 0; i < l-1; i++ {
		for j := 0; j < step && i+j < l; j++ {
			masked := false
			if step > 2 && chain && j != 0 && j != step-1 {
				masked = true
			}
			if masked {
				buf.WriteString("*")
			} else {
				buf.WriteString(src[i+j])
			}
			if addSpacer && (j+1)%step != 0 {
				buf.WriteString(" ")
			}
		}
		if i != l-2 {
			buf.WriteString("&&")
		}
	}
	return strings.Split(buf.String(), "&&")
}
