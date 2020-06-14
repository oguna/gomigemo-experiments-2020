package migemo_test

import (
	"testing"
	"unicode/utf16"

	"github.com/oguna/gomigemo-experiments-2020/migemo"
)

func TestPatriciaTrie_Lookup(t *testing.T) {
	words := []string{"baby", "bad", "bank", "box", "dad", "dance"}
	keys := make([][]uint16, len(words))
	for i := 0; i < len(words); i++ {
		keys[i] = utf16.Encode([]rune(words[i]))
	}
	trie, outs := migemo.BuildLoudsPatriciaTrie(keys)
	expectedNodes := []uint32{8, 9, 10, 5, 6, 7}
	for i := 0; i < len(keys); i++ {
		actual := trie.Lookup(keys[i])
		if actual != int(expectedNodes[i]) {
			t.Errorf("word:%s expected:%d actual:%d\n", words[i], expectedNodes[i], actual)
		}
		if !outs.Get(int(expectedNodes[i])) {
			t.Errorf("%s", words[i])
		}
	}
}

func TestPatriciaTrie_Lookup_Fail(t *testing.T) {
	words := []string{"baby", "bad", "bank", "box", "dad", "dance"}
	keys := make([][]uint16, len(words))
	for i := 0; i < len(words); i++ {
		keys[i] = utf16.Encode([]rune(words[i]))
	}
	trie, _ := migemo.BuildLoudsPatriciaTrie(keys)
	wordsToFail := map[string]int{"a": -1, "d": -3, "dan": -7, "danc": -7, "dancea": -1}
	for k, v := range wordsToFail {
		s := utf16.Encode([]rune(k))
		actual := trie.Lookup(s)
		if actual != v {
			t.Errorf("word:%s expected:%d actual:%d\n", k, v, actual)
		}
	}
}

func TestPatriciaTrie_Lookup_Patricia(t *testing.T) {
	words := []string{"a", "bad", "badya"}
	keys := make([][]uint16, len(words))
	for i := 0; i < len(words); i++ {
		keys[i] = utf16.Encode([]rune(words[i]))
	}
	trie, _ := migemo.BuildLoudsPatriciaTrie(keys)
	wordsToFail := map[string]int{"bad": 3, "badya": 4, "ba": -3, "b": -3}
	for k, v := range wordsToFail {
		s := utf16.Encode([]rune(k))
		actual := trie.Lookup(s)
		if actual != v {
			t.Errorf("word:%s expected:%d actual:%d\n", k, v, actual)
		}
	}
}

func TestPatriciaTrie_ReverseLookup(t *testing.T) {
	words := []string{"baby", "bad", "bank", "box", "dad", "dance"}
	keys := make([][]uint16, len(words))
	for i := 0; i < len(words); i++ {
		keys[i] = utf16.Encode([]rune(words[i]))
	}
	trie, _ := migemo.BuildLoudsPatriciaTrie(keys)
	nodes := []uint32{8, 9, 10, 5, 6, 7}
	for i := 0; i < len(words); i++ {
		s := make([]uint16, 0)
		length := trie.ReverseLookup(nodes[i], &s)
		if length != len(keys[i]) {
			t.Errorf("invalid return value. word:%s expected:%d actual:%d", words[i], len(keys[i]), length)
		}
		if len(s) != len(keys[i]) {
			t.Errorf("invalid key length. word:%s expected:%d actual:%d", words[i], len(keys[i]), len(s))
		}
		a := string(utf16.Decode(s))
		if a != words[i] {
			t.Errorf("invalid key. expected:%s actual:%s", words[i], a)
		}
	}

}

func TestPatriciaTrie_PredictiveSearch(t *testing.T) {
	words := []string{"baby", "bad", "bank", "box", "dad", "dance"}
	keys := make([][]uint16, len(words))
	for i := 0; i < len(words); i++ {
		keys[i] = utf16.Encode([]rune(words[i]))
	}
	trie, _ := migemo.BuildLoudsPatriciaTrie(keys)
	nodes := make([]int, 0)
	trie.PredictiveSearchBreadthFirst(4, func(n int) {
		nodes = append(nodes, n)
	})
	expectedNodes := []int{4, 8, 9, 10}
	if len(nodes) != len(expectedNodes) {
		t.Error()
	}
	for i := 0; i < len(nodes); i++ {
		if nodes[i] != expectedNodes[i] {
			t.Error()
		}
	}
}
