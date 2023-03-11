package node 

import (
	"pckilgore/app/model"
	"pckilgore/app/pointers"

	"pckilgore/app/store/gormstore"
	"pckilgore/app/store/pagination"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type node struct {
	ID   model.ID[node]
  ParentID model.ID[node]
	Name string
}

type ID model.ID[node]

type DatabaseNode struct {
	ID   string
  ParentID string
	Name string
}

type NodeParams struct {
	IDs *[]ID
	ParentIDs *[]ID

	pagination.Pagination
}

func (d DatabaseNode) GetParentID() string {
	return d.ParentID
}

func (DatabaseNode) TableName() string {
	return "nodes"
}

func (w NodeParams) GormFilter(db *gorm.DB) *gorm.DB {
	return db.Scopes(
		gormstore.ColumnInIDs("id", w.IDs),
		gormstore.ColumnInIDs("parent_id", w.ParentIDs),
	)
}

type filterIn[T any] func(cur T) bool

func filterMany[T any](items []T, filters ...filterIn[T]) []T {
	var result []T
	for _, item := range items {
		for _, f := range filters {
			if f(item) {
				result = append(result, item)
			}
		}
	}

	return result
}

func find[T any](criteria *[]T, p func(t T) bool) bool {
	if criteria == nil {
		return true
	}

	for _, t := range *criteria {
		if p(t) {
			return true
		}
	}

	return false
}

func (w NodeParams) MemoryFilter(in []DatabaseNode) []DatabaseNode {
	return filterMany(
		in,
		func(item DatabaseNode) bool {
			return find(w.IDs, func(id ID) bool {
				return getIDFromDatabaseID(item.ID) == id
			})
		},
	)
}

// NodeTemplate describes desired mutation on an Node. Nil values indicate
// no mutation is desired.
type NodeTemplate struct {
	ID   *string
	Name *string
}

func (w node) Kind() string {
	return "node"
}

func (w node) SetID(id string) {
	w.ID = model.ID[node](getIDFromDatabaseID(id))
}

func getIDFromDatabaseID(dbID string) ID {
	return ID(dbID)
}

func maybeGetIDFromDatabaseID(dbID *string) *ID {
	if dbID != nil {
		return pointers.Make(getIDFromDatabaseID(*dbID))
	}

	return nil
}

func (DatabaseNode) NewID() string {
	return uuid.NewString()
}

func Serialize(w node) DatabaseNode {
	return DatabaseNode{
		ID:   model.ParseID(w.ID),
		Name: w.Name,
	}
}

func Deserialize(d *DatabaseNode) (*node, error) {
	w := new(node)
	w.ID = model.ID[node](getIDFromDatabaseID(d.ID))
	w.Name = d.Name
	return w, nil
}

func (d DatabaseNode) GetID() string {
	return d.ID
}

func CreateID() model.ID[node] {
	return model.NewID[node](DatabaseNode{}.NewID())
}

