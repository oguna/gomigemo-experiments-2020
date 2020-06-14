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
	fmt.Printf("%d\n", dict.IoSize())
}
