package gormstore

import (
	"context"
	"pckilgore/app/store"

	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Create serializes a Model into the database. Returns the model after it's
// written, in case the model pushes logic into the database.
func (s DBStore[Model, DatabaseModel]) Create(c context.Context, m Model) (*Model, error) {
	db := s.db.WithContext(c)
	serializable := s.model.Serialize(m)

	result := db.Create(serializable)
	if result.Error != nil {
		return nil, errors.Wrap(result.Error, "failed to create record")
	}

	// Re-fetch in case there are calculated fields.
	retrieved, found, err := s.Retrieve(c, serializable.GetID())
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve newly-created model")
	} else if !found {
		return nil, errors.New("failed to find newly-created model")
	}

	return retrieved, nil
}

// Retrieve a model.
func (s DBStore[Model, DatabaseModel]) Retrieve(c context.Context, id string) (*Model, bool, error) {
	db := s.db.WithContext(c)
	query := db.Unscoped()

	var d DatabaseModel
	resp := query.First(
		&d,
		clause.Where{Exprs: []clause.Expression{
			clause.Eq{Column: "id", Value: id},
		}},
	)

	if resp.Error != nil {
		if errors.Is(resp.Error, gorm.ErrRecordNotFound) {
			return nil, false, nil
		}
		return nil, false, errors.Wrap(resp.Error, "failed to retrieve model")
	}

	model, err := s.model.Deserialize(d)
	if err != nil {
		return nil, false, errors.Wrap(err, "deserialization failed")
	}

	return model, true, nil
}

// Retrieve a model.
func (s DBStore[Model, DatabaseModel]) Delete(c context.Context, id string) (bool, error) {
	db := s.db.WithContext(c)

	result := db.Where("id = ?", id).Delete(new(DatabaseModel))
	if result.Error != nil {
		return false, errors.Wrap(result.Error, "failed to delete record")
	} else if result.RowsAffected == 0 {
		return false, nil
	}

	return true, nil
}

type DBStore[Model any, DatabaseModel store.Storable[DatabaseModel]] struct {
	db      *gorm.DB
	dbModel DatabaseModel
	model   store.Serder[Model, DatabaseModel]
}

func New[DatabaseModel store.Storable[DatabaseModel], Model store.Serder[Model, DatabaseModel]](
	db *gorm.DB,
) DBStore[Model, DatabaseModel] {
	return DBStore[Model, DatabaseModel]{
		db:      db,
		dbModel: *new(DatabaseModel),
		model:   *new(Model),
	}
}
