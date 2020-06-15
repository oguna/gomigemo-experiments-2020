package migemo

// Tail は、末尾文字列を効率的に格納する
type Tail struct {
	bits  *BitVector
	chars []uint16
}

// Get is ...
func (tail *Tail) Get(index int) []uint16 {
	tailStart := tail.bits.Select(uint32(index), false)
	tailEnd := tail.bits.NextClearBit(tailStart+1) - 1
	tailSize := tailEnd - tailStart
	tailOffset := tail.bits.Rank(tailStart, true)
	return tail.chars[tailOffset : tailOffset+tailSize]
}

func (tail *Tail) Match(index int, s []uint16) int {
	return 0
}

func BuildTail(tails [][]uint16) *Tail {
	return nil
}
