package safemaths

import "math"

func DivideFloat64(a, b float64) float64 {
	c := a / b
	if math.IsNaN(c) || math.IsInf(c, 0) {
		return 0.0
	}
	return c
}

func DivideInt(a, b int) int {
	if b == 0 {
		return 0
	}
	return a / b
}

func DivideUint64(a, b uint64) uint64 {
	if b == 0 {
		return 0
	}
	return a / b
}
