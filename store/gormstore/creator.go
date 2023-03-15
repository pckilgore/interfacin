package gormstore

import (
	"context"
	"pckilgore/app/store"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type Creator[D store.Storable] struct {
	db *gorm.DB
	r  store.Retriever[D]
}

func NewCreator[D store.Storable](db *gorm.DB, r store.Retriever[D]) *Creator[D] {
	return &Creator[D]{db: db, r: r}
}

// Create serializes a Model into the database. Returns the model after it's
// written, in case the model pushes logic into the database.
func (s *Creator[D]) Create(c context.Context, m D) (*D, error) {
	db := s.db.WithContext(c)

	result := db.Create(m)
	if result.Error != nil {
		return nil, errors.Wrap(result.Error, "failed to create record")
	}

	// Re-fetch in case there are calculated fields.
	retrieved, found, err := s.r.Retrieve(c, m.GetID())
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve newly-created model")
	} else if !found {
		return nil, errors.New("failed to find newly-created model")
	}

	return retrieved, nil
}
