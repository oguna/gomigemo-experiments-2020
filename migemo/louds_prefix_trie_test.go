package migemo_test

import (
	"testing"
	"unicode/utf16"

	"github.com/oguna/gomigemo-experiments-2020/migemo"
)

func TestLoudsPrefixTrie_Build(t *testing.T) {
	words := []string{"baby", "bad", "bank", "box", "dad", "dance"}
	keys := make([][]uint16, len(words))
	for i := 0; i < len(words); i++ {
		keys[i] = utf16.Encode([]rune(words[i]))
	}
	trie, nodes := migemo.BuildLoudsPrefixTrie(keys)
	expectedNodes := []uint32{7, 8, 9, 5, 10, 11}
	for i := 0; i < len(nodes); i++ {
		if nodes[i] != expectedNodes[i] {
			t.Errorf("word:%s expected:%d actual:%d\n", words[i], expectedNodes[i], nodes[i])
		}
	}
	for i := 0; i < len(nodes); i++ {
		actual := uint32(trie.Lookup(keys[i]))
		if actual != expectedNodes[i] {
			t.Errorf("word:%s expected:%d actual:%d\n", words[i], expectedNodes[i], actual)
		}
	}
	wordsToFail := map[string]int{"dan": -11, "danc": -11, "dancea": -1}
	for k, v := range wordsToFail {
		s := utf16.Encode([]rune(k))
		actual := trie.Lookup(s)
		if actual != v {
			t.Errorf("word:%s expected:%d actual:%d\n", k, v, actual)
		}
	}
}
func TestLoudsPrefixTrie_GetTail(t *testing.T) {
	words := []string{"baby", "bad", "bank", "box", "dad", "dance"}
	keys := make([][]uint16, len(words))
	for i := 0; i < len(words); i++ {
		keys[i] = utf16.Encode([]rune(words[i]))
	}
	trie, _ := migemo.BuildLoudsPrefixTrie(keys)
	data := map[int]string{
		5:  "x",
		7:  "y",
		9:  "k",
		11: "ce",
	}
	for k, v := range data {
		tail := trie.GetTail(k)
		if len(tail) != len(v) {
			t.Errorf("invalid length. word:%s expected:%d actual:%d", v, len(v), len(tail))
		}
		for i := 0; i < len(tail); i++ {
			if tail[i] != uint16(v[i]) {
				t.Errorf("invalid chars. word:%s", v)
			}
		}
	}
}

func TestLoudsPrefixTrie_ExtractTail(t *testing.T) {
	words := []string{"a", "aaa", "b", "cc"}
	keys := make([][]uint16, len(words))
	for i := 0; i < len(words); i++ {
		keys[i] = utf16.Encode([]rune(words[i]))
	}
	tails := migemo.ExtractTailU16Strings(keys)
	expectedTails := []uint32{0, 1, 0, 1}
	if len(tails) != len(expectedTails) {
		t.Error()
	}
	for i := 0; i < len(tails); i++ {
		if tails[i] != expectedTails[i] {
			t.Fatalf("#%d expected:%d actual:%d", i, expectedTails[i], tails[i])
		}
	}
}

func TestLoudsPrefixTrie_ReverseLookup(t *testing.T) {
	words := []string{"baby", "bad", "bank", "box", "dad", "dance"}
	keys := make([][]uint16, len(words))
	for i := 0; i < len(words); i++ {
		keys[i] = utf16.Encode([]rune(words[i]))
	}
	trie, nodes := migemo.BuildLoudsPrefixTrie(keys)
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

func TestLoudsPrefixTrie_PredictiveSearch(t *testing.T) {
	words := []string{"baby", "bad", "bank", "box", "dad", "dance"}
	keys := make([][]uint16, len(words))
	for i := 0; i < len(words); i++ {
		keys[i] = utf16.Encode([]rune(words[i]))
	}
	trie, _ := migemo.BuildLoudsPrefixTrie(keys)
	nodes := make([]int, 0)
	trie.PredictiveSearchBreadthFirst(3, func(n int) {
		nodes = append(nodes, n)
	})
	expectedNodes := []int{3, 6, 10, 11}
	for i := 0; i < len(expectedNodes); i++ {
		if nodes[i] != expectedNodes[i] {
			t.Error()
		}
	}
}
