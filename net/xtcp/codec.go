package xtcp

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"

	"github.com/xsuners/mo/log"
)

const (
	_messageLenBytes = 4
	_messageMaxBytes = 1 << 23 // 8M
)

// Codecc .
type Codecc struct{}

var _ Codec = Codecc{}

// Decode decodes the bytes data into Message
// func Decode(raw io.Reader) (*Message, error) {
func (Codecc) Decode(raw io.Reader) ([]byte, error) {
	ch := make(chan []byte)
	ech := make(chan error)

	go func(bc chan []byte, ec chan error) {
		lengthBytes := make([]byte, _messageLenBytes)
		_, err := io.ReadFull(raw, lengthBytes)
		if err != nil {
			ec <- err
			close(bc)
			close(ec)
			return
		}
		bc <- lengthBytes
	}(ch, ech)

	select {
	case err := <-ech:
		return nil, err
	case lengthBytes := <-ch:
		lengthBuf := bytes.NewReader(lengthBytes)
		var msgLen uint32
		err := binary.Read(lengthBuf, binary.LittleEndian, &msgLen)
		if err != nil {
			return nil, err
		}
		if msgLen > _messageMaxBytes {
			log.Errorf("message has bytes(%d) beyond max %d", msgLen, _messageMaxBytes)
			return nil, errors.New("codec: message length over max bytes")
		}

		msgBytes := make([]byte, msgLen)
		if _, err = io.ReadFull(raw, msgBytes); err != nil {
			return nil, err
		}
		return msgBytes, nil

		// message := &Message{}
		// err = proto.Unmarshal(msgBytes, message)
		// if err != nil {
		// 	return nil, err
		// }
		// return message, nil
	}
}

// Encode encodes the message into bytes data.
func (Codecc) Encode(data []byte) ([]byte, error) {
	// data, err := proto.Marshal(message)
	// if err != nil {
	// 	return nil, err
	// }
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, int32(len(data)))
	buf.Write(data)
	data = buf.Bytes()
	return data, nil
}
