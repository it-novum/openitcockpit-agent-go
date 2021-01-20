package safemaths

import "math"

func DivideFloat64(numerator, denominator float64) float64 {
	c := numerator / denominator
	if math.IsNaN(c) || math.IsInf(c, 0) {
		return 0.0
	}
	return c
}

// "Fix" Division by zero
func DivideInt(numerator, denominator int) int {
	if denominator == 0 {
		return 0
	}
	return numerator / denominator
}

func DivideUint64(numerator, denominator uint64) uint64 {
	if denominator == 0 {
		return 0
	}
	return numerator / denominator
}

func DivideInt64(numerator, denominator int64) int64 {
	if denominator == 0 {
		return 0
	}
	return numerator / denominator
}
