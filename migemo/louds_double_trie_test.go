package migemo_test

import (
	"testing"
	"unicode/utf16"

	"github.com/oguna/gomigemo-experiments-2020/migemo"
)

func TestLoudsDoubleTrie_Build(t *testing.T) {
	words := []string{"baby", "bad", "bank", "box", "dad", "dance"}
	keys := make([][]uint16, len(words))
	for i := 0; i < len(words); i++ {
		keys[i] = utf16.Encode([]rune(words[i]))
	}
	trie, nodes := migemo.BuildLoudsDoubleTrie(keys)
	expectedNodes := []uint32{7, 8, 9, 5, 10, 11}
	for i := 0; i < len(nodes); i++ {
		if nodes[i] != expectedNodes[i] {
			t.Errorf("word:%s expected:%d actual:%d\n", words[i], expectedNodes[i], nodes[i])
		}
	}
	for i := 0; i < len(nodes); i++ {
		actual := trie.Lookup(keys[i])
		if actual != int(expectedNodes[i]) {
			t.Errorf("word:%s expected:%d actual:%d\n", words[i], expectedNodes[i], actual)
		}
	}
	wordsToFail := map[string]int{"dan": -11, "danc": -11, "dancea": -1}
	for k, v := range wordsToFail {
		s := utf16.Encode([]rune(k))
		actual := trie.Lookup(s)
		if v != actual {
			t.Errorf("word:%s expected:%d actual:%d\n", k, v, actual)
		}
	}
}
func TestLoudsDoubleTrie_ReverseLookup(t *testing.T) {
	words := []string{"baby", "bad", "bank", "box", "dad", "dance"}
	keys := make([][]uint16, len(words))
	for i := 0; i < len(words); i++ {
		keys[i] = utf16.Encode([]rune(words[i]))
	}
	trie, nodes := migemo.BuildLoudsDoubleTrie(keys)
	a := make([]uint16, 0)
	for i := 0; i < len(nodes); i++ {
		a = a[:0]
		wordLen := trie.ReverseLookup(nodes[i], &a)
		s := string(utf16.Decode(a))
		if len(s) != len(words[i]) {
			t.Errorf("%s", words[i])
		}
		if len(keys[i]) != wordLen {
			t.Errorf("%s", words[i])
		}
		for j := 0; j < len(s); j++ {
			if s[j] != words[i][j] {
				t.Errorf("node:%d expected:%s actual:%s", i, words[i], s)
				break
			}
		}
	}
}
