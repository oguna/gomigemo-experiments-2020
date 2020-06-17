package migemo

import (
	"bufio"
	"io"
	"sort"
	"strings"
	"unicode/utf16"
)

// CompactDictionaryU8 は、読み毎に複数の単語を格納した辞書
type CompactDictionaryU8 struct {
	keyTrie          *LoudsTrieU8
	valueTrie        *LoudsTrieU8
	mappingBitVector *BitVector
	mapping          []uint32
	// hasMappingBitList は、あるノードがマッピングを持つかを格納する
	hasMappingBitList *BitList
}

// NewCompactDictionary は、バイト配列からCompactDictionaryを読み込む
/*
func NewCompactDictionary(buffer []uint8) *CompactDictionary {
	var offset = 0
	keyTrie, offset := readTrie(buffer, offset, true)
	valueTrie, offset := readTrie(buffer, offset, false)
	var mappingBitVectorSize = binary.BigEndian.Uint32(buffer[offset:])
	offset = offset + 4
	var mappingBitVectorWords = make([]uint64, (mappingBitVectorSize+63)/64)
	for i := 0; i < len(mappingBitVectorWords); i++ {
		mappingBitVectorWords[i] = binary.BigEndian.Uint64(buffer[offset:])
		offset = offset + 8
	}
	mappingBitVector := NewBitVector(mappingBitVectorWords, mappingBitVectorSize)
	var mappingSize = binary.BigEndian.Uint32(buffer[offset:])
	offset = offset + 4
	mapping := make([]uint32, mappingSize)
	for i := uint32(0); i < mappingSize; i++ {
		mapping[i] = binary.BigEndian.Uint32(buffer[offset:])
		offset += 4
	}
	if offset != len(buffer) {
		return nil
	}
	hasMappingBitList := createHasMappingBitList(mappingBitVector)
	return &CompactDictionary{
		keyTrie,
		valueTrie,
		mappingBitVector,
		mapping,
		hasMappingBitList,
	}
}
*/

// Search は、キーに一致する単語をコールバック関数fに返す
func (compactDictionary *CompactDictionaryU8) Search(key []uint8, f func([]uint8)) {
	var keyIndex = compactDictionary.keyTrie.Lookup(key)
	if keyIndex != -1 {
		var valueStartPos = compactDictionary.mappingBitVector.Select(uint32(keyIndex), false)
		var valueEndPos = compactDictionary.mappingBitVector.NextClearBit(valueStartPos + 1)
		var size = uint(valueEndPos - valueStartPos - 1)
		if size > 0 {
			var offset = compactDictionary.mappingBitVector.Rank(valueStartPos, false)
			word := make([]uint8, 0, 16)
			for i := uint(0); i < size; i++ {
				compactDictionary.valueTrie.ReverseLookup(compactDictionary.mapping[valueStartPos-offset+i], &word)
				f(word)
				word = word[:0]
			}
		}
	}
}

// PredictiveSearch は、接頭辞がkeyに一致する全ての単語をコールバック関数fに返す
func (compactDictionary *CompactDictionaryU8) PredictiveSearch(key []uint8, f func([]uint8)) {
	var keyIndex = compactDictionary.keyTrie.Lookup(key)
	word := make([]uint8, 0, 16)
	if keyIndex > 1 {
		compactDictionary.keyTrie.PredictiveSearchBreadthFirst(keyIndex, func(i int) {
			if compactDictionary.hasMappingBitList.Get(i) {
				var valueStartPos uint = compactDictionary.mappingBitVector.Select(uint32(i), false)
				var valueEndPos uint = compactDictionary.mappingBitVector.NextClearBit(valueStartPos + 1)
				var size = valueEndPos - valueStartPos - 1
				var offset = compactDictionary.mappingBitVector.Rank(valueStartPos, false)
				for j := uint(0); j < size; j++ {
					compactDictionary.valueTrie.ReverseLookup(compactDictionary.mapping[valueStartPos-offset+j], &word)
					f(word)
					word = word[:0]
				}
			}
		})
	}
}

