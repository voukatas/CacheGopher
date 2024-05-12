package errorutil

import (
	"fmt"
	"runtime"
)

// a package to wrap and keep the context of the errors plus adding a stacktrace

type StackError struct {
	err        error
	stackTrace string
}

func NewStackError(msg error, stackTrace string) *StackError {
	return &StackError{
		err:        msg,
		stackTrace: stackTrace,
	}
}

func (s *StackError) Error() string {
	//return s.err.Error()
	return s.FullError()
}

func (s *StackError) FullError() string {
	//log.Printf("%s\nStackTrace:\n%s", s.err.Error(), s.stackTrace)
	return fmt.Sprintf("%s\nStackTrace:\n%s", s.err.Error(), s.stackTrace)
}

func stackTrace() []byte {

	buf := make([]byte, 2048)
	n := runtime.Stack(buf, false)
	return buf[:n]
}

func Wrap(err error, msg string) error {
	if err == nil {
		return nil
	}
	werr := fmt.Errorf("%s: %w", msg, err)
	stack := stackTrace()
	newError := NewStackError(werr, string(stack))
	return newError
	//return fmt.Errorf("%s: %w\nStack Trace:\n%s", msg, err, stackTrace())
}
