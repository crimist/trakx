package inmemory

import (
	"bufio"
	"bytes"
	"encoding/gob"
)

func (db *InMemory) encodeGob() ([]byte, error) {
	var buff bytes.Buffer
	w := bufio.NewWriter(&buff)
	encoder := gob.NewEncoder(w)

	db.mutex.RLock()
	if err := encoder.Encode(db.hashes); err != nil {
		return nil, err
	}
	db.mutex.RUnlock()

	if err := w.Flush(); err != nil {
		return nil, err
	}

	return buff.Bytes(), nil
}

func (db *InMemory) decodeGob(data []byte) (err error) {
	db.make()
	buff := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(bufio.NewReader(buff))

	err = decoder.Decode(&db.hashes)

	return
}
