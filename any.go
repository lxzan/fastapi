package fastapi

type Any map[string]interface{}

func (this Any) Set(k string, v interface{}) {
	this[k] = v
}

func (this Any) Get(k string) (v interface{}, exist bool) {
	v, exist = this[k]
	return
}

func (this Any) GetString(key string) (string, bool) {
	v1, ok1 := this[key]
	if !ok1 {
		return "", false
	}
	v2, ok2 := v1.(string)
	if !ok2 {
		return "", false
	}
	return v2, true
}

func (this Any) GetUint8(key string) (uint8, bool) {
	v1, ok1 := this[key]
	if !ok1 {
		return 0, false
	}
	v2, ok2 := v1.(uint8)
	if !ok2 {
		return 0, false
	}
	return v2, true
}

func (this Any) GetInt64(key string) (int64, bool) {
	v1, ok1 := this[key]
	if !ok1 {
		return 0, false
	}
	v2, ok2 := v1.(int64)
	if !ok2 {
		return 0, false
	}
	return v2, true
}

func (this Any) GetFloat64(key string) (float64, bool) {
	v1, ok1 := this[key]
	if !ok1 {
		return 0, false
	}
	v2, ok2 := v1.(float64)
	if !ok2 {
		return 0, false
	}
	return v2, true
}

func (this Any) GetBool(key string) (bool, bool) {
	v1, ok1 := this[key]
	if !ok1 {
		return false, false
	}
	v2, ok2 := v1.(bool)
	if !ok2 {
		return false, false
	}
	return v2, true
}
