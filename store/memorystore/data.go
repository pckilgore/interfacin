package memorystore

import "sync"

type data[T any] struct {
	mu    sync.RWMutex
	store map[string]T
}

func NewData[T any](initial map[string]T) *data[T] {
	if initial == nil {
		initial = make(map[string]T)
	}

	return &data[T]{store: initial}
}

type InitialData[T any] map[string]T
