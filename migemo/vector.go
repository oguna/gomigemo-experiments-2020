package migemo

// IoSizeUint64Array is ...
func IoSizeUint64Array(array []uint64) int {
	return len(array)*8 + 4
}

// IoSizeUint32Array is ...
func IoSizeUint32Array(array []uint32) int {
	return len(array)*4 + 4
}

// IoSizeUint16Array is ...
func IoSizeUint16Array(array []uint16) int {
	return len(array)*2 + 4
}

// IoSizeUint8Array is ...
func IoSizeUint8Array(array []uint8) int {
	return len(array) + 4
}
