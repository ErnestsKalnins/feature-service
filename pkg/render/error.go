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
