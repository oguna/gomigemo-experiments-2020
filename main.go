package main

import (
	"fmt"
	"os"

	"github.com/oguna/gomigemo-experiments-2020/migemo"
)

func main() {
	fp, err := os.Open("testdata/migemo-dict")
	if err != nil {
		panic(err)
	}
	defer fp.Close()
	dict := migemo.BuildDictionaryFromMigemoDictFile(fp)
	fmt.Printf("IoSize: %d\n", dict.IoSize())
	a, b := dict.NodeSize()
	fmt.Printf("#Nodes: %d %d\n", a, b)
	operator := migemo.NewRegexOperator("|", "(", ")", "[", "]", "")
	fmt.Printf(migemo.Query("kensaku", dict, operator))
}
