

// Package exit provides stubs so that unit tests can exercise code that calls
// os.Exit(1).
package exit

import "os"

var real = func() { os.Exit(1) }

// Exit normally terminates the process by calling os.Exit(1). If the package
// is stubbed, it instead records a call in the testing spy.
func Exit() {
	real()
}

// A StubbedExit is a testing fake for os.Exit.
type StubbedExit struct {
	Exited bool
	prev   func()
}

// Stub substitutes a fake for the call to os.Exit(1).
func Stub() *StubbedExit {
	s := &StubbedExit{prev: real}
	real = s.exit
	return s
}

// WithStub runs the supplied function with Exit stubbed. It returns the stub
// used, so that users can test whether the process would have crashed.
func WithStub(f func()) *StubbedExit {
	s := Stub()
	defer s.Unstub()
	f()
	return s
}

// Unstub restores the previous exit function.
func (se *StubbedExit) Unstub() {
	real = se.prev
}

func (se *StubbedExit) exit() {
	se.Exited = true
}
