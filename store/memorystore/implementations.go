package memorystore

import (
	"pckilgore/app/pointers"
	"pckilgore/app/store"
	"sort"
	"strings"
	"sync"

	"github.com/pkg/errors"
)

// Implementation of Create for anything implementing storable.
func createImpl[T store.Storable](m *sync.RWMutex, d map[string]T, s T) (*T, error) {
	m.Lock()
	defer m.Unlock()

	if _, exists := d[s.GetID()]; exists {
		return nil, errors.New("a record with that ID already exists")
	}

	d[s.GetID()] = s

	return &s, nil
}

// Implementation of Retrieve for anything implementing storable.
func retrieveImpl[T store.Storable](m *sync.RWMutex, d map[string]T, id string) (*T, bool, error) {
	m.RLock()
	defer m.RUnlock()

	if model, exists := d[id]; exists {
		return &model, true, nil
	}

	return nil, false, nil
}

// Implementation of Delete for anything implementing storable.
func deleteImpl[T store.Storable](m *sync.RWMutex, d map[string]T, id string) (bool, error) {
	m.Lock()
	defer m.Unlock()
	if _, exists := d[id]; exists {
		delete(d, id)
		return true, nil
	}

	return false, nil
}

// Implementation of List for anything implementing storable.
func listImpl[D store.Storable, P MemoryParams[D]](m *sync.RWMutex, data map[string]D, params P) (store.ListResponse[D], error) {
	m.RLock()
	defer m.RUnlock()
	limit := params.Limit()

	var after *string
	var before *string

	if params.After() != nil {
		after = pointers.Make(params.After().Value())
	}
	if params.Before() != nil {
		before = pointers.Make(params.Before().Value())
	}

	if after != nil && before != nil {
		return store.ListResponse[D]{}, store.NewInvalidPaginationParamsErr(
			"invalid parameters (only one of after or before can be set)",
		)
	}

	var result []D
	for _, m := range data {
		result = append(result, m)
	}
	sort.SliceStable(result, func(i, j int) bool {
		return result[i].GetID() < result[j].GetID()
	})

	result = params.MemoryFilter(result)

	startIndex := 0
	endIndex := limit
	if limit > len(result) {
		endIndex = len(result)
	}

	if before != nil {
		if newEnd, found := sort.Find(len(result), func(i int) int {
			return strings.Compare(*before, result[i].GetID())
		}); found {
			endIndex = newEnd
			if newEnd-limit > 0 {
				startIndex = endIndex - limit
			}
		}
	}

	if after != nil {
		if newStart, found := sort.Find(len(result), func(i int) int {
			return strings.Compare(*after, result[i].GetID())
		}); found {
			startIndex = newStart
			endIndex = newStart + limit
			if endIndex > len(result) {
				endIndex = len(result)
			}
		}
	}

	var nextBefore *store.Cursor
	if startIndex > 0 {
		nextBefore = pointers.Make(store.NewCursor(result[startIndex].GetID()))
	}

	var nextAfter *store.Cursor
	if endIndex < len(result) {
		nextAfter = pointers.Make(store.NewCursor(result[endIndex].GetID()))
	}

	return store.ListResponse[D]{
		Items:  result[startIndex:endIndex],
		Count:  len(result),
		After:  nextAfter,
		Before: nextBefore,
	}, nil
}
