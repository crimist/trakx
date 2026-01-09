package udpprotocol

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/crimist/trakx/storage"
)

func TestAnnounceRequestMarshall(t *testing.T) {
	request := &AnnounceRequest{
		ConnectionID:  123456789,
		Action:        1,
		TransactionID: 987654321,
		InfoHash:      storage.Hash{},
		PeerID:        storage.PeerID{},
		Downloaded:    1000,
		Left:          500,
		Uploaded:      200,
		Event:         0,
		IP:            127001,
		Key:           4321,
		NumWant:       -1,
		Port:          6881,
	}

	data, err := request.Marshal()
	if err != nil {
		t.Fatalf("Failed to marshall AnnounceRequest: %v", err)
	}

	expectedSize := 98
	if len(data) != expectedSize {
		t.Errorf("Expected data length = %d; got %d", expectedSize, len(data))
	}

	expectedData := []byte{0, 0, 0, 0, 7, 91, 205, 21, 0, 0, 0, 1, 58, 222, 104, 177, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 3, 232, 0, 0, 0, 0, 0, 0, 1, 244, 0, 0, 0, 0, 0, 0, 0, 200, 0, 0, 0, 0, 0, 1,
		240, 25, 0, 0, 16, 225, 255, 255, 255, 255, 26, 225}
	if !bytes.Equal(data, expectedData) {
		t.Errorf("Expected data = %v; got %v", expectedData, data)
	}
}

func TestNewAnnounceRequest(t *testing.T) {
	request := &AnnounceRequest{
		ConnectionID:  123456789,
		Action:        1,
		TransactionID: 987654321,
		InfoHash:      storage.Hash{},
		PeerID:        storage.PeerID{},
		Downloaded:    1000,
		Left:          500,
		Uploaded:      200,
		Event:         0,
		IP:            127001,
		Key:           4321,
		NumWant:       -1,
		Port:          6881,
	}

	data, err := request.Marshal()
	if err != nil {
		t.Fatalf("Failed to marshall AnnounceRequest: %v", err)
	}

	decodedRequest, err := NewAnnounceRequest(data)
	if err != nil {
		t.Fatalf("Failed to decode AnnounceRequest: %v", err)
	}

	if !reflect.DeepEqual(request, decodedRequest) {
		t.Errorf("Decoded request doesn't match original, original = %+v decoded = %+v", request, decodedRequest)
	}
}

func TestAnnounceResponseMarshall(t *testing.T) {
	response := &AnnounceResponse{
		Action:        1,
		TransactionID: 987654321,
		Interval:      1800,
		Leeches:       50,
		Seeds:         100,
		Peers:         []byte{192, 168, 1, 1, 0x1A, 0xE1},
	}

	data, err := response.Marshal()
	if err != nil {
		t.Fatalf("Failed to marshall AnnounceResponse: %v", err)
	}

	expectedSize := 26
	if len(data) != expectedSize {
		t.Errorf("Expected data length = %d; got %d", expectedSize, len(data))
	}
	expectedData := []byte{0, 0, 0, 1, 58, 222, 104, 177, 0, 0, 7, 8, 0, 0, 0, 50, 0, 0, 0, 100, 192, 168, 1, 1, 26, 225}
	if !bytes.Equal(data, expectedData) {
		t.Errorf("Expected data = %v; got %v", expectedData, data)
	}
}

func TestNewAnnounceResponse(t *testing.T) {
	response := &AnnounceResponse{
		Action:        1,
		TransactionID: 987654321,
		Interval:      1800,
		Leeches:       50,
		Seeds:         100,
		Peers:         []byte{192, 168, 1, 1, 0x1A, 0xE1},
	}

	data, err := response.Marshal()
	if err != nil {
		t.Fatalf("Failed to marshall AnnounceResponse: %v", err)
	}

	decodedResponse, err := NewAnnounceResponse(data)
	if err != nil {
		t.Fatalf("Failed to decode AnnounceResponse: %v", err)
	}

	if !reflect.DeepEqual(response, decodedResponse) {
		t.Errorf("Decoded response doesn't match original, original = %+v decoded = %+v", response, decodedResponse)
	}
}
