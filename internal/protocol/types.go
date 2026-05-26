package protocol

import "fmt"

type RespType int

const (
	SimpleString RespType = iota
	Error
	Integer
	BulkString
	Array
)

type Value struct {
	Type  RespType
	Str   string      // For SimpleString, Error, BulkString
	Int   int64       // For Integer
	Array []*Value    // For Array
	Null  bool        // For nil bulk strings
}

func (v *Value) String() string {
	switch v.Type {
	case SimpleString:
		return fmt.Sprintf("SimpleString(%q)", v.Str)
	case Error:
		return fmt.Sprintf("Error(%q)", v.Str)
	case Integer:
		return fmt.Sprintf("Integer(%d)", v.Int)
	case BulkString:
		if v.Null {
			return "BulkString(nil)"
		}
		return fmt.Sprintf("BulkString(%q)", v.Str)
	case Array:
		return fmt.Sprintf("Array(len=%d)", len(v.Array))
	}
	return "Unknown"
}

func SimpleStringValue(s string) *Value {
	return &Value{Type: SimpleString, Str: s}
}

func ErrorValue(msg string) *Value {
	return &Value{Type: Error, Str: msg}
}

func IntegerValue(n int64) *Value {
	return &Value{Type: Integer, Int: n}
}

func BulkStringValue(s string) *Value {
	return &Value{Type: BulkString, Str: s}
}

func NilValue() *Value {
	return &Value{Type: BulkString, Null: true}
}

func ArrayValue(values ...*Value) *Value {
	return &Value{Type: Array, Array: values}
}
