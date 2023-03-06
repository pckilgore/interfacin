// Package widget implements CoreObject for widget.
package widget

import (
	"pckilgore/app/model"
	"pckilgore/app/pointers"

	"pckilgore/app/store/gormstore"
	"pckilgore/app/store/pagination"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// This is the version of this struct where we're all adults, and understand
// that once we mutate this fucker outside the service it really isn't something
// we can trust any more.
//
// If we really want to make something read only, though, it's as simple as
// hiding the field and writing the appropriate getter.
type widget struct {
	ID   model.ID[widget]
	Name string
}

type ID model.ID[widget]

type DatabaseWidget struct {
	ID   string
	Name string
}

type WidgetParams struct {
	IDs *[]ID

	pagination.Pagination
}

func (DatabaseWidget) TableName() string {
	return "widgets"
}

func (w WidgetParams) GormFilter(db *gorm.DB) *gorm.DB {
	return db.Scopes(
		gormstore.ColumnInIDs("id", w.IDs),
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

func (w WidgetParams) MemoryFilter(in []DatabaseWidget) []DatabaseWidget {
	return filterMany(
		in,
		func(item DatabaseWidget) bool {
			return find(w.IDs, func(id ID) bool {
				return getIDFromDatabaseID(item.ID) == id
			})
		},
	)
}

// WidgetTemplate describes desired mutation on an Widget. Nil values indicate
// no mutation is desired.
type WidgetTemplate struct {
	ID   *string
	Name *string
}

func (w widget) Kind() string {
	return "widget"
}

func (w widget) SetID(id string) {
	w.ID = model.ID[widget](getIDFromDatabaseID(id))
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

func (DatabaseWidget) NewID() string {
	return uuid.NewString()
}

func Serialize(w widget) DatabaseWidget {
	return DatabaseWidget{
		ID:   model.ParseID(w.ID),
		Name: w.Name,
	}
}

func Deserialize(d *DatabaseWidget) (*widget, error) {
	w := new(widget)
	w.ID = model.ID[widget](getIDFromDatabaseID(d.ID))
	w.Name = d.Name
	return w, nil
}

func (d DatabaseWidget) GetID() string {
	return d.ID
}

func CreateID() model.ID[widget] {
	return model.NewID[widget](DatabaseWidget{}.NewID())
}
