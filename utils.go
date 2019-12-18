package main

// Max retruns maximum of a and b.
func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Min returns minimum of a and b.
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Clamp limits value v to [a, b] inclusive range.
func Clamp(v, a, b int) int {
	return Max(a, Min(v, b))
}
