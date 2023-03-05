// Memorystore is a toy implementation of store using a in-memory map.
package memorystore

import (
	"context"
	"pckilgore/app/pointers"
	"pckilgore/app/store"
	"sort"
	"strings"

	"github.com/pkg/errors"
)

// Filter applies parameters to data.
type Filter[S store.Storable, P store.Parameterized] func(d []S, params P) []S

func emptyFilter[S store.Storable, P store.Parameterized](d []S, params P) []S {
	return d
}

type memorystore[S store.Storable, P store.Parameterized] struct {
	data   map[string]S
	filter Filter[S, P]
}

func New[S store.Storable, P store.Parameterized](f *Filter[S, P]) memorystore[S, P] {
	filter := emptyFilter[S, P]
	if f != nil {
		filter = *f
	}
	return memorystore[S, P]{data: make(map[string]S), filter: filter}
}

// Store a model.
func (s memorystore[D, P]) Create(_ context.Context, m D) (*D, error) {
	if _, exists := s.data[m.GetID()]; exists {
		return nil, errors.New("a record with that ID already exists")
	}

	s.data[m.GetID()] = m

	return &m, nil
}

// Retrieve a model.
func (s memorystore[D, P]) Retrieve(_ context.Context, id string) (*D, bool, error) {
	if model, exists := s.data[id]; exists {
		return &model, true, nil
	}

	return nil, false, nil
}

// Delete a model.
func (s memorystore[D, P]) Delete(c context.Context, id string) (bool, error) {
	if _, exists := s.data[id]; exists {
		delete(s.data, id)
		return true, nil
	}

	return false, nil
}

// List a model.
func (s memorystore[D, P]) List(c context.Context, params P) (store.ListResponse[D], error) {
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
	for _, m := range s.data {
		result = append(result, m)
	}
	sort.SliceStable(result, func(i, j int) bool {
		return result[i].GetID() < result[j].GetID()
	})

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
