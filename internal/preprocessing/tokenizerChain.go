package preprocessing

import (
	"bytes"
	"io"
	"strings"
	"sync"
	"unicode"
)

type TokenizerChain struct {
	stopWords map[string]struct{}
}

var onceFuzzy sync.Once

func NewTokenizerFuzzy(stopWord map[string]struct{}) *TokenizerChain {
	var tokenizerChain *TokenizerChain
	onceFuzzy.Do(func() {
		tokenizerChain = &TokenizerChain{
			stopWords: stopWord,
		}
	})
	return tokenizerChain
}

// getMultibyte  get multibyte text from the src
func (t *TokenizerChain) getMultibyte(src string, unitary bool, binary bool, ternary bool, chain bool) []string {
	srcRunes := []rune(src)
	srcLength := len(srcRunes)
	var token = make([]byte, 0, 0)
	buffer := bytes.NewBuffer(token)

	// non-ascii characters
	var distRunes = make([]rune, 0)
	for i := 0; i < srcLength; i++ {
		if srcRunes[i] > 255 {
			distRunes = append(distRunes, srcRunes[i])
		}
	}

	// unitary tokenize
	if unitary {
		for _, c := range distRunes {
			buffer.WriteRune(c)
			buffer.WriteRune('|')
		}
	}

	// binary tokenize
	if binary {
		for i := 0; i < len(distRunes)-1; i++ {
			buffer.WriteRune(distRunes[i])
			buffer.WriteRune(distRunes[i+1])
			buffer.WriteRune('|')
		}
	}

	// ternary tokenize
	if ternary {
		for i := 0; i < len(distRunes)-2; i++ {
			buffer.WriteRune(distRunes[i])
			buffer.WriteRune(distRunes[i+1])
			buffer.WriteRune(distRunes[i+2])
			buffer.WriteRune('|')
		}

		// ternary tokenize with chain
		if chain {
			for i := 0; i < len(distRunes)-2; i++ {
				buffer.WriteRune(distRunes[i])
				buffer.WriteRune('*')
				buffer.WriteRune(distRunes[i+2])
				buffer.WriteRune('|')
			}
		}
	}

	var dist = buffer.String()
	dist = strings.TrimRight(dist, "|")
	return strings.Split(dist, "|")
}

// TokenizeAscii split text with ascii characters and single character tokenization
func (t *TokenizerChain) TokenizeAscii(b string, unitary bool, binary bool, ternary bool, fuzzy bool) []string {
	var tokenBuf = bytes.NewBufferString(b)
	var distBuf = bytes.NewBuffer(make([]byte, 0))

	for {
		c, _, err := tokenBuf.ReadRune()
		if c == 0 && err == io.EOF {
			break
		}

		if c == '\r' || c == '\n' || c == '\uFEFF' {
			distBuf.WriteRune('|')
			continue
		}
		if c >= 255 || !unicode.IsLetter(c) {
			distBuf.WriteRune('|')
			continue
		}

		if c > 64 && c < 91 {
			distBuf.WriteRune(c + 32)
			continue
		}
		distBuf.WriteRune(c)
	}

	var tokenString = distBuf.String()
	tokenString = strings.TrimRight(tokenString, "|")
	var tmpList = make([]string, 0)
	tl := strings.Split(tokenString, "|")

	if unitary {
		for _, token := range tl {
			lowerToken := strings.ToLower(token)
			if len(token) < 2 || len(token) > 16 {
				continue
			} else if _, ok := t.stopWords[lowerToken]; ok {
				continue
			} else {
				tmpList = append(tmpList, lowerToken)
			}
		}
	}

	if binary {
		for i := 0; i < len(tl)-1; i++ {
			if len(tl[i]) > 1 && len(tl[i+1]) > 1 {
				t := tl[i] + " " + tl[i+1]
				tmpList = append(tmpList, t)
			}
		}
	}

	if ternary {
		for i := 0; i < len(tl)-2; i++ {
			if len(tl[i]) > 1 && len(tl[i+1]) > 1 && len(tl[i+2]) > 1 {
				t := tl[i] + " " + tl[i+1] + " " + tl[i+2]
				tmpList = append(tmpList, t)
			}
		}

		if fuzzy {
			for i := 0; i < len(tl)-2; i++ {
				if len(tl[i]) > 1 && len(tl[i+1]) > 1 {
					t := tl[i] + " * " + tl[i+2]
					tmpList = append(tmpList, t)
				}
			}
		}
	}
	return tmpList
}

func (t *TokenizerChain) Cut(text string, unitary bool, binary bool, ternary bool, fuzzy bool) (tokenList []string) {
	var asciiTokens = t.TokenizeAscii(text, unitary, binary, ternary, fuzzy)
	var multiTokens = t.getMultibyte(text, unitary, binary, ternary, fuzzy)
	for _, token := range asciiTokens {
		if len(token) > 1 {
			tokenList = append(tokenList, token)
		}
	}

	for _, token := range multiTokens {
		if len(token) >= 1 {
			if _, ok := t.stopWords[token]; ok {
				continue
			}
			tokenList = append(tokenList, token)
		}
	}
	return tokenList
}
