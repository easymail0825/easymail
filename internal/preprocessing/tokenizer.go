package preprocessing

type Tokenizer interface {
	Cut(text string, unitary bool, binary bool, ternary bool, fuzzy bool) []string
}
