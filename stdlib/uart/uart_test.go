package uart_test

import (
	"testing"
	"time"

	"AxiomLang/stdlib/uart"
)

func TestUART_Loopback(t *testing.T) {
	cfg := uart.DefaultConfig("/dev/ttyVirtual0")
	cfg.Timeout = 100 * time.Millisecond
	port, err := uart.Open(cfg)
	if err != nil {
		t.Fatalf("failed to open UART: %v", err)
	}
	defer port.Close()

	payload := []byte("HELLO_AXIOM_UART_115200")
	n, err := port.Write(payload)
	if err != nil || n != len(payload) {
		t.Fatalf("failed to write UART: %v", err)
	}

	txData := port.DrainTx()
	if string(txData) != string(payload) {
		t.Fatalf("expected tx %q, got %q", payload, txData)
	}

	// Test RX injection
	port.InjectRx(payload)
	rxBuf := make([]byte, len(payload))
	rxN, err := port.Read(rxBuf)
	if err != nil {
		t.Fatalf("failed to read UART: %v", err)
	}
	if rxN != len(payload) || string(rxBuf) != string(payload) {
		t.Fatalf("expected rx %q, got %q", payload, rxBuf)
	}
}
