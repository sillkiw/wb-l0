package validation

import (
	"errors"
	"fmt"
	"strings"
)

type Code string

const (
	CodeRequired     Code = "required"
	CodeFormat       Code = "format"
	CodeOutOfRange   Code = "out_of_range"
	CodeInconsistent Code = "inconsistent"
)

type FieldError struct {
	Path    string // "order_uid", "items[0].price"
	Code    Code
	Message string
}

func (e FieldError) String() string {
	if e.Message != "" {
		return fmt.Sprintf("%s: %s (%s)", e.Path, e.Message, e.Code)
	}
	return fmt.Sprintf("%s: %s", e.Path, e.Code)
}

type MultiError struct {
	Fields    []FieldError
	Retriable bool
}

func (m *MultiError) Add(path string, code Code, msg string) {
	m.Fields = append(m.Fields, FieldError{Path: path, Code: code, Message: msg})
}
func (m *MultiError) HasErrors() bool { return len(m.Fields) > 0 }

func (m *MultiError) Error() string {
	var b strings.Builder
	b.WriteString("validation failed: ")
	for i, f := range m.Fields {
		if i > 0 {
			b.WriteString("; ")
		}
		b.WriteString(f.String())
	}
	return b.String()
}

var Err = errors.New("validation error")

func Wrap(m *MultiError) error {
	if m == nil || !m.HasErrors() {
		return nil
	}
	return fmt.Errorf("%w: %s", Err, m.Error())
}

func Is(err error) bool { return errors.Is(err, Err) }