// Save は、ファイルに辞書の内容を書き込む
/*
func (compactDictionary *CompactDictionary) Save(fp *os.File) {
	writer := bufio.NewWriter(fp)
	buffer := make([]byte, 8)
	// output key trie
	keyTriEdgesLength := uint32((len(compactDictionary.keyTrie.edges)))
	binary.BigEndian.PutUint32(buffer, keyTriEdgesLength)
	writer.Write(buffer[0:4])
	for i := 0; i < len(compactDictionary.keyTrie.edges); i++ {
		buffer[0] = encode(compactDictionary.keyTrie.edges[i])
		writer.Write(buffer[:1])
	}
	binary.BigEndian.PutUint32(buffer, uint32(compactDictionary.keyTrie.bitVector.Size()))
	writer.Write(buffer[0:4])
	keyTrieBitVectorWords := compactDictionary.keyTrie.bitVector.words
	for i := 0; i < len(keyTrieBitVectorWords); i++ {
		binary.BigEndian.PutUint64(buffer, keyTrieBitVectorWords[i])
		writer.Write(buffer)
	}
	// output value trie
	valueTriEdgesLength := uint32((len(compactDictionary.valueTrie.edges)))
	binary.BigEndian.PutUint32(buffer, valueTriEdgesLength)
	writer.Write(buffer[:4])
	for i := 0; i < len(compactDictionary.valueTrie.edges); i++ {
		binary.BigEndian.PutUint16(buffer, compactDictionary.valueTrie.edges[i])
		writer.Write(buffer[:2])
	}
	binary.BigEndian.PutUint32(buffer, uint32(compactDictionary.valueTrie.bitVector.Size()))
	writer.Write(buffer[0:4])
	valueTrieBitVectorWords := compactDictionary.valueTrie.bitVector.words
	for i := 0; i < len(valueTrieBitVectorWords); i++ {
		binary.BigEndian.PutUint64(buffer, valueTrieBitVectorWords[i])
		writer.Write(buffer)
	}

	// output mapping trie
	binary.BigEndian.PutUint32(buffer, uint32(compactDictionary.mappingBitVector.sizeInBits))
	writer.Write(buffer[:4])
	for _, w := range compactDictionary.mappingBitVector.words {
		binary.BigEndian.PutUint64(buffer, w)
		writer.Write(buffer)
	}
	binary.BigEndian.PutUint32(buffer, uint32(len(compactDictionary.mapping)))
	writer.Write(buffer[:4])
	for _, x := range compactDictionary.mapping {
		binary.BigEndian.PutUint32(buffer, x)
		writer.Write(buffer[:4])
	}
	writer.Flush()
}
*/

// IoSize is ...
func (compactDictionary *CompactDictionaryU8) IoSize() int {
	return compactDictionary.keyTrie.IoSize() + compactDictionary.valueTrie.IoSize() + compactDictionary.mappingBitVector.IoSize() + IoSizeUint32Array(compactDictionary.mapping)
}

// NodeSize is ...
func (compactDictionary *CompactDictionaryU8) NodeSize() (int, int) {
	return compactDictionary.keyTrie.Size(), compactDictionary.valueTrie.Size()
}

// BuildDictionaryU8FromMigemoDictFile は、ファイルからCompactDictionaryを読み込む
func BuildDictionaryU8FromMigemoDictFile(fp io.Reader) *CompactDictionaryU8 {
	scanner := bufio.NewScanner(fp)
	dict := make(map[string][]string)
	keys := make([]string, 0, 1024)
	values := make(map[string]struct{}, 1024)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, ";") || len(line) == 0 {
			continue
		}
		columns := strings.Split(line, "\t")
		key := columns[0]
		var skip = false
		for _, c := range utf16.Encode([]rune(key)) {
			if encode(c) == 0 {
				println("skip this word: ", key)
				skip = true
				break
			}
		}
		if skip {
			continue
		}
		keys = append(keys, key)
		for _, w := range columns[1:] {
			values[w] = struct{}{}
		}
		dict[key] = columns[1:]
	}

	// build key trie
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	keyTrieBuilder := NewLoudsTrieBuilderU8()
	for _, key := range keys {
		keyTrieBuilder.Add(key)
	}
	keyTrie := keyTrieBuilder.Build()
	//keyTrie, _ := BuildLoudsTrieU8(keys)

	// build value trie
	valuesUtf8 := make([]string, 0, len(values))
	for k := range values {
		valuesUtf8 = append(valuesUtf8, k)
	}
	sort.Slice(valuesUtf8, func(i, j int) bool { return valuesUtf8[i] < valuesUtf8[j] })
	valueTrieBuilder := NewLoudsTrieBuilderU8()
	for _, key := range valuesUtf8 {
		valueTrieBuilder.Add(key)
	}
	valueTrie := valueTrieBuilder.Build()
	//valueTrie, _ := BuildLoudsTrieU8(valuesUtf8)

	// build mapping from key trie to value trie
	mappingCount := 0
	for _, v := range dict {
		mappingCount += len(v)
	}
	mapping := make([]uint32, mappingCount)
	mappingIndex := 0
	mappingBitList := NewBitList()
	key := make([]uint8, 0, 16)
	for i := 1; i <= keyTrie.Size(); i++ {
		key = key[:0]
		keyTrie.ReverseLookup(uint32(i), &key)
		mappingBitList.Add(false)
		words, ok := dict[string(key)]
		if ok {
			for j := 0; j < len(words); j++ {
				mappingBitList.Add(true)
				a := valueTrie.Lookup([]byte(words[j]))
				b := sort.SearchStrings(valuesUtf8, words[j])
				if a <= 0 {
					panic("")
				}
				if valuesUtf8[b] != words[j] {
					panic("")
				}
				mapping[mappingIndex] = uint32(a)
				mappingIndex++
			}
		}
	}
	mappingBitVector := NewBitVector(mappingBitList.Words, uint32(mappingBitList.Size))

	return &CompactDictionaryU8{
		keyTrie:           keyTrie,
		valueTrie:         valueTrie,
		mapping:           mapping,
		mappingBitVector:  mappingBitVector,
		hasMappingBitList: createHasMappingBitList(mappingBitVector),
	}
}
