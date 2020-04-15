package fastapi

import "strconv"

func ToInt(s string) int64 {
	num, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return int64(num)
}

func appendSpace(s string, n int) string {
	var m = len(s)
	for i := 0; i < n-m; i++ {
		s += " "
	}
	return s
}
