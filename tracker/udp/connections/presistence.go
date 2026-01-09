package connections

import (
	"bufio"
	"bytes"
	"encoding/gob"

	"github.com/pkg/errors"
)

func (connections *Connections) Marshal() ([]byte, error) {
	var buffer bytes.Buffer
	writer := bufio.NewWriter(&buffer)

	connections.mutex.Lock()
	gob.NewEncoder(writer).Encode(connections.associations)
	connections.mutex.Unlock()

	if err := writer.Flush(); err != nil {
		return nil, errors.Wrap(err, "failed to flush writer")
	}

	return buffer.Bytes(), nil
}

func (connections *Connections) Unmarshal(data []byte) error {
	reader := bufio.NewReader(bytes.NewBuffer(data))
	return gob.NewDecoder(reader).Decode(&connections.associations)
}
