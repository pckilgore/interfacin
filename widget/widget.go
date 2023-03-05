// Package widget implements CoreObject for widget.
package widget

import (
	"pckilgore/app/model"
	"pckilgore/app/pointers"

	"pckilgore/app/store/pagination"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ID = model.ID[widget]

// This is the version of this struct where we're all adults, and understand
// that once we mutate this fucker outside the service it really isn't something
// we can trust any more.
//
// If we really want to make something read only, though, it's as simple as
// hiding the field and writing the appropriate getter.
type widget struct {
	ID   ID
	Name string
}

type DatabaseWidget struct {
	ID   string
	Name string
}

type WidgetParams struct {
	IDs *[]model.ID[widget]

	pagination.Pagination
}

func (DatabaseWidget) TableName() string {
	return "widgets"
}

func (w WidgetParams) GormFilter(db *gorm.DB) *gorm.DB {
	return db
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
	w.ID = getIDFromDatabaseID(id)
}

func getIDFromDatabaseID(dbID string) ID {
	return model.NewID[widget](dbID)
}

func maybeGetIDFromDatabaseID(dbID *string) *model.ID[widget] {
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
	w.ID = getIDFromDatabaseID(d.ID)
	w.Name = d.Name
	return w, nil
}

func (d DatabaseWidget) GetID() string {
	return d.ID
}

func CreateID() ID {
	return model.NewID[widget](DatabaseWidget{}.NewID())
}
