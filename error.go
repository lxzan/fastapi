package fastapi

type Error struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

func Throw(err interface{}) {
	panic(err)
}
