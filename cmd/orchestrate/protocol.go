package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"time"
)

// MaxMessageSize is the maximum size of a protocol message (1MB).
const MaxMessageSize = 1 << 20

var ErrMessageTooLong = errors.New("message exceeds maximum size")

// Message types for the orchestration protocol.
const (
	MsgTypeStart    = "START"
	MsgTypeReady    = "READY"
	MsgTypeDone     = "DONE"
	MsgTypeStopped  = "STOPPED"
	MsgTypePing     = "PING"
	MsgTypePong     = "PONG"
	MsgTypeError    = "ERROR"
	MsgTypeGetState = "GET_STATE"
	MsgTypeState    = "STATE"
)

// Message is the base protocol message.
type Message struct {
	Type      string `json:"type"`
	Version   int    `json:"version"`
	Timestamp string `json:"timestamp,omitempty"`
}

// StartMessage requests the server to start the tracker.
type StartMessage struct {
	Message
	UDPRoutines  int `json:"udp_routines"`
	HTTPRoutines int `json:"http_routines"`
}

// ReadyMessage indicates the tracker is ready.
type ReadyMessage struct {
	Message
	UDPRoutines  int    `json:"udp_routines"`
	HTTPRoutines int    `json:"http_routines"`
	HTTPAddr     string `json:"http_addr"`
	UDPAddr      string `json:"udp_addr"`
}

// DoneMessage indicates the benchmark is complete.
type DoneMessage struct {
	Message
	UDPRoutines  int    `json:"udp_routines"`
	HTTPRoutines int    `json:"http_routines"`
	ResultPath   string `json:"result_path"`
	Success      bool   `json:"success"`
	ErrorMsg     string `json:"error,omitempty"`
}

// StoppedMessage indicates the tracker has been stopped.
type StoppedMessage struct {
	Message
	UDPRoutines  int `json:"udp_routines"`
	HTTPRoutines int `json:"http_routines"`
}

// ErrorMessage indicates an error occurred.
type ErrorMessage struct {
	Message
	Code        int    `json:"code"`
	ErrorMsg    string `json:"message"`
	Recoverable bool   `json:"recoverable"`
}

// StateMessage returns the current server state.
type StateMessage struct {
	Message
	Status   string `json:"status"`
	UptimeMS int64  `json:"uptime_ms"`
}

// Conn wraps a network connection for protocol messages.
type Conn struct {
	conn   net.Conn
	reader *bufio.Reader
}

// NewConn creates a new protocol connection.
func NewConn(conn net.Conn) *Conn {
	return &Conn{
		conn:   conn,
		reader: bufio.NewReader(conn),
	}
}

// Close closes the connection.
func (c *Conn) Close() error {
	return c.conn.Close()
}

// RemoteAddr returns the remote address.
func (c *Conn) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

// SetDeadline sets the read/write deadline.
func (c *Conn) SetDeadline(t time.Time) error {
	return c.conn.SetDeadline(t)
}

// Send sends any value as a JSON message.
func (c *Conn) Send(msg any) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal message: %w", err)
	}
	data = append(data, '\n')
	_, err = c.conn.Write(data)
	return err
}

// Recv receives and decodes a message into the provided value.
func (c *Conn) Recv(v any) error {
	line, err := c.readLine()
	if err != nil {
		return err
	}
	return json.Unmarshal(line, v)
}

// RecvAny receives a message and returns its type along with the raw JSON.
// Use this when you need to dispatch based on message type.
func (c *Conn) RecvAny() (msgType string, data []byte, err error) {
	line, err := c.readLine()
	if err != nil {
		return "", nil, err
	}

	var base Message
	if err := json.Unmarshal(line, &base); err != nil {
		return "", nil, fmt.Errorf("unmarshal base message: %w", err)
	}

	return base.Type, line, nil
}

// readLine reads a newline-terminated line with size limit.
func (c *Conn) readLine() ([]byte, error) {
	var line []byte
	for {
		b, err := c.reader.ReadByte()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil, io.EOF
			}
			return nil, fmt.Errorf("read message: %w", err)
		}
		if b == '\n' {
			break
		}
		line = append(line, b)
		if len(line) > MaxMessageSize {
			// Drain remaining bytes until newline to resync.
			for {
				b, err := c.reader.ReadByte()
				if err != nil || b == '\n' {
					break
				}
			}
			return nil, fmt.Errorf("%w: exceeded %d bytes", ErrMessageTooLong, MaxMessageSize)
		}
	}
	return line, nil
}

// Helper methods for sending typed messages.

func (c *Conn) SendStart(udpRoutines, httpRoutines int) error {
	return c.Send(StartMessage{
		Message: Message{
			Type:      MsgTypeStart,
			Version:   1,
			Timestamp: time.Now().Format(time.RFC3339),
		},
		UDPRoutines:  udpRoutines,
		HTTPRoutines: httpRoutines,
	})
}

func (c *Conn) SendReady(udpRoutines, httpRoutines int, httpAddr, udpAddr string) error {
	return c.Send(ReadyMessage{
		Message: Message{
			Type:      MsgTypeReady,
			Version:   1,
			Timestamp: time.Now().Format(time.RFC3339),
		},
		UDPRoutines:  udpRoutines,
		HTTPRoutines: httpRoutines,
		HTTPAddr:     httpAddr,
		UDPAddr:      udpAddr,
	})
}

func (c *Conn) SendDone(udpRoutines, httpRoutines int, resultPath string, success bool, errorMsg string) error {
	return c.Send(DoneMessage{
		Message: Message{
			Type:      MsgTypeDone,
			Version:   1,
			Timestamp: time.Now().Format(time.RFC3339),
		},
		UDPRoutines:  udpRoutines,
		HTTPRoutines: httpRoutines,
		ResultPath:   resultPath,
		Success:      success,
		ErrorMsg:     errorMsg,
	})
}

func (c *Conn) SendStopped(udpRoutines, httpRoutines int) error {
	return c.Send(StoppedMessage{
		Message: Message{
			Type:      MsgTypeStopped,
			Version:   1,
			Timestamp: time.Now().Format(time.RFC3339),
		},
		UDPRoutines:  udpRoutines,
		HTTPRoutines: httpRoutines,
	})
}

func (c *Conn) SendError(code int, message string, recoverable bool) error {
	return c.Send(ErrorMessage{
		Message: Message{
			Type:      MsgTypeError,
			Version:   1,
			Timestamp: time.Now().Format(time.RFC3339),
		},
		Code:        code,
		ErrorMsg:    message,
		Recoverable: recoverable,
	})
}

func (c *Conn) SendPing() error {
	return c.Send(Message{
		Type:      MsgTypePing,
		Version:   1,
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

func (c *Conn) SendPong() error {
	return c.Send(Message{
		Type:      MsgTypePong,
		Version:   1,
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

func (c *Conn) SendState(status string, uptimeMS int64) error {
	return c.Send(StateMessage{
		Message: Message{
			Type:      MsgTypeState,
			Version:   1,
			Timestamp: time.Now().Format(time.RFC3339),
		},
		Status:   status,
		UptimeMS: uptimeMS,
	})
}
