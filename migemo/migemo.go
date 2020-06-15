package migemo

import (
	"regexp"
	"strings"
)

// QueryAWord は、migemoクエリを処理する
func QueryAWord(word string, dict *CompactDictionaryU8, operator *RegexOperator) string {
	var utf32word = []rune(word)
	var generator = NewTernaryRegexGenerator(*operator)
	generator.Add(utf32word)
	var lower = strings.ToLower(word)
	if dict != nil {
		var utf8lower = []byte(string([]rune(lower)))
		dict.PredictiveSearch(utf8lower, func(word []uint8) {
			generator.Add([]rune(string(word)))
		})
	}
	var zen = ConvertHan2Zen(word)
	generator.Add([]rune(zen))
	var han = ConvertZen2Han(word)
	generator.Add([]rune(han))

	var romajiProcessor = NewRomajiProcessor2()
	var hiraganaResult = romajiProcessor.RomajiToHiraganaPredictively(lower)
	for _, a := range hiraganaResult.Suffixes {
		var hira = hiraganaResult.Prefix + a
		var utf32hira = []rune(hira)
		var utf8hira = []byte(string(utf32hira))
		generator.Add(utf32hira)
		if dict != nil {
			dict.PredictiveSearch(utf8hira, func(word []uint8) {
				generator.Add([]rune(string(word)))
			})
		}
		var kata = ConvertHira2Kata(string(utf8hira))
		generator.Add([]rune(kata))
		generator.Add([]rune(ConvertZen2Han(kata)))
	}
	return string(generator.Generate())
}

// Query は、migemoクエリを処理する
func Query(word string, dict *CompactDictionaryU8, operator *RegexOperator) string {
	if len(word) == 0 {
		return ""
	}
	words := parseQuery(word)
	results := make([]string, len(words))
	for i, w := range words {
		results[i] = QueryAWord(w, dict, operator)
	}
	return strings.Join(results, "")
}

func parseQuery(query string) []string {
	// TODO: regexpの処理は遅いため、別の実装に置き換えるべき
	var re = regexp.MustCompile("[^A-Z\\s]+|[A-Z]{2,}|([A-Z][^A-Z\\s]+)|([A-Z]\\s*$)")
	return re.FindAllString(query, -1)
}
