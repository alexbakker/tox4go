package state

import "fmt"

// GlobalCookieError represents an error that occurs during a magic number check.
// It provides the value it expected and the value that it actually found.
type GlobalCookieError struct {
	expected uint32
	actual   uint32
}

// InnerCookieError represents an error that occurs during a magic number check.
// It provides the value it expected and the value that it actually found.
type InnerCookieError struct {
	expected uint16
	actual   uint16
}

func (e GlobalCookieError) Error() string {
	return fmt.Sprintf("incorrect global cookie: 0x%x (should be 0x%x)",
		e.actual, e.expected)
}

func (e InnerCookieError) Error() string {
	return fmt.Sprintf("incorrect inner cookie: 0x%x (should be 0x%x)",
		e.actual, e.expected)
}
