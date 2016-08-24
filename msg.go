package scp

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"strconv"
	"strings"
)

const (
	CopyMessage  = 'C'
	ErrorMessage = 0x1
	WarnMessage  = 0x2
)

//Message is scp control message
type Message struct {
	Type     byte
	Error    error
	Mode     string
	Size     int64
	FileName string
}

func (m *Message) readByte(reader io.Reader) (byte, error) {
	buff := make([]byte, 1)
	_, err := io.ReadFull(reader, buff)
	if err != nil {
		return 0, err
	}
	return buff[0], nil

}

func (m *Message) readOpCode(reader io.Reader) error {
	var err error
	m.Type, err = m.readByte(reader)
	return err
}

func (m *Message) ReadError(reader io.Reader) error {
	msg, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}
	m.Error = errors.New(strings.TrimSpace(string(msg)))
	return nil
}

func (m *Message) Readline(reader io.Reader) (string, error) {
	line := ""
	b, err := m.readByte(reader)
	if err != nil {
		return "", err
	}
	for b != 10 {
		line += string(b)
		b, err = m.readByte(reader)
		if err != nil {
			return "", err
		}
	}
	return line, nil
}

func (m *Message) ReadCopy(reader io.Reader) error {
	line, err := m.Readline(reader)
	if err != nil {
		return err
	}
	parts := strings.Split(line, " ")
	if len(parts) < 2 {
		return errors.New("Invalid copy line: " + line)
	}
	m.Mode = parts[0]
	m.Size, err = strconv.ParseInt(parts[1], 10, 0)
	if err != nil {
		return err
	}
	m.FileName = parts[2]
	return nil
}

func (m *Message) ReadFrom(reader io.Reader) error {
	err := m.readOpCode(reader)
	if err != nil {
		return err
	}
	switch m.Type {
	case CopyMessage:
		err = m.ReadCopy(reader)
		if err != nil {
			return err
		}
	case ErrorMessage, WarnMessage:
		err = m.ReadError(reader)
		if err != nil {
			return err
		}
	default:
		return errors.New(fmt.Sprintf("Unsupported opcode: %v", m.Type))
	}
	return nil
}

func NewMessageFromReader(reader io.Reader) (*Message, error) {
	m := new(Message)
	err := m.ReadFrom(reader)
	if err != nil {
		return nil, err
	}
	return m, nil
}
