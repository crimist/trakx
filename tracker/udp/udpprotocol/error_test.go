package udpprotocol

import (
	"bytes"
	"reflect"
	"testing"
)

func TestErrorResponseMarshal(t *testing.T) {
	errorResponse := &ErrorResponse{
		Action:        3,
		TransactionID: 123456789,
		ErrorString:   []uint8("Error: something went wrong"),
	}

	data, err := errorResponse.Marshal()
	if err != nil {
		t.Fatalf("Failed to marshal ErrorResponse: %v", err)
	}

	expectedSize := 8 + len(errorResponse.ErrorString)
	if len(data) != expectedSize {
		t.Errorf("Expected data length = %d; got %d", expectedSize, len(data))
	}

	expectedData := []byte{0, 0, 0, 3, 7, 91, 205, 21}
	expectedData = append(expectedData, []byte("Error: something went wrong")...)

	if !bytes.Equal(data, expectedData) {
		t.Errorf("Expected data = %v; got %v", expectedData, data)
	}
}

func TestNewErrorResponse(t *testing.T) {
	expectedErrorResponse := &ErrorResponse{
		Action:        3,
		TransactionID: 123456789,
		ErrorString:   []uint8("Error: something went wrong"),
	}

	data, err := expectedErrorResponse.Marshal()
	if err != nil {
		t.Fatalf("Failed to marshal ErrorResponse: %v", err)
	}

	decodedErrorResponse, err := NewErrorResponse(data)
	if err != nil {
		t.Fatalf("Failed to decode ErrorResponse: %v", err)
	}

	if !reflect.DeepEqual(expectedErrorResponse, decodedErrorResponse) {
		t.Errorf("Decoded ErrorResponse doesn't match original, original = %+v decoded = %+v", expectedErrorResponse, decodedErrorResponse)
	}
}
