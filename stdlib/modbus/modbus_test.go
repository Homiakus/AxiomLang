package modbus_test

import (
	"encoding/binary"
	"testing"
	"time"

	"AxiomLang/stdlib/modbus"
	"AxiomLang/stdlib/uart"
)

func TestModbusRTU_ReadHoldingRegisters(t *testing.T) {
	cfg := uart.DefaultConfig("COM_TEST")
	cfg.Timeout = 100 * time.Millisecond
	port, _ := uart.Open(cfg)
	defer port.Close()

	client := modbus.NewClient(port)

	// Background worker responding as slave #1
	go func() {
		for {
			tx := port.DrainTx()
			if len(tx) >= 8 {
				// Expecting Function 0x03
				slaveID := tx[0]
				funcCode := tx[1]
				if slaveID == 1 && funcCode == 0x03 {
					// Simulate returning 2 registers: [0x1234, 0x5678]
					respPayload := []byte{0x04, 0x12, 0x34, 0x56, 0x78}
					resp := make([]byte, 2+len(respPayload)+2)
					resp[0] = slaveID
					resp[1] = funcCode
					copy(resp[2:], respPayload)
					crc := modbus.CRC16(resp[:2+len(respPayload)])
					binary.LittleEndian.PutUint16(resp[2+len(respPayload):], crc)

					port.InjectRx(resp)
					return
				}
			}
			time.Sleep(5 * time.Millisecond)
		}
	}()

	regs, err := client.ReadHoldingRegisters(1, 0x0000, 2)
	if err != nil {
		t.Fatalf("ReadHoldingRegisters failed: %v", err)
	}

	if len(regs) != 2 || regs[0] != 0x1234 || regs[1] != 0x5678 {
		t.Fatalf("unexpected registers: %v", regs)
	}
}

func TestModbusRTU_CRC16(t *testing.T) {
	// Standard Modbus test vector: 01 03 00 00 00 0A -> CRC = C5 CD
	packet := []byte{0x01, 0x03, 0x00, 0x00, 0x00, 0x0A}
	crc := modbus.CRC16(packet)
	if crc != 0xCDC5 {
		t.Fatalf("expected CRC 0xCDC5, got 0x%04X", crc)
	}
}
