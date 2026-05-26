package server

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/ByteCrafter-rgb/kv-store/internal/protocol"
	"github.com/ByteCrafter-rgb/kv-store/internal/store"
)

type Handler struct {
	conn  net.Conn
	store *store.Store
}

func NewHandler(conn net.Conn, s *store.Store) *Handler {
	return &Handler{conn: conn, store: s}
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
	case "GET":
		return h.cmdGet(args)
	case "SET":
		return h.cmdSet(args)
	case "DEL":
		return h.cmdDel(args)
	case "EXISTS":
		return h.cmdExists(args)
	case "EXPIRE":
		return h.cmdExpire(args)
	case "TTL":
		return h.cmdTTL(args)
	case "COMMAND":
		return protocol.ArrayValue()
	default:
		return protocol.ErrorValue(fmt.Sprintf("ERR unknown command '%s'", cmdName))
	}
}

func (h *Handler) cmdGet(args []*Value) *protocol.Value {
	if len(args) != 1 {
		return protocol.ErrorValue("ERR wrong number of arguments for 'get' command")
	}

	key := args[0].Str
	val, exists := h.store.Get(key)
	if !exists {
		return protocol.NilValue()
	}
	return protocol.BulkStringValue(val)
}

func (h *Handler) cmdSet(args []*Value) *protocol.Value {
	if len(args) < 2 {
		return protocol.ErrorValue("ERR wrong number of arguments for 'set' command")
	}

	key := args[0].Str
	value := args[1].Str
	var expireSeconds *int64

	if len(args) >= 4 {
		opt := strings.ToUpper(args[2].Str)
		if opt == "EX" {
			sec, err := strconv.ParseInt(args[3].Str, 10, 64)
			if err != nil {
				return protocol.ErrorValue("ERR value is not an integer or out of range")
			}
			expireSeconds = &sec
		}
	}

	if expireSeconds != nil {
		expireDuration := *store.DurationFromSeconds(*expireSeconds)
		h.store.Set(key, value, &expireDuration)
	} else {
		h.store.Set(key, value, nil)
	}

	return protocol.SimpleStringValue("OK")
}

func (h *Handler) cmdDel(args []*Value) *protocol.Value {
	if len(args) == 0 {
		return protocol.ErrorValue("ERR wrong number of arguments for 'del' command")
	}

	count := int64(0)
	for _, arg := range args {
		if h.store.Delete(arg.Str) {
			count++
		}
	}
	return protocol.IntegerValue(count)
}

func (h *Handler) cmdExists(args []*Value) *protocol.Value {
	if len(args) == 0 {
		return protocol.ErrorValue("ERR wrong number of arguments for 'exists' command")
	}

	count := int64(0)
	for _, arg := range args {
		if h.store.Exists(arg.Str) {
			count++
		}
	}
	return protocol.IntegerValue(count)
}

func (h *Handler) cmdExpire(args []*Value) *protocol.Value {
	if len(args) != 2 {
		return protocol.ErrorValue("ERR wrong number of arguments for 'expire' command")
	}

	key := args[0].Str
	seconds, err := strconv.ParseInt(args[1].Str, 10, 64)
	if err != nil {
		return protocol.ErrorValue("ERR value is not an integer or out of range")
	}

	if h.store.Expire(key, seconds) {
		return protocol.IntegerValue(1)
	}
	return protocol.IntegerValue(0)
}

func (h *Handler) cmdTTL(args []*Value) *protocol.Value {
	if len(args) != 1 {
		return protocol.ErrorValue("ERR wrong number of arguments for 'ttl' command")
	}

	key := args[0].Str
	ttl := h.store.TTL(key)
	return protocol.IntegerValue(ttl)
}

type Value = protocol.Value
