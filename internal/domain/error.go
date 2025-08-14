package domain

import (
	"fmt"
	"runtime"
)

type Error struct {
	origin error
	msg    string
	code   ErrorCode
	file   string
	line   int
}

type ErrorCode uint

const (
	ErrUnknown ErrorCode = iota
	ErrNotFound
	ErrInvalidArgument
	ErrNetworkError
)

func (e *Error) Code() ErrorCode {
	return e.code
}

func (e *Error) Location() string {
	return fmt.Sprintf("%s:%d", e.file, e.line)
}

func (e *Error) Message() string {
	return e.msg
}

func (e *Error) Error() string {
	location := fmt.Sprintf("%s:%d", e.file, e.line)
	if e.origin != nil {
		return fmt.Sprintf("%s [%s]: %v", e.msg, location, e.origin)
	}

	return fmt.Sprintf("%s [%s]", e.msg, location)
}

func (e *Error) Unwrap() error {
	return e.origin
}

func NewError(code ErrorCode, format string, args ...any) *Error {
	_, file, line, _ := runtime.Caller(1)
	return &Error{
		msg:    fmt.Sprintf(format, args...),
		origin: nil,
		code:   code,
		file:   file,
		line:   line,
	}
}

func WrapError(origin error, code ErrorCode, format string, args ...any) *Error {
	_, file, line, _ := runtime.Caller(1)
	return &Error{
		origin: origin,
		msg:    fmt.Sprintf(format, args...),
		code:   code,
		file:   file,
		line:   line,
	}
}
