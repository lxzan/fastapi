package fastapi

type Any map[string]interface{}

func (this Any) Set(k string, v interface{}) {
	this[k] = v
}

func (this Any) Get(k string) (v interface{}, exist bool) {
	v, exist = this[k]
	return
}
