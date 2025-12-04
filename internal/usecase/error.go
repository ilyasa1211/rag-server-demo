package usecase

type ErrorWithStatusCode struct {
	Error error
	Code  int
}

func NewErrorWithStatusCode(err error, code int) *ErrorWithStatusCode {
	return &ErrorWithStatusCode{
		Error: err,
		Code:  code,
	}
}
