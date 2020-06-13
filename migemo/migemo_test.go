package migemo_test

import (
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/oguna/gomigemo-experiments-2020/migemo"
)

func LoadTestdata() {

}

func BenchmarkMigemo_UTF8(b *testing.B) {
	bytes := ioutil.ReadFile("../testdata/migemo-compact-dict")
	dict := migemo.NewCompactDictionary(bytes)
	operator = migemo.NewRegexOperator("|", "(", ")", "[", "]", "")
	b.ResetTimer()
	r := migemo.Query("kensaku", dict, operator)
	fmt.Printf("%s", r)
}
