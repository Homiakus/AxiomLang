package collections

import (
	"errors"
	"sync"
)

// Map represents a concurrent-safe or lock-free high performance dictionary
type Map struct {
	items map[string]interface{}
	mu    sync.RWMutex
}

func NewMap() *Map {
	return &Map{
		items: make(map[string]interface{}, 32),
	}
}

func (m *Map) Set(key string, val interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.items[key] = val
}

func (m *Map) Get(key string) (interface{}, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	v, ok := m.items[key]
	return v, ok
}

func (m *Map) Delete(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.items, key)
}

func (m *Map) Len() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.items)
}

// Result encapsulates Ok(T) or Error(E) without panics
type Result[T any] struct {
	value T
	err   error
	isOk  bool
}

func Ok[T any](val T) Result[T] {
	return Result[T]{value: val, isOk: true}
}

func Err[T any](err error) Result[T] {
	return Result[T]{err: err, isOk: false}
}

func (r Result[T]) IsOk() bool {
	return r.isOk
}

func (r Result[T]) Unwrap() (T, error) {
	if !r.isOk {
		var zero T
		return zero, r.err
	}
	return r.value, nil
}

func (r Result[T]) UnwrapOr(def T) T {
	if !r.isOk {
		return def
	}
	return r.value
}

// Option encapsulates Some(T) or None
type Option[T any] struct {
	value   T
	hasValue bool
}

func Some[T any](val T) Option[T] {
	return Option[T]{value: val, hasValue: true}
}

func None[T any]() Option[T] {
	return Option[T]{hasValue: false}
}

func (o Option[T]) IsSome() bool {
	return o.hasValue
}

func (o Option[T]) Unwrap() (T, error) {
	if !o.hasValue {
		var zero T
		return zero, errors.New("cannot unwrap None")
	}
	return o.value, nil
}
