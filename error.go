package fastapi

type Error struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

func (this *Error) Wrap(msg string) *Error {
	return &Error{
		Code: this.Code,
		Msg:  msg,
	}
}

func (this *Error) Error() string {
	return this.Msg
}

func Throw(err interface{}) {
	panic(err)
}
