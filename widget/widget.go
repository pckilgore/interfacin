// Package widget implements CoreObject for widget.
package widget

import (
	"pckilgore/app/model"
	"pckilgore/app/store"

	"github.com/google/uuid"
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

type DatabaseWidget struct {
	ID   string
	Name string
}

type WidgetParams struct {
	IDs *[]model.ID[widget]

	store.Pagination
}

func (DatabaseWidget) TableName() string {
	return "widgets"
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
	w.ID = model.NewID[widget](id)
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
	w.ID = model.NewID[widget](d.ID)
	w.Name = d.Name
	return w, nil
}

func (d DatabaseWidget) GetID() string {
	return d.ID
}

func createID() model.ID[widget] {
	return model.NewID[widget](DatabaseWidget{}.NewID())
}
