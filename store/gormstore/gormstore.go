package gormstore

import (
	"context"
	"pckilgore/app/pointers"
	"pckilgore/app/store"

	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	extraClausePlugin "github.com/WinterYukky/gorm-extra-clause-plugin"
	"github.com/WinterYukky/gorm-extra-clause-plugin/exclause"
)

// Create serializes a Model into the database. Returns the model after it's
// written, in case the model pushes logic into the database.
func (s *DBStore[D, P]) Create(c context.Context, m D) (*D, error) {
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
func (s *DBStore[D, P]) Retrieve(c context.Context, id string) (*D, bool, error) {
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

// Delete a model.
func (s *DBStore[D, P]) Delete(c context.Context, id string) (bool, error) {
	db := s.db.WithContext(c)

	result := db.Where("id = ?", id).Delete(new(D))
	if result.Error != nil {
		return false, errors.Wrap(result.Error, "failed to delete record")
	} else if result.RowsAffected == 0 {
		return false, nil
	}

	return true, nil
}

type Layer struct {
	PathLength int `gorm:""`
}

func (s *DBStore[D, P]) ListAncestors(c context.Context, rootId string) (store.TreeResponse[D], error) {
	db := s.db.WithContext(c)
	model := *new(D)

	rows, err := db.Debug().Clauses(
		exclause.With{
			Recursive: true,
			CTEs: []exclause.CTE{
				{
					Name: "ancestors",
					Subquery: exclause.Subquery{
						DB: db.
							Select(
								"?.*, 0 AS path_length",
								clause.Table{Name: model.TableName()}).
							Table(model.TableName()).
							Where(
								"?.? = ?",
								clause.Table{Name: model.TableName()},
								clause.Column{Name: "id"},
								rootId,
							).
							Clauses(exclause.NewUnion(
								"ALL ?",
								db.
									Select(
										"?.*, ?.path_length + 1 AS path_length",
										clause.Table{Name: "possible_parents"},
										clause.Table{Name: "ancestors"},
									).
									Table("ancestors").
									Joins(
										"join ? on ?.? = ?.?",
										clause.Table{Name: model.TableName(), Alias: "possible_parents"},
										clause.Table{Name: "ancestors"},
										clause.Column{Name: model.GetParentIDField()},
										clause.Table{Name: "possible_parents"},
										clause.Column{Name: "id"},
									),
							)),
					},
				},
			},
		},
	).
		Table("ancestors").
		Order("path_length").
		Rows()
	defer rows.Close()

	if err != nil {
		return store.TreeResponse[D]{}, errors.New("failed to get rows")
	}

	count := 0
	layerMap := make(map[int]*store.Layer[D])
	for rows.Next() {
		// Grab path_length column.
		var l Layer
		err := db.ScanRows(rows, &l)
		if err != nil {
			return store.TreeResponse[D]{}, errors.New("failed to scan path_length")
		}

		// Grab other cols into model.
		var model D
		err = db.ScanRows(rows, &model)
		if err != nil {
			return store.TreeResponse[D]{}, errors.New("failed to scan model")
		}
		count++

		if layer, layerExists := layerMap[l.PathLength]; layerExists {
			layer.Items = append(layer.Items, model)
		} else {
			layerMap[l.PathLength] = &store.Layer[D]{
				Items:      []D{model},
				PathLength: l.PathLength,
			}
		}

	}

	layers := make([]store.Layer[D], len(layerMap))
	for _, v := range layerMap {
		layers[v.PathLength] = *v
	}

	return store.TreeResponse[D]{
		Layers: layers,
		Count:  count,
	}, nil
}

func (s *DBStore[D, P]) ListDescendants(c context.Context, rootId string) (store.TreeResponse[D], error) {
	//db := s.db.WithContext(c)
	return store.TreeResponse[D]{}, nil
}

// List a model.
func (s *DBStore[D, P]) List(c context.Context, params P) (store.ListResponse[D], error) {
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

type GormParameters interface {
	store.Parameterized

	// GormFilter converts P to gorm query constraints.
	GormFilter(*gorm.DB) *gorm.DB
}

type DBStore[D store.TreeStorable, P GormParameters] struct {
	db      *gorm.DB
	dbModel D
}

func New[D store.TreeStorable, P GormParameters](db *gorm.DB) (*DBStore[D, P], error) {
	err := db.Use(extraClausePlugin.New())
	if err != nil {
		return nil, errors.Wrap(err, "could not add required plugins to gorm store")
	}
	return &DBStore[D, P]{db: db, dbModel: *new(D)}, nil
}
