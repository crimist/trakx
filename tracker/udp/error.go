package udp

import (
	"bytes"
	"encoding/binary"
)

type Error struct {
	Action        int32
	TransactionID int32
	ErrorString   []uint8
}

func (e *Error) Marshall() []byte {
	buff := new(bytes.Buffer)
	binary.Write(buff, binary.BigEndian, e.Action)
	binary.Write(buff, binary.BigEndian, e.TransactionID)
	binary.Write(buff, binary.BigEndian, e.ErrorString)
	return buff.Bytes()
}
