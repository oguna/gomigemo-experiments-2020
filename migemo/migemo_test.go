package migemo_test

import (
	"bufio"
	"os"
	"testing"

	"github.com/oguna/gomigemo-experiments-2020/migemo"
)

func LoadTestdata() []string {
	fp, err := os.Open("../testdata/ruby-uniq.txt")
	if err != nil {
		panic(err)
	}
	defer fp.Close()
	scanner := bufio.NewScanner(fp)
	keys := make([]string, 0)
	for scanner.Scan() {
		keys = append(keys, scanner.Text())
	}
	return keys
}

func LoadMigemoDictionary() *migemo.CompactDictionaryU8 {
	fp, err := os.Open("../testdata/migemo-dict")
	if err != nil {
		panic(err)
	}
	defer fp.Close()
	dict := migemo.BuildDictionaryU8FromMigemoDictFile(fp)
	return dict
}

func BenchmarkMigemo_UTF8(b *testing.B) {
	dict := LoadMigemoDictionary()
	operator := migemo.NewRegexOperator("|", "(", ")", "[", "]", "")
	keys := LoadTestdata()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, key := range keys {
			migemo.Query(key, dict, operator)
		}
	}
}
