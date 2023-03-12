package gormstore

import (
	"context"
	"pckilgore/app/store"

	"gorm.io/gorm"
)

type GormParameters interface {
	store.Parameterized
	// GormFilter converts P to gorm query constraints.
	GormFilter(*gorm.DB) *gorm.DB
}

func New[D store.Storable, P GormParameters](db *gorm.DB) *DBStore[D, P] {
	return &DBStore[D, P]{db: db, dbModel: *new(D)}
}

type DBStore[D store.Storable, P GormParameters] struct {
	db      *gorm.DB
	dbModel D
}

// Create serializes a Model into the database. Returns the model after it's
// written, in case the model pushes logic into the database.
func (s *DBStore[D, P]) Create(c context.Context, m D) (*D, error) {
	return createImpl[D](c, s.db, s, m)
}

// Retrieve a model.
func (s *DBStore[D, P]) Retrieve(c context.Context, id string) (*D, bool, error) {
	return retrieveImpl[D](c, s.db, s, id)
}

// Delete a model.
func (s *DBStore[D, P]) Delete(c context.Context, id string) (bool, error) {
	return deleteImpl[D](c, s.db, id)
}

// List a model.
func (s *DBStore[D, P]) List(c context.Context, params P) (store.ListResponse[D], error) {
	return listImpl[D](c, s.db, params)
}
