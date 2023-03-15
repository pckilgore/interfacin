package gormstore

import (
	"context"
	"pckilgore/app/store"

	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Retriever[D store.Storable] struct {
	db *gorm.DB
}

func NewRetriever[D store.Storable](db *gorm.DB) *Retriever[D] {
	return &Retriever[D]{db: db}
}

// Retrieve a model.
func (s *Retriever[D]) Retrieve(c context.Context, id string) (*D, bool, error) {
	db := s.db.WithContext(c)
	query := db.Unscoped()

	var d D
	resp := query.First(
		&d,
		clause.Where{
			Exprs: []clause.Expression{
				clause.Eq{Column: "id", Value: id},
			},
		},
	)

	if resp.Error != nil {
		if errors.Is(resp.Error, gorm.ErrRecordNotFound) {
			return nil, false, nil
		}
		return nil, false, errors.Wrap(resp.Error, "failed to retrieve model")
	}

	return &d, true, nil
}
