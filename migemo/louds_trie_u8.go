package migemo

// LoudsTrieU8 は、LOUDS(level order unary degree sequence)を実装したもの
type LoudsTrieU8 struct {
	bitVector *BitVector
	edges     []uint8
}

// NewLoudsTrieU8 は、LoudsTrieを初期化する
func NewLoudsTrieU8(bitVector *BitVector, edges []uint8) *LoudsTrieU8 {
	return &LoudsTrieU8{
		bitVector,
		edges,
	}
}

// ReverseLookup は、ノード番号indexからkeyを復元する
func (trie *LoudsTrieU8) ReverseLookup(index uint32, key *[]uint8) int {
	offset := len(*key)
	for index > 1 {
		*key = append(*key, trie.edges[index])
		index = trie.Parent(index)
	}
	for i, j := offset, len(*key)-1; i < j; i, j = i+1, j-1 {
		(*key)[i], (*key)[j] = (*key)[j], (*key)[i]
	}
	return len(*key) - offset
}

// Parent は、ノード番号indexの親を返す
func (trie *LoudsTrieU8) Parent(x uint32) uint32 {
	return uint32(trie.bitVector.Rank(trie.bitVector.Select(x, true), false))
}

// FirstChild は、ノード番号xのはじめの子供のノード番号を返す。子供がなければ-1
func (trie *LoudsTrieU8) FirstChild(x uint32) int {
	y := trie.bitVector.Select(x, false) + 1
	if trie.bitVector.Get(uint32(y)) {
		return int(trie.bitVector.Rank(y, true)) + 1
	}
	return -1
}

// Traverse は、ノード番号indexの子ノードのうち、ラベルcを持つノード番号を返す。見つからなければ-1
func (trie *LoudsTrieU8) Traverse(index uint32, c uint8) int {
	firstChild := trie.FirstChild(index)
	if firstChild == -1 {
		return -1
	}
	var childStartBit = trie.bitVector.Select(uint32(firstChild), true)
	var childEndBit = trie.bitVector.NextClearBit(childStartBit)
	var childSize = childEndBit - childStartBit
	var result = binarySearchUint8(trie.edges, uint32(firstChild), uint32(firstChild)+uint32(childSize), c)
	if result >= 0 {
		return result
	}
	return -1
}

// Lookup は、検索対象keyのノード番号を返す。見つからければ-1
func (trie *LoudsTrieU8) Lookup(key []uint8) int {
	var nodeIndex int = 1
	for _, c := range key {
		nodeIndex = trie.Traverse(uint32(nodeIndex), c)
		if nodeIndex == -1 {
			break
		}
	}
	if nodeIndex >= 0 {
		return nodeIndex
	}
	return -1
}

// PredictiveSearchDepthFirst は、指定したノードから葉の方向に全てのノードを深さ優先で巡る
func (trie *LoudsTrieU8) PredictiveSearchDepthFirst(index int, f func(int, []uint8)) {
	key := make([]uint8, 0, 8)
	f(index, key)
	childPos := trie.bitVector.Select(uint32(index), false) + 1
	if trie.bitVector.Get(uint32(childPos)) {
		child := int(trie.bitVector.Rank(childPos, true)) + 1
		for trie.bitVector.Get(uint32(childPos)) {
			key = append(key, trie.edges[child])
			trie.predictiveSearchDepthFirstInternal(child, &key, f)
			key = (key)[:len(key)-1]
			child++
			childPos++
		}
	}
}

func (trie *LoudsTrieU8) predictiveSearchDepthFirstInternal(index int, key *[]uint8, f func(int, []uint8)) {
	f(index, *key)
	childPos := trie.bitVector.Select(uint32(index), false) + 1
	if trie.bitVector.Get(uint32(childPos)) {
		child := int(trie.bitVector.Rank(childPos, true)) + 1
		for trie.bitVector.Get(uint32(childPos)) {
			*key = append(*key, trie.edges[child])
			trie.predictiveSearchDepthFirstInternal(child, key, f)
			*key = (*key)[:len(*key)-1]
			child++
			childPos++
		}
	}
}

// PredictiveSearchBreadthFirst は、指定したノードから葉の方向に全てのノードを幅優先で巡る．
func (trie *LoudsTrieU8) PredictiveSearchBreadthFirst(node int, f func(int)) {
	lower := uint(node)
	upper := uint(node + 1)
	for upper-lower > 0 {
		for i := lower; i < upper; i++ {
			f(int(i))
		}
		lower = trie.bitVector.Rank(trie.bitVector.Select(uint32(lower), false)+1, true) + 1
		upper = trie.bitVector.Rank(trie.bitVector.Select(uint32(upper), false)+1, true) + 1
	}
}

// Size は、ノードの個数を返す
func (trie *LoudsTrieU8) Size() int {
	return len(trie.edges) - 2
}

func binarySearchUint8(a []uint8, fromIndex uint32, toIndex uint32, key uint8) int {
	var low = fromIndex
	var high = toIndex - 1
	for low <= high {
		var mid = (low + high) >> 1
		var midVal = a[mid]
		if midVal < key {
			low = mid + 1
		} else if midVal > key {
			high = mid - 1
		} else {
			return int(mid)
		}
	}
	return -int(low + 1)
}

// BuildLoudsTrieU8 は、UTF16文字列の配列からLoudsTrieを生成する
func BuildLoudsTrieU8(keys []string) (*LoudsTrieU8, []uint32) {
	for i := 0; i < len(keys)-1; i++ {
		if keys[i] >= keys[i+1] {
			panic("invalid key order")
		}
	}
	var nodes = make([]uint32, len(keys))
	for i := 0; i < len(nodes); i++ {
		nodes[i] = 1
	}
	var cursor = 0
	var currentNode uint32 = 1
	var edges = []uint8{0x20, 0x20}
	var louds = NewBitList()
	louds.Add(true)
	for true {
		var lastChar uint8 = 0
		var lastParent uint32 = 0
		var restKeys uint32 = 0
		for i := 0; i < len(keys); i++ {
			if len(keys[i]) < cursor {
				continue
			}
			if len(keys[i]) == cursor {
				louds.Add(false)
				lastParent = nodes[i]
				lastChar = 0
				continue
			}
			var currentChar = keys[i][cursor]
			var currentParent = nodes[i]
			if lastParent != currentParent {
				louds.Add(false)
				louds.Add(true)
				edges = append(edges, currentChar)
				currentNode = currentNode + 1
			} else if lastChar != currentChar {
				louds.Add(true)
				edges = append(edges, currentChar)
				currentNode = currentNode + 1
			}
			nodes[i] = currentNode
			lastChar = currentChar
			lastParent = currentParent
			restKeys++
		}
		if restKeys == 0 {
			break
		}
		cursor++
	}
	var bitVector = NewBitVector(louds.Words, uint32(louds.Size))
	return NewLoudsTrieU8(bitVector, edges), nodes
}
