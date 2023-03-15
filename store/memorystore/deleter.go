package memorystore

import (
	"context"

	"pckilgore/app/store"
)

type Deleter[D store.Storable] struct {
	d *data[D]
}

func NewDeleter[D store.Storable](d *data[D]) *Deleter[D] {
	return &Deleter[D]{d: d}
}

func (deleter *Deleter[D]) Delete(_ context.Context, id string) (bool, error) {
	deleter.d.mu.Lock()
	defer deleter.d.mu.Unlock()
	if _, exists := deleter.d.store[id]; exists {
		delete(deleter.d.store, id)
		return true, nil
	}

	return false, nil
}
