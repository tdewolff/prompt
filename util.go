package prompt

func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func Max(a, b int) int {
	if a < b {
		return b
	}
	return a
}

func Clip(x, a, b int) int {
	if x < a {
		return a
	} else if b < x {
		return b
	}
	return x
}
