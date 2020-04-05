package fastapi

import "strconv"

type Any map[string]interface{}

func (this Any) Set(k string, v interface{}) {
	this[k] = v
}

func (this Any) Get(k string) (v interface{}, exist bool) {
	v, exist = this[k]
	return
}

func ToInt(s string) int64 {
	num, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return int64(num)
}
