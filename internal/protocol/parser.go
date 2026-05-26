package protocol

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type Parser struct {
	reader *bufio.Reader
}

func NewParser(r io.Reader) *Parser {
	return &Parser{reader: bufio.NewReader(r)}
}

func (p *Parser) Parse() (*Value, error) {
	line, err := p.reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("read line: %w", err)
	}

	line = strings.TrimRight(line, "\r\n")
	if len(line) == 0 {
		return nil, fmt.Errorf("empty line")
	}

	switch line[0] {
	case '+':
		return SimpleStringValue(line[1:]), nil
	case '-':
		return ErrorValue(line[1:]), nil
	case ':':
		n, err := strconv.ParseInt(line[1:], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("parse integer: %w", err)
		}
		return IntegerValue(n), nil
	case '$':
		return p.parseBulkString(line)
	case '*':
		return p.parseArray(line)
	default:
		return nil, fmt.Errorf("unknown RESP type: %c", line[0])
	}
}

func (p *Parser) parseBulkString(line string) (*Value, error) {
	length, err := strconv.Atoi(line[1:])
	if err != nil {
		return nil, fmt.Errorf("parse bulk string length: %w", err)
	}

	if length == -1 {
		return NilValue(), nil
	}

	data := make([]byte, length+2) // +2 for \r\n
	_, err = io.ReadFull(p.reader, data)
	if err != nil {
		return nil, fmt.Errorf("read bulk string data: %w", err)
	}

	return BulkStringValue(string(data[:length])), nil
}

func (p *Parser) parseArray(line string) (*Value, error) {
	count, err := strconv.Atoi(line[1:])
	if err != nil {
		return nil, fmt.Errorf("parse array length: %w", err)
	}

	if count == -1 {
		return NilValue(), nil
	}

	arr := make([]*Value, count)
	for i := 0; i < count; i++ {
		val, err := p.Parse()
		if err != nil {
			return nil, fmt.Errorf("parse array element %d: %w", i, err)
		}
		arr[i] = val
	}

	return ArrayValue(arr...), nil
}
