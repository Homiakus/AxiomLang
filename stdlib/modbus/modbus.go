package modbus

import (
	"encoding/binary"
	"errors"
	"fmt"
	"sync"
	"time"

	"AxiomLang/stdlib/uart"
)

// Function Codes
const (
	FuncReadCoils              uint8 = 0x01
	FuncReadDiscreteInputs     uint8 = 0x02
	FuncReadHoldingRegisters   uint8 = 0x03
	FuncReadInputRegisters     uint8 = 0x04
	FuncWriteSingleCoil        uint8 = 0x05
	FuncWriteSingleRegister    uint8 = 0x06
	FuncWriteMultipleCoils     uint8 = 0x0F
	FuncWriteMultipleRegisters uint8 = 0x10
)

// CRC16 calculates standard Modbus RTU polynomial 0xA001 checksum
func CRC16(data []byte) uint16 {
	var crc uint16 = 0xFFFF
	for _, b := range data {
		crc ^= uint16(b)
		for i := 0; i < 8; i++ {
			if (crc & 0x0001) != 0 {
				crc = (crc >> 1) ^ 0xA001
			} else {
				crc >>= 1
			}
		}
	}
	return crc
}

// Client represents a Modbus RTU master controller communicating over UART
type Client struct {
	port    *uart.Port
	timeout time.Duration
	mu      sync.Mutex
}

// NewClient creates a new Modbus RTU client binding to a UART port
func NewClient(p *uart.Port) *Client {
	return &Client{
		port:    p,
		timeout: 500 * time.Millisecond,
	}
}

// RequestFrame builds Modbus RTU frame with CRC16
func RequestFrame(slaveID uint8, funcCode uint8, payload []byte) []byte {
	frame := make([]byte, 2+len(payload)+2)
	frame[0] = slaveID
	frame[1] = funcCode
	copy(frame[2:], payload)

	crc := CRC16(frame[:2+len(payload)])
	binary.LittleEndian.PutUint16(frame[2+len(payload):], crc)
	return frame
}

// ReadHoldingRegisters reads 16-bit registers from slave (Function 0x03)
func (c *Client) ReadHoldingRegisters(slaveID uint8, address uint16, count uint16) ([]uint16, error) {
	if count == 0 || count > 125 {
		return nil, errors.New("E_MODBUS_INVALID_COUNT: count must be 1..125")
	}

	payload := make([]byte, 4)
	binary.BigEndian.PutUint16(payload[0:], address)
	binary.BigEndian.PutUint16(payload[2:], count)

	req := RequestFrame(slaveID, FuncReadHoldingRegisters, payload)

	c.mu.Lock()
	defer c.mu.Unlock()

	c.port.Flush()
	if _, err := c.port.Write(req); err != nil {
		return nil, fmt.Errorf("E_MODBUS_TX: %w", err)
	}

	// Response format: [SlaveID, FuncCode, ByteCount (count*2), Data..., CRC_LO, CRC_HI]
	expectedLen := 3 + int(count)*2 + 2
	resp := make([]byte, expectedLen)
	n, err := c.port.Read(resp)
	if err != nil {
		return nil, fmt.Errorf("E_MODBUS_RX: %w", err)
	}
	if n < expectedLen {
		return nil, fmt.Errorf("E_MODBUS_SHORT_FRAME: got %d bytes, expected %d", n, expectedLen)
	}

	// Validate CRC
	rxCRC := binary.LittleEndian.Uint16(resp[expectedLen-2:])
	calcCRC := CRC16(resp[:expectedLen-2])
	if rxCRC != calcCRC {
		return nil, fmt.Errorf("E_MODBUS_CRC_MISMATCH: got 0x%04X, calculated 0x%04X", rxCRC, calcCRC)
	}

	// Check exception code
	if resp[1] == (FuncReadHoldingRegisters | 0x80) {
		return nil, fmt.Errorf("E_MODBUS_EXCEPTION: code 0x%02X", resp[2])
	}

	byteCount := int(resp[2])
	if byteCount != int(count)*2 {
		return nil, fmt.Errorf("E_MODBUS_PAYLOAD_MISMATCH: byteCount %d != expected %d", byteCount, count*2)
	}

	regs := make([]uint16, count)
	for i := 0; i < int(count); i++ {
		regs[i] = binary.BigEndian.Uint16(resp[3+i*2:])
	}
	return regs, nil
}

// WriteSingleRegister writes a 16-bit register to slave (Function 0x06)
func (c *Client) WriteSingleRegister(slaveID uint8, address uint16, value uint16) error {
	payload := make([]byte, 4)
	binary.BigEndian.PutUint16(payload[0:], address)
	binary.BigEndian.PutUint16(payload[2:], value)

	req := RequestFrame(slaveID, FuncWriteSingleRegister, payload)

	c.mu.Lock()
	defer c.mu.Unlock()

	c.port.Flush()
	if _, err := c.port.Write(req); err != nil {
		return fmt.Errorf("E_MODBUS_TX: %w", err)
	}

	// Echo response: [SlaveID, FuncCode, Addr_HI, Addr_LO, Val_HI, Val_LO, CRC_LO, CRC_HI]
	resp := make([]byte, 8)
	n, err := c.port.Read(resp)
	if err != nil {
		return fmt.Errorf("E_MODBUS_RX: %w", err)
	}
	if n < 8 {
		return fmt.Errorf("E_MODBUS_SHORT_FRAME: got %d bytes, expected 8", n)
	}

	rxCRC := binary.LittleEndian.Uint16(resp[6:])
	calcCRC := CRC16(resp[:6])
	if rxCRC != calcCRC {
		return fmt.Errorf("E_MODBUS_CRC_MISMATCH: got 0x%04X, calculated 0x%04X", rxCRC, calcCRC)
	}

	return nil
}
