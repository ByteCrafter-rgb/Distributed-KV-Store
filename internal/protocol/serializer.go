package protocol

import (
	"fmt"
	"io"
)

type Serializer struct {
	writer io.Writer
}

func NewSerializer(w io.Writer) *Serializer {
	return &Serializer{writer: w}
}

func (s *Serializer) Write(v *Value) error {
	switch v.Type {
	case SimpleString:
		_, err := fmt.Fprintf(s.writer, "+%s\r\n", v.Str)
		return err
	case Error:
		_, err := fmt.Fprintf(s.writer, "-%s\r\n", v.Str)
		return err
	case Integer:
		_, err := fmt.Fprintf(s.writer, ":%d\r\n", v.Int)
		return err
	case BulkString:
		if v.Null {
			_, err := fmt.Fprint(s.writer, "$-1\r\n")
			return err
		}
		_, err := fmt.Fprintf(s.writer, "$%d\r\n%s\r\n", len(v.Str), v.Str)
		return err
	case Array:
		if v.Null {
			_, err := fmt.Fprint(s.writer, "*-1\r\n")
			return err
		}
		_, err := fmt.Fprintf(s.writer, "*%d\r\n", len(v.Array))
		if err != nil {
			return err
		}
		for _, elem := range v.Array {
			if err := s.Write(elem); err != nil {
				return err
			}
		}
		return nil
	}
	return fmt.Errorf("unknown RESP type: %v", v.Type)
}
