package gormstore

import (
	"context"
	"pckilgore/app/pointers"
	"pckilgore/app/store"

	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Lister[D store.Storable, P GormParameters] struct {
	db *gorm.DB
}

func NewLister[D store.Storable, P GormParameters](db *gorm.DB) *Lister[D, P] {
	return &Lister[D, P]{db: db}
}

// List a model.
func (s *Lister[D, P]) List(c context.Context, params P) (store.ListResponse[D], error) {
	db := s.db.WithContext(c)
	limit := params.Limit()
	reverse := false

	after := params.After()
	before := params.Before()
	if after != nil && before != nil {
		return store.ListResponse[D]{}, store.NewInvalidPaginationParamsErr(
			"invalid parameters (only one of after or before can be set)",
		)
	}

	model := *new(D)
	table := model.TableName()
	db = db.Table(table)
	db = params.GormFilter(db)

	var count int64
	result := db.Count(&count)
	if result.Error != nil {
		return store.ListResponse[D]{}, errors.Wrap(result.Error, "failed to get total with pagination")
	}

	if after != nil {
		db = db.Where("id > ?", after.Value()).Order(
			clause.OrderByColumn{Column: clause.Column{Name: "id"}},
		)
	} else if before != nil {
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
