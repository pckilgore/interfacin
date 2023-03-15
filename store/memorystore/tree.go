package memorystore

import (
	"context"
	"pckilgore/app/store"
	"sort"

	"github.com/pkg/errors"
)

type Tree[D store.TreeStorable] struct {
	d *data[D]
}

func NewTree[D store.TreeStorable](d *data[D]) *Tree[D] {
	return &Tree[D]{d: d}
}

func (t *Tree[D]) ListAncestors(c context.Context, rootId string) (store.TreeResponse[D], error) {
	t.d.mu.RLock()
	defer t.d.mu.RUnlock()

	next, ok := t.d.store[rootId]
	if !ok {
		return store.TreeResponse[D]{}, errors.New("root not found")
	}

	// Follow the pointers!
	layers := []store.Layer[D]{{PathLength: 0, Items: []D{next}}}

	height := 1
	for next.GetParentID() != nil {
		id := next.GetParentID()
		if maybeNext, ok := t.d.store[*id]; ok {
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

func (t *Tree[D]) ListDescendants(c context.Context, rootId string) (store.TreeResponse[D], error) {
	t.d.mu.RLock()
	defer t.d.mu.RUnlock()

	start, ok := t.d.store[rootId]
	if !ok {
		return store.TreeResponse[D]{}, errors.New("root not found")
	}

	// memoize parentId => []children
	mapParentChildren := make(map[string][]string, len(t.d.store))
	for _, node := range t.d.store {
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
			items = append(items, t.d.store[childId])
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
