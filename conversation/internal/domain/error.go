package domain

import (
	"fmt"
)

//go:generate stringer -type=Err -trimprefix=Err
type Err uint8

const (
	_ Err = iota
	ErrNotFound
	ErrAlreadyExists
	ErrInvalidInput
	ErrUnauthorized
)

type Error struct {
	Err Err
	Msg string
}

func (e Err) New(format string, a ...any) *Error {
	return &Error{
		Err: e,
		Msg: fmt.Sprintf(format, a...),
	}
}

func (e Err) Error() *Error {
	return &Error{
		Err: e,
		Msg: e.String(),
	}
}

func (e *Error) Error() string {
	return e.Msg
}
