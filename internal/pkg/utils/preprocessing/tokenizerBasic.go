package preprocessing

import (
	"strings"
	"sync"
)

// TokenizerBasic is a simple tokenizer.
type TokenizerBasic struct {
	separator map[rune]struct{}
	keepSep   bool
}

var onceBasic sync.Once

func NewTokenizerBasic(separator []string) *TokenizerBasic {
	var tokenizerBasic *TokenizerBasic
	var separatorMap = make(map[rune]struct{})
	if len(separator) == 0 {
		separator = defaultSeparator
	}
	for _, s := range separator {
		separatorMap[rune(s[0])] = struct{}{}
	}
	onceBasic.Do(func() {
		tokenizerBasic = &TokenizerBasic{
			separator: separatorMap,
		}
	})
	return tokenizerBasic
}

func (t *TokenizerBasic) String() string {
	return "basic tokenizer"
}

func (t *TokenizerBasic) Cut(text string, unitary bool, binary bool, ternary bool, fuzzy bool) (tokens []string) {
	text = strings.ReplaceAll(text, string('\uFEFF'), "")
	src := []rune(text)

	length := len(src)
	positions := make([]int, length+1)
	isMultiCharacter := false

	for i := 0; i < length; i++ {
		r := src[i]

		if _, ok := t.separator[r]; ok {
			positions[i] = 1
		}
		if src[i] > 255 {
			positions[i] = 1
			isMultiCharacter = true
		} else {
			if isMultiCharacter {
				positions[i] = 1
				isMultiCharacter = false
			}
		}
	}

	lastPosition := 0
	for i, p := range positions {
		if i == 0 {
			continue
		}
		if p == 1 {
			if _, ok := t.separator[src[lastPosition]]; ok {
				if lastPosition+1 < i {
					tokens = append(tokens, string(src[lastPosition+1:i]))
				}
			} else {
				tokens = append(tokens, string(src[lastPosition:i]))
			}
			lastPosition = i
		}
	}
	if lastPosition < length {
		if _, ok := t.separator[src[lastPosition]]; ok {
			if lastPosition+1 < length {
				tokens = append(tokens, string(src[lastPosition+1:length]))
			}
		} else {
			tokens = append(tokens, string(src[lastPosition:length]))
		}
	}
	return tokens
}
