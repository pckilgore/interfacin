package memorystore

import (
	"context"
	"errors"
	"sort"
	"sync"

	"pckilgore/app/store"
)

type memorytreestore[S store.TreeStorable, P MemoryParams[S]] struct {
	mu   sync.RWMutex
	data map[string]S
}

func NewTree[S store.TreeStorable, P MemoryParams[S]]() *memorytreestore[S, P] {
	return &memorytreestore[S, P]{data: make(map[string]S)}
}

// Store a model.
func (s *memorytreestore[D, P]) Create(_ context.Context, m D) (*D, error) {
	return createImpl(&s.mu, s.data, m)
}

// Retrieve a model.
func (s *memorytreestore[D, P]) Retrieve(_ context.Context, id string) (*D, bool, error) {
	return retrieveImpl(&s.mu, s.data, id)
}

// Delete a model.
func (s *memorytreestore[D, P]) Delete(_ context.Context, id string) (bool, error) {
	return deleteImpl(&s.mu, s.data, id)
}

// List a model.
func (s *memorytreestore[D, P]) List(_ context.Context, params P) (store.ListResponse[D], error) {
	return listImpl(&s.mu, s.data, params)
}

func (s *memorytreestore[D, P]) ListAncestors(c context.Context, rootId string) (store.TreeResponse[D], error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	next, ok := s.data[rootId]
	if !ok {
		return store.TreeResponse[D]{}, errors.New("root not found")
	}

	// Follow the pointers!
	layers := []store.Layer[D]{{PathLength: 0, Items: []D{next}}}

	height := 1
	for next.GetParentID() != nil {
		id := next.GetParentID()
		if maybeNext, ok := s.data[*id]; ok {
			layers = append(layers, store.Layer[D]{PathLength: height, Items: []D{maybeNext}})
			next = maybeNext
		} else {
			return store.TreeResponse[D]{}, errors.New("parent referenced but not found")
		}
	}

	return store.TreeResponse[D]{
		Layers: layers,
		Count:  len(layers),
	}, nil
}

func (s *memorytreestore[D, P]) ListDescendants(c context.Context, rootId string) (store.TreeResponse[D], error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	start, ok := s.data[rootId]
	if !ok {
		return store.TreeResponse[D]{}, errors.New("root not found")
	}

	// memoize parentId => []children
	mapParentChildren := make(map[string][]string, len(s.data))
	for _, node := range s.data {
		parentId := node.GetParentID()
		if parentId == nil {
			continue
		}

		children, ok := mapParentChildren[*parentId]
		if !ok {
			mapParentChildren[*parentId] = []string{node.GetID()}
		} else {
			mapParentChildren[*parentId] = append(children, node.GetID())
		}
	}

	// More efficient to insert each child in correct order above, but fuck it.
	for _, children := range mapParentChildren {
		sort.SliceStable(children, func(i, j int) bool {
			return children[i] < children[j]
		})
	}

	layers := []store.Layer[D]{{PathLength: 0, Items: []D{start}}}
	count := 1

	// I mean, sure use a real queue if you want better perf, but we're ok here.
	queue := []string{start.GetID()}
	depth := 1
	for len(queue) > 0 {
		next := queue[0]
		queue = queue[1:] // Deuque.

		var items []D
		for _, childId := range mapParentChildren[next] {
			count++
			items = append(items, s.data[childId])
			queue = append(queue, childId)
		}

		layers = append(layers, store.Layer[D]{PathLength: depth, Items: items})
		depth++
	}

	return store.TreeResponse[D]{
		Layers: layers,
		Count:  count,
	}, nil
}
