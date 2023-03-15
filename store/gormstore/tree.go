package gormstore

import (
	"context"
	"pckilgore/app/store"

	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/WinterYukky/gorm-extra-clause-plugin/exclause"
)

type layer struct {
	PathLength int
}

type Tree[D store.TreeStorable] struct {
	db *gorm.DB
}

func NewTree[D store.TreeStorable](db *gorm.DB) *Tree[D] {
	return &Tree[D]{db: db}
}

func (s *Tree[D]) ListAncestors(c context.Context, rootId string) (store.TreeResponse[D], error) {
	db := s.db.WithContext(c)
	model := *new(D)

	rows, err := db.Clauses(
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
		return store.TreeResponse[D]{}, errors.Wrap(err, "failed to get rows")
	}

	count := 0
	layerMap := make(map[int]*store.Layer[D])
	for rows.Next() {
		// Grab path_length column.
		var l layer
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

func (s *Tree[D]) ListDescendants(c context.Context, rootId string) (store.TreeResponse[D], error) {
	db := s.db.WithContext(c)
	model := *new(D)

	rows, err := db.Clauses(
		exclause.With{
			Recursive: true,
			CTEs: []exclause.CTE{
				{
					Name: "descendants",
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
										clause.Table{Name: "possible_children"},
										clause.Table{Name: "descendants"},
									).
									Table("descendants").
									Joins(
										"join ? on ?.? = ?.?",
										clause.Table{Name: model.TableName(), Alias: "possible_children"},
										clause.Table{Name: "descendants"},
										clause.Column{Name: "id"},
										clause.Table{Name: "possible_children"},
										clause.Column{Name: model.GetParentIDField()},
									),
							)),
					},
				},
			},
		},
	).
		Table("descendants").
		Order("path_length").
		Rows()

	defer rows.Close()

	if err != nil {
		return store.TreeResponse[D]{}, errors.Wrap(err, "failed to get rows")
	}

	count := 0
	layerMap := make(map[int]*store.Layer[D])
	for rows.Next() {
		// Grab path_length column.
		var l layer
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
