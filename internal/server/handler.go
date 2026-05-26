package server

import (
	"fmt"
	"net"
	"strings"

	"github.com/ByteCrafter-rgb/kv-store/internal/protocol"
)

type Handler struct {
	conn net.Conn
}

func NewHandler(conn net.Conn) *Handler {
	return &Handler{conn: conn}
}

func (h *Handler) Run() {
	parser := protocol.NewParser(h.conn)
	serializer := protocol.NewSerializer(h.conn)

	for {
		val, err := parser.Parse()
		if err != nil {
			fmt.Printf("Parse error: %v\n", err)
			return
		}

		response := h.handleCommand(val)
		if err := serializer.Write(response); err != nil {
			fmt.Printf("Write error: %v\n", err)
			return
		}
	}
}

func (h *Handler) handleCommand(val *Value) *protocol.Value {
	if val.Type != protocol.Array {
		return protocol.ErrorValue("ERR expected array")
	}

	if len(val.Array) == 0 {
		return protocol.ErrorValue("ERR empty command")
	}

	cmd := val.Array[0]
	if cmd.Type != protocol.BulkString {
		return protocol.ErrorValue("ERR invalid command format")
	}

	cmdName := strings.ToUpper(cmd.Str)
	args := val.Array[1:]

	switch cmdName {
	case "PING":
		if len(args) == 0 {
			return protocol.SimpleStringValue("PONG")
		}
		return protocol.BulkStringValue(args[0].Str)
	case "ECHO":
		if len(args) != 1 {
			return protocol.ErrorValue("ERR wrong number of arguments for 'echo' command")
		}
		return protocol.BulkStringValue(args[0].Str)
	case "COMMAND":
		return protocol.ArrayValue()
	default:
		return protocol.ErrorValue(fmt.Sprintf("ERR unknown command '%s'", cmdName))
	}
}

type Value = protocol.Value
