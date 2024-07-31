package v1

import (
	"encoding/json"
)

func (e *Error) Error() string {
	b, _ := json.Marshal(e)
	return string(b)
}

func (e *MultiError) Append(err ...*Error) {
	e.Errors = append(e.Errors, err...)
}

func (e *MultiError) HasErrors() bool {
	return len(e.Errors) > 0
}

func (e *MultiError) Error() string {
	b, _ := json.Marshal(e)
	return string(b)
}
