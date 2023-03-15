package memorystore

import (
	"context"
	"pckilgore/app/pointers"
	"pckilgore/app/store"
	"sort"
	"strings"
)

type Lister[D store.Storable, P MemoryParams[D]] struct {
	d *data[D]
}

func NewLister[D store.Storable, P MemoryParams[D]](d *data[D]) *Lister[D, P] {
	return &Lister[D, P]{d: d}
}

func (s *Lister[D, P]) List(_ context.Context, params P) (store.ListResponse[D], error) {
	s.d.mu.RLock()
	defer s.d.mu.RUnlock()
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
	for _, m := range s.d.store {
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
