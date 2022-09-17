package render

import "net/http"

type tagErr struct {
	err  error
	code int
}

func (e tagErr) Error() string {
	return e.err.Error()
}

func (e tagErr) Code() int {
	return e.code
}

// TagBadRequest tags an error with the 400 Bad Request status code.
func TagBadRequest(err error) error {
	return tagErr{
		err:  err,
		code: http.StatusBadRequest,
	}
}

type codeErr struct {
	msg  string
	code int
}

func (e codeErr) Error() string {
	return e.msg
}

func (e codeErr) Code() int {
	return e.code
}

// NewBadRequest creates a new error with the 400 Bad Request status code.
func NewBadRequest(msg string) error {
	return codeErr{
		msg:  msg,
		code: http.StatusBadRequest,
	}
}
