package fastapi

type TransError struct {
	Message string
	Field   string
}

func (this *TransError) Error() string {
	return this.Message
}

func NewError(code Code, msg string) *Error {
	return &Error{Code: code, Msg: msg}
}

type Error struct {
	Code Code   `json:"code"`
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

func Throw(err error) {
	panic(err)
}
