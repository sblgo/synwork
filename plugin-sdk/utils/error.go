package utils

import "fmt"

type composedError struct {
	origin  error
	message string
}

func (c *composedError) Error() string {
	return fmt.Sprintf("%s [caused by>] %s", c.message, c.origin.Error())
}

func CompError(err error, format string, v ...interface{}) error {
	return &composedError{err, fmt.Sprintf(format, v...)}
}
