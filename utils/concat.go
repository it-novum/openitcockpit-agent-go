package utils

func ConcatStringSlice(slices ...[]string) []string {
	size := 0
	for _, slice := range slices {
		size += len(slice)
	}
	result := make([]string, size)

	offset := 0
	for _, slice := range slices {
		for i := 0; i < len(slice); i++ {
			result[offset+i] = slice[i]
		}
		offset += len(slice)
	}

	return result
}
