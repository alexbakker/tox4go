package main

import "fmt"

type errorKeyLength struct {
	kind     string
	expected int
	actual   int
}

func (e errorKeyLength) Error() string {
	return fmt.Sprintf("invalid %s key length, expected: %d, actual: %d",
		e.kind, e.expected, e.actual)
}
