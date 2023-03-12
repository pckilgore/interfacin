package gormstore

import (
	"context"
	"pckilgore/app/pointers"
	"pckilgore/app/store"

	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func createImpl[D store.Storable](
	c context.Context,
	db *gorm.DB,
	r store.Retriever[D],
	m D,
) (*D, error) {
	db = db.WithContext(c)

	result := db.Create(m)
	if result.Error != nil {
		return nil, errors.Wrap(result.Error, "failed to create record")
	}

	// Re-fetch in case there are calculated fields.
	retrieved, found, err := r.Retrieve(c, m.GetID())
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve newly-created model")
	} else if !found {
		return nil, errors.New("failed to find newly-created model")
	}

	return retrieved, nil
}

func retrieveImpl[D store.Storable](
	c context.Context,
	db *gorm.DB,
	r store.Retriever[D],
	id string,
) (*D, bool, error) {
	db = db.WithContext(c)
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

func deleteImpl[D store.Storable](c context.Context, db *gorm.DB, id string) (bool, error) {
	db = db.WithContext(c)

	result := db.Where("id = ?", id).Delete(new(D))
	if result.Error != nil {
		return false, errors.Wrap(result.Error, "failed to delete record")
	} else if result.RowsAffected == 0 {
		return false, nil
	}

	return true, nil
}

func listImpl[D store.Storable, P GormParameters](
	c context.Context,
	db *gorm.DB,
	params P,
) (store.ListResponse[D], error) {
	db = db.WithContext(c)
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
