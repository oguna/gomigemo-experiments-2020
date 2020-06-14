package migemo

// LoudsPrefixTrie は、あるノード以降において分岐がない場合、
// TAIL配列に文字列を格納することで、トライのノード数を削減している．
type LoudsPrefixTrie struct {
	// trie は、トライのノードのうち分岐するものを格納する
	trie *LoudsTrieU16
	// outs は、そのノードがキーであるかを格納する
	outs *BitVector
	// links は、そのノードがTAILへのリンクを持つのかを格納する
	links *BitVector
	// tailBits は、TAIL配列の区切りを格納する
	tailBits *BitVector
	// tailChars は、トライにおいて分岐のない文字列を格納するTAIL配列
	tailChars []uint16
}

// Lookup is ...
func (trie *LoudsPrefixTrie) Lookup(key []uint16) int {
	var nodeIndex int = 1
	cursor := 0
	for cursor < len(key) {
		c := key[cursor]
		nodeIndex = trie.trie.Traverse(uint32(nodeIndex), c)
		if nodeIndex == -1 {
			return -1
		}
		if trie.links.Get(uint32(nodeIndex)) {
			cursor++
			tail := trie.GetTail(nodeIndex)
			if len(key) > len(tail)+cursor {
				return -1
			}
			end := len(tail)
			if len(key)-cursor < end {
				end = len(key) - cursor
			}
			j := 0
			for ; j < end; j++ {
				if key[cursor+j] != tail[j] {
					return -1
				}
			}
			if len(tail) == j {
				return nodeIndex
			}
			return -nodeIndex
		}
		cursor++
	}
	return nodeIndex
}

// ReverseLookup は、指定されたノード番号からキーを復元する
func (trie *LoudsPrefixTrie) ReverseLookup(index uint32, key *[]uint16) int {
	prefixLength := trie.trie.ReverseLookup(index, key)
	// 指定されたノード番号がTailへのリンクを持つなら、末尾にTailを追加
	if trie.links.Get(uint32(index)) {
		tail := trie.GetTail(int(index))
		*key = append(*key, tail...)
		return prefixLength + len(tail)
	}
	return prefixLength
}

// GetTail is ...
func (trie *LoudsPrefixTrie) GetTail(node int) []uint16 {
	if trie.links.Get(uint32(node)) {
		tailIndex := trie.links.Rank(uint(node), true) + 1
		tailStart := trie.tailBits.Select(uint32(tailIndex), false)
		tailEnd := trie.tailBits.NextClearBit(tailStart+1) - 1
		tailSize := tailEnd - tailStart
		tailOffset := trie.tailBits.Rank(tailStart, true)
		return trie.tailChars[tailOffset : tailOffset+tailSize]
	}
	return nil
}

// PredictiveSearchBreadthFirst は、指定したノードから葉の方向に全てのノードを幅優先で巡る．
func (trie *LoudsPrefixTrie) PredictiveSearchBreadthFirst(node int, f func(int)) {
	trie.trie.PredictiveSearchBreadthFirst(node, f)
}

// BuildLoudsPrefixTrie は、ソート済みのkeysからトライを作成する
func BuildLoudsPrefixTrie(keys [][]uint16) (*LoudsPrefixTrie, []uint32) {
	// TAIL文字列を抽出
	tailList := ExtractTailU16Strings(keys)
	// Prefixトライを作成
	prefixStringList := make([][]uint16, len(keys))
	for i := 0; i < len(keys); i++ {
		key := keys[i]
		prefixStringList[i] = key[:len(key)-int(tailList[i])]
	}
	prefixTrie, prefixStringNodes := BuildLoudsTrie(prefixStringList)
	// Outsを作成
	outs := NewBitListWithSize(prefixTrie.Size() + 2)
	for i := 0; i < len(prefixStringNodes); i++ {
		outs.Set(int(prefixStringNodes[i]), true)
	}
	// links,tailBits,tailCharsを作成
	linkList := NewBitListWithSize(prefixTrie.Size() + 2)
	tailSize := 0
	for i := 0; i < len(tailList); i++ {
		if tailList[i] > 0 {
			linkList.Set(int(prefixStringNodes[i]), true)
			tailSize++
		}
	}
	links := NewBitVector(linkList.Words, uint32(linkList.Size))
	tailChars := make([]uint16, 0)
	tailBitList := NewBitList()
	a := make([]int, linkList.Size)
	for i := 0; i < len(prefixStringNodes); i++ {
		a[prefixStringNodes[i]] = i
	}
	for i := 0; i < linkList.Size; i++ {
		if linkList.Get(i) {
			keyIndex := a[i]
			tail := keys[keyIndex][len(keys[keyIndex])-int(tailList[keyIndex]):]
			// tail配列に文字を追加
			tailChars = append(tailChars, tail...)
			// tailBitsに文字の区切りと長さを追加
			tailBitList.Add(false)
			for j := 0; j < len(tail); j++ {
				tailBitList.Add(true)
			}
		}
	}
	tailBits := NewBitVector(tailBitList.Words, uint32(tailBitList.Size))

	// インスタンスを生成
	trie := &LoudsPrefixTrie{
		trie:      prefixTrie,
		links:     links,
		outs:      NewBitVector(outs.Words, uint32(outs.Size)),
		tailBits:  tailBits,
		tailChars: tailChars,
	}
	return trie, prefixStringNodes
}

// ExtractTailU16Strings は、文字列の配列から分岐のない末尾(TAIL)を抽出する
func ExtractTailU16Strings(words [][]uint16) []uint32 {
	tails := make([]uint32, len(words))
	for i := 0; i < len(words); i++ {
		prevWord := []uint16{}
		if i != 0 {
			prevWord = words[i-1]
		}
		currentWord := words[i]
		nextWord := []uint16{}
		if i != len(words)-1 {
			nextWord = words[i+1]
		}
		cursor := 0
		for true {
			prevChar := uint16(0)
			currentChar := uint16(0)
			nextChar := uint16(0)
			if cursor < len(prevWord) {
				prevChar = prevWord[cursor]
			}
			if cursor < len(currentWord) {
				currentChar = currentWord[cursor]
			}
			if cursor < len(nextWord) {
				nextChar = nextWord[cursor]
			}
			if prevChar == 0 && currentChar == 0 && nextChar == 0 {
				break
			}
			if prevChar != currentChar && currentChar != nextChar {
				break
			}
			cursor++
		}
		if cursor+1 < len(currentWord) {
			tails[i] = uint32(len(currentWord)) - uint32(cursor) - 1
		} else {
			tails[i] = 0
		}
	}
	return tails
}

// Size is ...
func (trie *LoudsPrefixTrie) Size() int {
	return trie.trie.Size()
}

// IoSize is ...
func (trie *LoudsPrefixTrie) IoSize() int {
	return trie.trie.IoSize() + trie.outs.IoSize() + trie.links.IoSize() + trie.tailBits.IoSize() + IoSizeUint16Array(trie.tailChars)
}
