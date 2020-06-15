package migemo

// LoudsPatriciaTrie is ...
type LoudsPatriciaTrie struct {
	trie      *LoudsTrieU16
	outs      *BitVector
	links     *BitVector
	tailBits  *BitVector
	tailChars []uint16
}

// Lookup is ...
func (trie *LoudsPatriciaTrie) Lookup(key []uint16) int {
	node := 1
	cursor := 0
	for cursor < len(key) {
		c := key[cursor]
		node = trie.trie.Traverse(uint32(node), c)
		if node == -1 {
			return -1
		}
		// パトリシアなら、tail配列で一致するか判定
		if trie.links.Get(uint32(node)) {
			tail := trie.GetTail(node)
			cursor++
			for i := 0; i < len(tail); i++ {
				if cursor == len(key) {
					return -node
				}
				if key[cursor] != tail[i] {
					return -1
				}
				cursor++
			}
			if len(key) == cursor {
				return node
			} else if len(key) < cursor {
				return -node
			}
			cursor--
		}
		cursor++
	}
	return node
}

// ReverseLookup is ...
func (trie *LoudsPatriciaTrie) ReverseLookup(node uint32, key *[]uint16) int {
	offset := len(*key)
	for node > 1 {
		// ノードがパトリシアであればtail文字列を追加
		if trie.links.Get(uint32(node)) {
			tail := trie.GetTail(int(node))
			for i := len(tail) - 1; 0 <= i; i-- {
				*key = append(*key, tail[i])
			}
		}
		// ノードのラベル文字を追加
		*key = append(*key, trie.trie.edges[node])
		// 親ノードに移動
		node = trie.trie.Parent(node)
	}
	for i, j := offset, len(*key)-1; i < j; i, j = i+1, j-1 {
		(*key)[i], (*key)[j] = (*key)[j], (*key)[i]
	}
	return len(*key) - offset
}

// GetTail is ...
func (trie *LoudsPatriciaTrie) GetTail(node int) []uint16 {
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

// PredictiveSearchBreadthFirst is ...
func (trie *LoudsPatriciaTrie) PredictiveSearchBreadthFirst(node int, f func(int)) {
	lower := uint(node)
	upper := uint(node + 1)
	for upper-lower > 0 {
		for i := lower; i < upper; i++ {
			f(int(i))
		}
		lower = trie.trie.bitVector.Rank(trie.trie.bitVector.Select(uint32(lower), false)+1, true) + 1
		upper = trie.trie.bitVector.Rank(trie.trie.bitVector.Select(uint32(upper), false)+1, true) + 1
	}
}

// BuildLoudsPatriciaTrie is ...
func BuildLoudsPatriciaTrie(keys [][]uint16) (*LoudsPatriciaTrie, *BitList) {
	trie, oldOuts := BuildLoudsTrie(keys)
	oldOutBits := NewBitListWithSize(trie.Size() + 2)
	for _, e := range oldOuts {
		oldOutBits.Set(int(e), true)
	}
	louds := NewBitList()
	louds.Add(true)
	labels := make([]uint16, 1)
	labels[0] = ' '
	nodes := make([]int, 1)
	nodes[0] = 1
	tailBits := NewBitList()
	links := NewBitList()
	links.Add(false)
	tailChars := make([]uint16, 0)
	outs := NewBitList()
	outs.Add(false)
	level := 0
	for len(nodes) > 0 {
		nextNodes := make([]int, 0)
		for i := 0; i < len(nodes); i++ {
			node := nodes[i]
			pos := trie.bitVector.Select(uint32(node), false)
			louds.Add(false)
			labels = append(labels, trie.edges[node])
			// もし子が一つしかいないなら、子が複数になるか末尾に着くまで、ノードをすすめる
			if level > 0 && trie.bitVector.Get(uint32(pos)+1) == true && trie.bitVector.Get(uint32(pos)+2) == false && !oldOutBits.Get(node) {
				// 子が1つだけなのでパトリシア
				links.Add(true)
				tailBits.Add(false)
				for true {
					// キーでない一人っ子が続く限り、子孫をたどる
					node = int(trie.bitVector.Rank(pos+1, true)) + 1
					pos = trie.bitVector.Select(uint32(node), false)
					tailChars = append(tailChars, trie.edges[node])
					tailBits.Add(true)
					if uint(trie.bitVector.sizeInBits) <= pos+2 || trie.bitVector.Get(uint32(pos)+1) != true || trie.bitVector.Get(uint32(pos)+2) != false || oldOutBits.Get(node) {
						break
					}
				}
			} else {
				links.Add(false)
			}
			outs.Add(oldOutBits.Get(node))
			if pos+1 >= uint(trie.bitVector.sizeInBits) || trie.bitVector.Get(uint32(pos)+1) == false {
				// 子がいないので終了
			} else {
				// 子が複数いるか、子がキーなので、パトリシアにしない
				firstChild := trie.FirstChild(uint32(node))
				for i := 0; trie.bitVector.Get(uint32(i) + uint32(pos) + 1); i++ {
					louds.Add(true)
					nextNodes = append(nextNodes, firstChild+i)
				}
			}
		}
		level++
		nodes = nextNodes
	}
	newLoudsTrie := NewLoudsTrie(NewBitVector(louds.Words, uint32(louds.Size)), labels)
	patriciaTrie := &LoudsPatriciaTrie{
		trie:      newLoudsTrie,
		outs:      NewBitVector(outs.Words, uint32(outs.Size)),
		links:     NewBitVector(links.Words, uint32(links.Size)),
		tailBits:  NewBitVector(tailBits.Words, uint32(tailBits.Size)),
		tailChars: tailChars,
	}
	return patriciaTrie, outs
}

// Size is ...
func (trie *LoudsPatriciaTrie) Size() int {
	return trie.trie.Size()
}

// IoSize is ...
func (trie *LoudsPatriciaTrie) IoSize() int {
	return trie.trie.IoSize() + trie.outs.IoSize() + trie.links.IoSize() + trie.tailBits.IoSize() + IoSizeUint16Array(trie.tailChars)
}
