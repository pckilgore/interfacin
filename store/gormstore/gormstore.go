package gormstore

import (
	"context"
	"pckilgore/app/store"

	extraClausePlugin "github.com/WinterYukky/gorm-extra-clause-plugin"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type GormParameters interface {
	store.Parameterized
	// GormFilter converts P to gorm query constraints.
	GormFilter(*gorm.DB) *gorm.DB
}

func NewStore[D store.Storable, P GormParameters](db *gorm.DB) *Store[D, P] {
	r := NewRetriever[D](db)
	return &Store[D, P]{
		r: r,
		c: NewCreator[D](db, r),
		d: NewDeleter[D](db),
		l: NewLister[D, P](db),
	}
}

type Store[D store.Storable, P GormParameters] struct {
	r store.Retriever[D]
	c store.Creator[D]
	d store.Deleter[D]
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

func NewTreeStore[D store.TreeStorable, P GormParameters](db *gorm.DB) (*TreeStore[D, P], error) {
	err := db.Use(extraClausePlugin.New())
	if err != nil {
		return nil, errors.Wrap(err, "could not add required plugins to gorm store")
	}

	return &TreeStore[D, P]{s: NewStore[D, P](db), t: NewTree[D](db)}, nil
}

type TreeStore[D store.TreeStorable, P GormParameters] struct {
	s store.Store[D, P]
	t store.Tree[D]
}

func (s *TreeStore[D, P]) Create(c context.Context, m D) (*D, error) {
	return s.s.Create(c, m)
}

func (s *TreeStore[D, P]) Retrieve(c context.Context, id string) (*D, bool, error) {
	return s.s.Retrieve(c, id)
}

func (s *TreeStore[D, P]) Delete(c context.Context, id string) (bool, error) {
	return s.s.Delete(c, id)
}

func (s *TreeStore[D, P]) List(c context.Context, params P) (store.ListResponse[D], error) {
	return s.s.List(c, params)
}

func (s *TreeStore[D, P]) ListDescendants(c context.Context, rootId string) (store.TreeResponse[D], error) {
	return s.t.ListDescendants(c, rootId)
}

func (s *TreeStore[D, P]) ListAncestors(c context.Context, rootId string) (store.TreeResponse[D], error) {
	return s.t.ListAncestors(c, rootId)
}
