package startstopper

import (
	"slices"
)

type errorCode struct {
	error
	code string
}

// NewErrorCode ...
func NewErrorCode(err error, code string) *errorCode {
	return &errorCode{
		error: err,
		code:  code,
	}
}

// ErrorCode ...
func (e *errorCode) ErrorCode() string {
	return e.code
}

// Unwrap ...
func (e *errorCode) Unwrap() error {
	return e.error
}

// Is ...
func (e *errorCode) Is(target error) bool {
	if target == e.error {
		return true
	}

	return MatchErrorCodes(target, e.code)
}

// As ...
func (e *errorCode) As(target any) bool {
	switch t := target.(type) {
	case interface{ MatchErrorCodes(error, ...string) bool }:
		return t.MatchErrorCodes(e.error, e.code)

	default:
		return false
	}

	return true
}

// ErrorCodeMatcher ...
type ErrorCodeMatcher struct {
	Error error
	Codes []string
}

// NewErrorCodeMatcher ...
func NewErrorCodeMatcher(codes []string) *ErrorCodeMatcher {
	return &ErrorCodeMatcher{
		Codes: codes,
	}
}

func (e *ErrorCodeMatcher) MatchErrorCodes(target error, errCodes ...string) bool {
	return slices.ContainsFunc(e.Codes, func(errCode string) bool {
		if slices.Contains(errCodes, errCode) {
			e.Error = target
			return true
		}
		return false
	})
}

// MatchErrorCodes ...
func MatchErrorCodes(target any, errCodes ...string) bool {
	if t, ok := target.(interface{ ErrorCode() string }); ok && slices.Contains(errCodes, t.ErrorCode()) {
		return true
	}

	if t, ok := target.(interface{ ErrorCodes() []string }); ok {
		slices.ContainsFunc(t.ErrorCodes(), func(errCode string) bool {
			return slices.Contains(errCodes, errCode)
		})
	}

	return false
}
