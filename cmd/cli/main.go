package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/ByteCrafter-rgb/kv-store/internal/protocol"
)

func main() {
	host := flag.String("host", "localhost", "server host")
	port := flag.Int("port", 6379, "server port")
	flag.Parse()

	addr := fmt.Sprintf("%s:%d", *host, *port)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect to server: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	fmt.Printf("Connected to %s\n", addr)
	fmt.Println("Type commands (PING, ECHO, etc.) - Ctrl+C to exit")

	reader := bufio.NewReader(os.Stdin)
	parser := protocol.NewParser(conn)
	serializer := protocol.NewSerializer(conn)

	for {
		fmt.Print("> ")
		input, err := reader.ReadString('\n')
		if err != nil {
			break
		}

		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		parts := strings.Fields(input)
		args := make([]*protocol.Value, len(parts))
		for i, part := range parts {
			args[i] = protocol.BulkStringValue(part)
		}

		cmd := protocol.ArrayValue(args...)
		if err := serializer.Write(cmd); err != nil {
			fmt.Fprintf(os.Stderr, "Error sending command: %v\n", err)
			continue
		}

		response, err := parser.Parse()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading response: %v\n", err)
			continue
		}

		printResponse(response)
	}
}

func printResponse(v *protocol.Value) {
	switch v.Type {
	case protocol.SimpleString:
		fmt.Printf("(OK) %s\n", v.Str)
	case protocol.Error:
		fmt.Printf("(error) %s\n", v.Str)
	case protocol.Integer:
		fmt.Printf("(integer) %d\n", v.Int)
	case protocol.BulkString:
		if v.Null {
			fmt.Println("(nil)")
		} else {
			fmt.Printf("\"%s\"\n", v.Str)
		}
	case protocol.Array:
		if v.Null {
			fmt.Println("(nil)")
		} else {
			for i, elem := range v.Array {
				fmt.Printf("%d) ", i+1)
				printResponse(elem)
			}
		}
	}
}
