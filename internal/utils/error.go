package utils

type RequestError struct {
	RateLimit bool
	Err       error
}

func (e *RequestError) Error() string {
	return e.Err.Error()
}

func NewRequestError(err error, rateLimit bool) *RequestError {
	return &RequestError{
		RateLimit: rateLimit,
		Err:       err,
	}
}
