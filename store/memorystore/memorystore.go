// Memorystore is a toy implementation of store using a in-memory map.
package memorystore

import (
	"context"
	"fmt"

	"pckilgore/app/store"
)

type MemoryParams[D store.Storable] interface {
	store.Parameterized

	// MemoryFilter applies parameters to data.
	MemoryFilter(pre []D) (post []D)
}

func NewStore[D store.Storable, P MemoryParams[D]](d ...InitialData[D]) *Store[D, P] {
	var data *data[D]
	if len(d) == 0 {
		data = NewData[D](nil)
	}

	if len(d) > 0 {
		data = NewData(d[0])
		if len(d) > 1 {
			fmt.Println("More than one set of initial data passed to store!! Using first.")
		}
	}

	return &Store[D, P]{
		d: NewDeleter(data),
		r: NewRetriever(data),
		c: NewCreator(data),
		l: NewLister[D, P](data),
	}
}

type Store[D store.Storable, P MemoryParams[D]] struct {
	d store.Deleter[D]
	r store.Retriever[D]
	c store.Creator[D]
	l store.Lister[D, P]
}

func (s *Store[D, P]) Create(c context.Context, m D) (*D, error) {
	return s.c.Create(c, m)
}

func (s *Store[D, P]) Retrieve(c context.Context, id string) (*D, bool, error) {
	return s.r.Retrieve(c, id)
}

func (s *Store[D, P]) Delete(c context.Context, id string) (bool, error) {
	return s.d.Delete(c, id)
}

func (s *Store[D, P]) List(c context.Context, params P) (store.ListResponse[D], error) {
	return s.l.List(c, params)
}

func NewTreeStore[D store.TreeStorable, P MemoryParams[D]](d ...InitialData[D]) *TreeStore[D, P] {
	var data *data[D]
	if len(d) == 0 {
		data = NewData[D](nil)
	}

	if len(d) > 0 {
		data = NewData(d[0])
		if len(d) > 1 {
			fmt.Println("More than one set of initial data passed to store!! Using first.")
		}
	}

	return &TreeStore[D, P]{
		store: &Store[D, P]{
			d: NewDeleter(data),
			r: NewRetriever(data),
			c: NewCreator(data),
			l: NewLister[D, P](data),
		},
		tree: NewTree(data),
	}
}

type TreeStore[D store.TreeStorable, P MemoryParams[D]] struct {
	store store.Store[D, P]
	tree  store.Tree[D]
}

func (s *TreeStore[D, P]) Create(c context.Context, m D) (*D, error) {
	return s.store.Create(c, m)
}

func (s *TreeStore[D, P]) Retrieve(c context.Context, id string) (*D, bool, error) {
	return s.store.Retrieve(c, id)
}

func (s *TreeStore[D, P]) Delete(c context.Context, id string) (bool, error) {
	return s.store.Delete(c, id)
}

func (s *TreeStore[D, P]) List(c context.Context, params P) (store.ListResponse[D], error) {
	return s.store.List(c, params)
}

func (s *TreeStore[D, P]) ListDescendants(c context.Context, rootId string) (store.TreeResponse[D], error) {
	return s.tree.ListDescendants(c, rootId)
}

func (s *TreeStore[D, P]) ListAncestors(c context.Context, rootId string) (store.TreeResponse[D], error) {
	return s.tree.ListAncestors(c, rootId)
}
