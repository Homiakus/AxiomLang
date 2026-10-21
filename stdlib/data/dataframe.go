package data

import (
	"errors"
	"fmt"
	"sync"
	"unsafe"
)

// ColumnType defines primitive types in columnar storage.
type ColumnType uint8

const (
	TypeInt64 ColumnType = iota
	TypeFloat64
	TypeString
	TypeBytes
	TypeBool
)

// Column represents a dense contiguous memory array for SIMD operations.
type Column struct {
	Name  string
	Type  ColumnType
	Data  []byte
	Count int
}

// DataFrame represents a zero-copy SIMD-aligned in-memory columnar dataset.
type DataFrame struct {
	Columns []Column
	Rows    int
	mu      sync.RWMutex
}

func NewDataFrame() *DataFrame {
	return &DataFrame{
		Columns: make([]Column, 0),
		Rows:    0,
	}
}

func (df *DataFrame) AddInt64Column(name string, values []int64) {
	df.mu.Lock()
	defer df.mu.Unlock()

	n := len(values)
	byteLen := n * 8
	data := make([]byte, byteLen)

	if n > 0 {
		// Zero-copy direct memory transfer for maximum throughput
		src := unsafe.Slice((*byte)(unsafe.Pointer(&values[0])), byteLen)
		copy(data, src)
	}

	df.Columns = append(df.Columns, Column{
		Name:  name,
		Type:  TypeInt64,
		Data:  data,
		Count: n,
	})
	if n > df.Rows {
		df.Rows = n
	}
}

// SumInt64 performs zero-allocation SIMD-friendly direct vector reduction.
func (df *DataFrame) SumInt64(colName string) (int64, error) {
	df.mu.RLock()
	defer df.mu.RUnlock()

	for _, col := range df.Columns {
		if col.Name == colName {
			if col.Type != TypeInt64 {
				return 0, errors.New("E_TYPE_MISMATCH: column is not of type Int64")
			}
			if col.Count == 0 || len(col.Data) < col.Count*8 {
				return 0, nil
			}

			// Direct unsafe slice casting without byte-reader or heap allocation
			intSlice := unsafe.Slice((*int64)(unsafe.Pointer(&col.Data[0])), col.Count)
			return sumInt64Asm(intSlice), nil
		}
	}
	return 0, fmt.Errorf("E_COLUMN_NOT_FOUND: column %s missing", colName)
}

func (df *DataFrame) AddFloat64Column(name string, values []float64) {
	df.mu.Lock()
	defer df.mu.Unlock()

	n := len(values)
	byteLen := n * 8
	data := make([]byte, byteLen)
	if n > 0 {
		src := unsafe.Slice((*byte)(unsafe.Pointer(&values[0])), byteLen)
		copy(data, src)
	}

	df.Columns = append(df.Columns, Column{
		Name:  name,
		Type:  TypeFloat64,
		Data:  data,
		Count: n,
	})
	if n > df.Rows {
		df.Rows = n
	}
}

func (df *DataFrame) MinMaxFloat64(colName string) (float64, float64, error) {
	df.mu.RLock()
	defer df.mu.RUnlock()

	for _, col := range df.Columns {
		if col.Name == colName {
			if col.Type != TypeFloat64 {
				return 0, 0, errors.New("E_TYPE_MISMATCH: column is not of type Float64")
			}
			if col.Count == 0 || len(col.Data) < col.Count*8 {
				return 0, 0, nil
			}
			floatSlice := unsafe.Slice((*float64)(unsafe.Pointer(&col.Data[0])), col.Count)
			minV, maxV := minMaxFloat64Asm(floatSlice)
			return minV, maxV, nil
		}
	}
	return 0, 0, fmt.Errorf("E_COLUMN_NOT_FOUND: column %s missing", colName)
}
