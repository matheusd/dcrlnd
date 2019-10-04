package shachain

// getBit returns bit on index at position.
func getBit(index index, position uint8) uint8 {
	return uint8((uint64(index) >> position) & 1)
}

func getPrefix(index index, position uint8) uint64 {
	//	+ -------------------------- +
	// 	| №  | value | mask | return |
	//	+ -- + ----- + ---- + ------ +
	//	| 63 |	 1   |  0   |	 0   |
	//	| 62 |	 0   |  0   |	 0   |
	//	| 61 |   1   |  0   |	 0   |
	//		....
	//	|  4 |	 1   |  0   |	 0   |
	//	|  3 |   1   |  0   |	 0   |
	//	|  2 |   1   |  1   |	 1   | <--- position
	//	|  1 |   0   |  1   |	 0   |
	//	|  0 |   1   |  1   |	 1   |
	//	+ -- + ----- + ---- + ------ +

	var zero uint64
	mask := (zero - 1) - uint64((1<<position)-1)
	return (uint64(index) & mask)
}

// countTrailingZeros counts number of trailing zero bits, this function is
// used to determine the number of element bucket.
func countTrailingZeros(index index) uint8 {
	var zeros uint8
	for ; zeros < maxHeight; zeros++ {
		if getBit(index, zeros) != 0 {
			break
		}
	}

	return zeros
}
