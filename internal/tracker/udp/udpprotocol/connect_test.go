package udpprotocol

import (
	"bytes"
	"reflect"
	"testing"
)

func TestConnectRequestMarshal(t *testing.T) {
	request := &ConnectRequest{
		ProtocolID:    ProtocolMagic,
		Action:        1,
		TransactionID: 123456789,
	}

	data, err := request.Marshal()
	if err != nil {
		t.Fatalf("Failed to marshal ConnectRequest: %v", err)
	}

	expectedSize := 16
	if len(data) != expectedSize {
		t.Errorf("Expected data length = %d; got %d", expectedSize, len(data))
	}

	expectedData := []byte{0, 0, 4, 23, 39, 16, 25, 128, 0, 0, 0, 1, 7, 91, 205, 21}
	if !bytes.Equal(data, expectedData) {
		t.Errorf("Expected data = %v; got %v", expectedData, data)
	}
}

func TestNewConnectRequest(t *testing.T) {
	expectedRequest := &ConnectRequest{
		ProtocolID:    ProtocolMagic,
		Action:        1,
		TransactionID: 123456789,
	}

	data, err := expectedRequest.Marshal()
	if err != nil {
		t.Fatalf("Failed to marshal ConnectRequest: %v", err)
	}

	decodedRequest, err := NewConnectRequest(data)
	if err != nil {
		t.Fatalf("Failed to decode ConnectRequest: %v", err)
	}

	if !reflect.DeepEqual(expectedRequest, decodedRequest) {
		t.Errorf("Decoded request doesn't match original, original = %+v decoded = %+v", expectedRequest, decodedRequest)
	}
}

func TestConnectResponseMarshal(t *testing.T) {
	response := &ConnectResponse{
		Action:        1,
		TransactionID: 123456789,
		ConnectionID:  987654321,
	}

	data, err := response.Marshal()
	if err != nil {
		t.Fatalf("Failed to marshal ConnectResponse: %v", err)
	}

	expectedSize := 16
	if len(data) != expectedSize {
		t.Errorf("Expected data length = %d; got %d", expectedSize, len(data))
	}

	expectedData := []byte{0, 0, 0, 1, 7, 91, 205, 21, 0, 0, 0, 0, 58, 222, 104, 177}
	if !bytes.Equal(data, expectedData) {
		t.Errorf("Expected data = %v; got %v", expectedData, data)
	}
}

func TestNewConnectResponse(t *testing.T) {
	expectedResponse := &ConnectResponse{
		Action:        1,
		TransactionID: 123456789,
		ConnectionID:  987654321,
	}

	data, err := expectedResponse.Marshal()
	if err != nil {
		t.Fatalf("Failed to marshal ConnectResponse: %v", err)
	}

	decodedResponse, err := NewConnectResponse(data)
	if err != nil {
		t.Fatalf("Failed to decode ConnectResponse: %v", err)
	}

	if !reflect.DeepEqual(expectedResponse, decodedResponse) {
		t.Errorf("Decoded response doesn't match original, original = %+v decoded = %+v", expectedResponse, decodedResponse)
	}
}
