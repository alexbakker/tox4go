package util

import (
	"bytes"
	"fmt"
)

func AssertReaderEOF(reader *bytes.Reader) error {
	if reader.Len() != 0 {
		return fmt.Errorf("bad packet: %d unexpected extra bytes", reader.Len())
	}
	return nil
}
