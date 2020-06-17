package migemo

import "errors"

// LevelU8 は、Loudsツリーの深さ毎のloudsやlabelを格納した構造体
type LevelU8 struct {
	louds  []bool
	outs   []bool
	labels []byte
}

// LoudsTrieBuilderU8 は、LoudsTrieを生成するための構造体
type LoudsTrieBuilderU8 struct {
	levels  []LevelU8
	lastKey []uint8
}

// NewLoudsTrieBuilderU8 は、LoudsTrieBuilderを初期化する
func NewLoudsTrieBuilderU8() *LoudsTrieBuilderU8 {
	level0 := LevelU8{
		louds:  []bool{true, false},
		outs:   []bool{false},
		labels: []uint8{' ', ' '},
	}
	level1 := LevelU8{
		louds: []bool{false},
	}
	levels := []LevelU8{level0, level1}
	return &LoudsTrieBuilderU8{
		levels:  levels,
		lastKey: []uint8{},
	}
}

// Add は、LoudsTrieBuliderにキーを追加する(追加するキーは辞書順)
func (builder *LoudsTrieBuilderU8) Add(key string) error {
	if string(key) <= string(builder.lastKey) {
		return errors.New("key must be larger than last added key")
	}
	if len(key) == 0 {
		builder.levels[0].outs[0] = true
		return nil
	}
	if len(key)+1 >= len(builder.levels) {
		builder.levels = append(builder.levels, make([]LevelU8, len(key)+2-len(builder.levels))...)
	}
	i := 0
	for ; i < len(key); i++ {
		var level = &builder.levels[i+1]
		if (i == len(builder.lastKey)) || key[i] != level.labels[len(level.labels)-1] {
			level.louds[len(builder.levels[i+1].louds)-1] = true
			level.louds = append(level.louds, false)
			level.outs = append(level.outs, false)
			level.labels = append(level.labels, key[i])
			break
		}
	}
	for i++; i < len(key); i++ {
		var level = &builder.levels[i+1]
		level.louds = append(level.louds, true, false)
		level.outs = append(level.outs, false)
		level.labels = append(level.labels, key[i])
	}
	builder.levels[len(key)+1].louds = append(builder.levels[len(key)+1].louds, true)
	builder.levels[len(key)].outs[len(builder.levels[len(key)].outs)-1] = true
	builder.lastKey = make([]uint8, len(key))
	copy(builder.lastKey, key)
	return nil
}

// Build は、LoudsTrieBuilderに追加した文字列からLoudsTrieを生成する
func (builder *LoudsTrieBuilderU8) Build() *LoudsTrieU8 {
	louds := []bool{}
	outs := []bool{}
	labels := []uint8{}
	for _, level := range builder.levels {
		louds = append(louds, level.louds...)
		outs = append(outs, level.outs...)
		labels = append(labels, level.labels...)
	}
	louds = louds[:len(louds)-1]
	words := make([]uint64, (len(louds)+63)/64)
	for i := 0; i < len(louds); i++ {
		if louds[i] {
			words[i/64] |= 1 << (i % 64)
		}
	}
	var bitVector = NewBitVector(words, uint32(len(louds)))
	return NewLoudsTrieU8(bitVector, labels)
}
