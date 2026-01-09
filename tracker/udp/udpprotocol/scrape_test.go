package udpprotocol

import (
	"bytes"
	"encoding/binary"
	"reflect"
	"testing"

	"github.com/crimist/trakx/storage"
)

func TestNewScrapeRequest(t *testing.T) {
	expectedScrapeRequest := &ScrapeRequest{
		ConnectionID:  123456789,
		Action:        2,
		TransactionID: 987654321,
		InfoHashes:    []storage.Hash{{0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xAA, 0xBB, 0xCC, 0xDD, 0xEE, 0xFF, 0x00, 0x11, 0x22, 0x33, 0x44}},
	}

	data := make([]byte, 16+len(expectedScrapeRequest.InfoHashes)*20)
	binary.BigEndian.PutUint64(data[0:8], uint64(expectedScrapeRequest.ConnectionID))
	binary.BigEndian.PutUint32(data[8:12], uint32(expectedScrapeRequest.Action))
	binary.BigEndian.PutUint32(data[12:16], uint32(expectedScrapeRequest.TransactionID))
	copy(data[16:], expectedScrapeRequest.InfoHashes[0][:])

	decodedScrapeRequest, err := NewScrapeRequest(data)
	if err != nil {
		t.Fatalf("Failed to decode ScrapeRequest: %v", err)
	}

	if !reflect.DeepEqual(expectedScrapeRequest, decodedScrapeRequest) {
		t.Errorf("Decoded ScrapeRequest doesn't match original, original = %+v decoded = %+v", expectedScrapeRequest, decodedScrapeRequest)
	}
}

func TestScrapeResponseMarshal(t *testing.T) {
	response := &ScrapeResponse{
		Action:        2,
		TransactionID: 987654321,
		Info: []ScrapeResponseInfo{
			{Complete: 5, Incomplete: 3, Downloaded: 10},
			{Complete: 7, Incomplete: 4, Downloaded: 15},
		},
	}

	data, err := response.Marshal()
	if err != nil {
		t.Fatalf("Failed to marshal ScrapeResponse: %v", err)
	}

	expectedData := make([]byte, 8+len(response.Info)*12)
	binary.BigEndian.PutUint32(expectedData[0:4], uint32(response.Action))
	binary.BigEndian.PutUint32(expectedData[4:8], uint32(response.TransactionID))

	offset := 8
	for _, info := range response.Info {
		binary.BigEndian.PutUint32(expectedData[offset:offset+4], uint32(info.Complete))
		binary.BigEndian.PutUint32(expectedData[offset+4:offset+8], uint32(info.Incomplete))
		binary.BigEndian.PutUint32(expectedData[offset+8:offset+12], uint32(info.Downloaded))
		offset += 12
	}

	if !bytes.Equal(data, expectedData) {
		t.Errorf("Expected data = %v; got %v", expectedData, data)
	}
}
