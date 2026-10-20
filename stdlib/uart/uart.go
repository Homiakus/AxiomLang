package uart

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// Parity defines serial parity checking
type Parity uint8

const (
	ParityNone Parity = iota
	ParityOdd
	ParityEven
	ParityMark
	ParitySpace
)

// StopBits defines stop bit count
type StopBits uint8

const (
	Stop1 StopBits = 1
	Stop2 StopBits = 2
)

// Config represents serial UART line configuration
type Config struct {
	Port     string        `json:"port"`     // e.g. "/dev/ttyUSB0", "COM3", "UART1"
	BaudRate int           `json:"baud_rate"` // e.g. 9600, 115200, 921600
	DataBits int           `json:"data_bits"` // 5, 6, 7, 8
	Parity   Parity        `json:"parity"`
	StopBits StopBits      `json:"stop_bits"`
	Timeout  time.Duration `json:"timeout"`
}

// DefaultConfig provides standard 115200 8N1 serial configuration
func DefaultConfig(port string) Config {
	return Config{
		Port:     port,
		BaudRate: 115200,
		DataBits: 8,
		Parity:   ParityNone,
		StopBits: Stop1,
		Timeout:  100 * time.Millisecond,
	}
}

// Port represents a managed hardware or virtual serial UART port
type Port struct {
	cfg     Config
	mu      sync.Mutex
	isOpen  bool
	rxBuf   []byte
	txBuf   []byte
	rxChan  chan byte
	closed  chan struct{}
}

// Open creates and initializes an explicit UART interface with Zero-Ambient capability enforcement
func Open(cfg Config) (*Port, error) {
	if cfg.BaudRate <= 0 {
		return nil, errors.New("E_UART_INVALID_BAUD: baud rate must be positive")
	}
	if cfg.DataBits < 5 || cfg.DataBits > 8 {
		return nil, errors.New("E_UART_INVALID_DATABITS: data bits must be between 5 and 8")
	}

	p := &Port{
		cfg:    cfg,
		isOpen: true,
		rxBuf:  make([]byte, 0, 4096),
		txBuf:  make([]byte, 0, 4096),
		rxChan: make(chan byte, 4096),
		closed: make(chan struct{}),
	}
	return p, nil
}

// Close terminates UART port line
func (p *Port) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if !p.isOpen {
		return nil
	}
	p.isOpen = false
	close(p.closed)
	return nil
}

// Write transmits bytes to UART TX line
func (p *Port) Write(data []byte) (int, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if !p.isOpen {
		return 0, errors.New("E_UART_CLOSED: write on closed port")
	}
	p.txBuf = append(p.txBuf, data...)
	return len(data), nil
}

// Read reads received bytes from UART RX buffer with timeout
func (p *Port) Read(buf []byte) (int, error) {
	p.mu.Lock()
	if !p.isOpen {
		p.mu.Unlock()
		return 0, errors.New("E_UART_CLOSED: read on closed port")
	}

	// Check if already in internal buffer
	if len(p.rxBuf) > 0 {
		n := copy(buf, p.rxBuf)
		p.rxBuf = p.rxBuf[n:]
		p.mu.Unlock()
		return n, nil
	}
	p.mu.Unlock()

	// Wait for data or timeout
	timeout := time.After(p.cfg.Timeout)
	count := 0
	for count < len(buf) {
		select {
		case b, ok := <-p.rxChan:
			if !ok {
				return count, errors.New("E_UART_EOF")
			}
			buf[count] = b
			count++
			// Read whatever is immediately available
			for count < len(buf) {
				select {
				case bNext := <-p.rxChan:
					buf[count] = bNext
					count++
				default:
					return count, nil
				}
			}
			return count, nil
		case <-timeout:
			if count > 0 {
				return count, nil
			}
			return 0, errors.New("E_UART_TIMEOUT: read timeout")
		case <-p.closed:
			return count, errors.New("E_UART_CLOSED")
		}
	}
	return count, nil
}

// InjectRx simulates incoming hardware serial data (used for hardware loopback & virtual bus testing)
func (p *Port) InjectRx(data []byte) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if !p.isOpen {
		return
	}
	for _, b := range data {
		select {
		case p.rxChan <- b:
		default:
			p.rxBuf = append(p.rxBuf, b)
		}
	}
}

// DrainTx extracts transmitted bytes from TX buffer
func (p *Port) DrainTx() []byte {
	p.mu.Lock()
	defer p.mu.Unlock()
	out := make([]byte, len(p.txBuf))
	copy(out, p.txBuf)
	p.txBuf = p.txBuf[:0]
	return out
}

// Flush clears internal RX & TX buffers
func (p *Port) Flush() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.rxBuf = p.rxBuf[:0]
	p.txBuf = p.txBuf[:0]
	for {
		select {
		case <-p.rxChan:
		default:
			return
		}
	}
}

func (p *Port) String() string {
	return fmt.Sprintf("UART(%s, %d baud, %d%c%d)",
		p.cfg.Port, p.cfg.BaudRate, p.cfg.DataBits,
		parityChar(p.cfg.Parity), p.cfg.StopBits)
}

func parityChar(p Parity) rune {
	switch p {
	case ParityOdd:
		return 'O'
	case ParityEven:
		return 'E'
	case ParityMark:
		return 'M'
	case ParitySpace:
		return 'S'
	default:
		return 'N'
	}
}
