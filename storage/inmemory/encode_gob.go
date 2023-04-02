package inmemory

import (
	"bufio"
	"bytes"
	"encoding/gob"
)

// gob coders scale better than binary coders, the break even point is around 1.5 million peers

func encodeGob(db *InMemory) ([]byte, error) {
	var buff bytes.Buffer
	writer := bufio.NewWriter(&buff)

	db.mutex.RLock()
	if err := gob.NewEncoder(writer).Encode(db.torrents); err != nil {
		db.mutex.RUnlock()
		return nil, err
	}
	db.mutex.RUnlock()

	if err := writer.Flush(); err != nil {
		return nil, err
	}

	return buff.Bytes(), nil
}

func decodeGob(db *InMemory, data []byte) (err error) {
	buff := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(bufio.NewReader(buff))

	err = decoder.Decode(&db.torrents)

	return
}
