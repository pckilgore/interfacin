package gormstore

import (
	"context"
	"pckilgore/app/pointers"
	"pckilgore/app/store"

	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Create serializes a Model into the database. Returns the model after it's
// written, in case the model pushes logic into the database.
func (s DBStore[D, P]) Create(c context.Context, m D) (*D, error) {
	db := s.db.WithContext(c)

	result := db.Create(m)
	if result.Error != nil {
		return nil, errors.Wrap(result.Error, "failed to create record")
	}

	// Re-fetch in case there are calculated fields.
	retrieved, found, err := s.Retrieve(c, m.GetID())
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve newly-created model")
	} else if !found {
		return nil, errors.New("failed to find newly-created model")
	}

	return retrieved, nil
}

// Retrieve a model.
func (s DBStore[D, P]) Retrieve(c context.Context, id string) (*D, bool, error) {
	db := s.db.WithContext(c)
	query := db.Unscoped()

	var d D
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

	return &d, true, nil
}

// Delete a model.
func (s DBStore[D, P]) Delete(c context.Context, id string) (bool, error) {
	db := s.db.WithContext(c)

	result := db.Where("id = ?", id).Delete(new(D))
	if result.Error != nil {
		return false, errors.Wrap(result.Error, "failed to delete record")
	} else if result.RowsAffected == 0 {
		return false, nil
	}

	return true, nil
}

// List a model.
func (s DBStore[D, P]) List(c context.Context, params P) (store.ListResponse[D], error) {
	db := s.db.WithContext(c)
	limit := params.Limit()
	reverse := false

	model := *new(D)
	table := model.TableName()
	db = db.Table(table)
	db = params.GormFilter(db)

	var count int64
	result := db.Count(&count)
	if result.Error != nil {
		return store.ListResponse[D]{}, errors.Wrap(result.Error, "failed to get total with pagination")
	}

	if after := params.After(); after != nil {
		db = db.Where("id > ?", after.Value()).Order(
			clause.OrderByColumn{Column: clause.Column{Name: "id"}},
		)
	} else if before := params.Before(); before != nil {
		db = db.Where("id < ?", before.Value()).Order(
			clause.OrderByColumn{Column: clause.Column{Name: "id"}, Desc: true},
		)
		reverse = true
	}

	var leftToPaginate int64
	result = db.Count(&leftToPaginate)
	if result.Error != nil {
		return store.ListResponse[D]{}, errors.Wrap(result.Error, "failed to get total with pagination")
	}

	var modelList []D
	result = db.Limit(limit).Find(&modelList)
	if result.Error != nil {
		return store.ListResponse[D]{}, errors.Wrapf(result.Error, "failed to list %s", table)
	}

	if reverse {
		var reversed []D
		for i := len(modelList) - 1; i >= 0; i-- {
			reversed = append(reversed, modelList[i])
		}
		modelList = reversed
	}

	var nextBefore *store.Cursor
	var nextAfter *store.Cursor

	more := len(modelList) < int(leftToPaginate)

	if params.After() != nil && len(modelList) > 0 {
		nextBefore = pointers.Make(store.NewCursor(modelList[0].GetID()))
		if more {
			nextAfter = pointers.Make(store.NewCursor(modelList[len(modelList)-1].GetID()))
		}
	} else if params.Before() != nil && len(modelList) > 0 {
		nextAfter = pointers.Make(store.NewCursor(modelList[len(modelList)-1].GetID()))
		if more {
			nextBefore = pointers.Make(store.NewCursor(modelList[0].GetID()))
		}
	} else if more {
		nextAfter = pointers.Make(store.NewCursor(modelList[len(modelList)-1].GetID()))
	}

	return store.ListResponse[D]{
		Items:  modelList,
		Count:  int(count),
		After:  nextAfter,
		Before: nextBefore,
	}, nil
}

type GormParameters interface {
	store.Parameterized

	// GormFilter converts P to gorm query constraints.
	GormFilter(*gorm.DB) *gorm.DB
}

type DBStore[D store.Storable, P GormParameters] struct {
	db      *gorm.DB
	dbModel D
}

func New[D store.Storable, P GormParameters](db *gorm.DB) DBStore[D, P] {
	return DBStore[D, P]{db: db, dbModel: *new(D)}
}
