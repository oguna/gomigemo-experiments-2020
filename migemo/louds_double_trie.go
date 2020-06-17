package migemo

import (
	"sort"
	"unicode/utf16"
)

// LoudsDoubleTrie は、あるノード以降において分岐がない場合、
// それ以降の文字列を別のトライに格納し容量を削減する．
type LoudsDoubleTrie struct {
	prefixTrie    *LoudsTrieU16
	tailTrie      *LoudsTrieU16
	outs          *BitVector
	linkBitVector *BitVector
	linkArray     []uint32
}

// Lookup is ...
func (trie *LoudsDoubleTrie) Lookup(key []uint16) int {
	var nodeIndex int = 1
	for i, c := range key {
		nodeIndex = trie.prefixTrie.Traverse(uint32(nodeIndex), c)
		if nodeIndex == -1 {
			break
		}
		if trie.linkBitVector.Get(uint32(nodeIndex)) {
			// Tailを持つので、Tailトライを探索
			linkIndex := trie.linkBitVector.Rank(uint(nodeIndex), true)
			tailNode := trie.linkArray[linkIndex]
			i++
			for tailNode > 1 {
				if len(key) <= i {
					return -nodeIndex
				}
				if key[i] != trie.tailTrie.edges[tailNode] {
					return -1
				}
				tailNode = trie.tailTrie.Parent(tailNode)
				i++
			}
			if i != len(key) {
				return -1
			}
			return nodeIndex
		}
	}
	if nodeIndex >= 0 {
		return nodeIndex
	}
	return -1
}

// ReverseLookup は、指定されたノード番号からキーを復元する
func (trie *LoudsDoubleTrie) ReverseLookup(index uint32, key *[]uint16) int {
	prefixLength := trie.prefixTrie.ReverseLookup(index, key)
	// 指定されたノード番号がTailへのリンクを持つなら、末尾にTailを追加
	if trie.linkBitVector.Get(index) {
		tailStartPos := len(*key)
		tailNode := trie.linkArray[trie.linkBitVector.Rank(uint(index), true)]
		tailLength := trie.tailTrie.ReverseLookup(tailNode, key)
		for i, j := tailStartPos, len(*key)-1; i < j; i, j = i+1, j-1 {
			(*key)[i], (*key)[j] = (*key)[j], (*key)[i]
		}
		return prefixLength + tailLength
	}
	return prefixLength
}

// PredictiveSearchBreadthFirst は、指定したノードから葉の方向に全てのノードを幅優先で巡る．
func (trie *LoudsDoubleTrie) PredictiveSearchBreadthFirst(node int, f func(int)) {
	trie.prefixTrie.PredictiveSearchBreadthFirst(node, f)
}

// BuildLoudsDoubleTrie は、ソート済みのkeysからトライを作成する
func BuildLoudsDoubleTrie(keys [][]uint16) (*LoudsDoubleTrie, []uint32) {
	numOfTailWord := 0
	// TAIL文字列を抽出
	tailList := ExtractTailU16Strings(keys)
	tailStrings := make(map[string]struct{})
	for i := 0; i < len(tailList); i++ {
		if tailList[i] > 0 {
			numOfTailWord++
			s := keys[i]
			tail := s[len(s)-int(tailList[i]):]
			tailStrings[string(utf16.Decode(tail))] = struct{}{}
		}
	}
	tailStringList := make([][]uint16, 0)
	for k := range tailStrings {
		s := utf16.Encode([]rune(k))
		for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
			s[i], s[j] = s[j], s[i]
		}
		tailStringList = append(tailStringList, s)
	}
	// Tailトライを作成
	sort.Slice(tailStringList, func(i, j int) bool { return CompareUtf16String(tailStringList[i], tailStringList[j]) < 0 })
	tailTrie, _ := BuildLoudsTrie(tailStringList)
	// Prefixトライを作成
	prefixStringList := make([][]uint16, len(keys))
	for i := 0; i < len(keys); i++ {
		s := keys[i]
		prefixStringList[i] = s[:len(s)-int(tailList[i])]
	}
	prefixTrie, prefixStringNodes := BuildLoudsTrie(prefixStringList)
	outs := NewBitListWithSize(prefixTrie.Size() + 2)
	for _, e := range prefixStringNodes {
		outs.Set(int(e), true)
	}
	// Linkを作成
	linkArray := make([]uint32, 0, numOfTailWord)
	linkBitList := NewBitListWithSize(prefixTrie.Size() + 2)
	for i := 0; i < len(prefixStringNodes); i++ {
		if tailList[i] > 0 {
			linkBitList.Set(int(prefixStringNodes[i]), true)
		}
	}
	a := make([]int32, prefixTrie.Size()+2)
	for i := 0; i < len(a); i++ {
		a[i] = -1
	}
	for i, e := range prefixStringNodes {
		a[e] = int32(i)
	}
	for i := 0; i < linkBitList.Size; i++ {
		if linkBitList.Get(i) {
			foundTail := int(a[i])
			if foundTail >= 0 {
				s := keys[foundTail]
				s2 := s[len(s)-int(tailList[foundTail]):]
				tailString := make([]uint16, len(s2))
				copy(tailString, s2)
				for i, j := 0, len(tailString)-1; i < j; i, j = i+1, j-1 {
					tailString[i], tailString[j] = tailString[j], tailString[i]
				}
				tailNode := tailTrie.Lookup(tailString)
				linkArray = append(linkArray, uint32(tailNode))
			}
		}
	}
	// インスタンスを生成
	trie := &LoudsDoubleTrie{
		prefixTrie:    prefixTrie,
		tailTrie:      tailTrie,
		outs:          NewBitVector(outs.Words, uint32(outs.Size)),
		linkArray:     linkArray,
		linkBitVector: NewBitVector(linkBitList.Words, uint32(linkBitList.Size)),
	}
	return trie, prefixStringNodes
}

// Size is ...
func (trie *LoudsDoubleTrie) Size() int {
	return trie.prefixTrie.Size()
}

// NumOfNodes is ...
func (trie *LoudsDoubleTrie) NumOfNodes() int {
	return trie.prefixTrie.Size() + trie.tailTrie.Size()
}

// IoSize is ...
func (trie *LoudsDoubleTrie) IoSize() int {
	return trie.prefixTrie.IoSize() + trie.tailTrie.IoSize() + trie.outs.IoSize() + trie.linkBitVector.IoSize() + IoSizeUint32Array(trie.linkArray)
}
