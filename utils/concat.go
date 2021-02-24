package utils

func ConcatStringSlice(slices ...[]string) []string {
	result := make([]string, 0)
	for _, slice := range slices {
		result = append(result, slice...)
	}
	return result
}
