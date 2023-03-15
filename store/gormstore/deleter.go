package gormstore

import (
	"context"
	"pckilgore/app/store"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type Deleter[D store.Storable] struct {
	db *gorm.DB
}

func NewDeleter[D store.Storable](db *gorm.DB) *Deleter[D] {
	return &Deleter[D]{db: db}
}

// Delete a model.
func (s *Deleter[D]) Delete(c context.Context, id string) (bool, error) {
	db := s.db.WithContext(c)

	result := db.Where("id = ?", id).Delete(new(D))
	if result.Error != nil {
		return false, errors.Wrap(result.Error, "failed to delete record")
	} else if result.RowsAffected == 0 {
		return false, nil
	}

	return true, nil
}
