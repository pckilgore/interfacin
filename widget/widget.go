// Package widget implements CoreObject for widget.
package widget

import (
	"context"
	"pckilgore/app/model"
	"pckilgore/app/pointers"
	"pckilgore/app/store"
	"pckilgore/app/store/gormstore"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

var name string = "widget"

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

func (widget) Serialize(w widget) DatabaseWidget {
	return DatabaseWidget{
		ID:   model.ParseID(w.ID),
		Name: w.Name,
	}
}

func (DatabaseWidget) NewID() string {
	return uuid.NewString()
}

func (widget) Deserialize(d DatabaseWidget) (*widget, error) {
	var result widget
	result.ID = model.NewID[widget](d.ID)
	result.Name = d.Name
	return &result, nil
}

func (d DatabaseWidget) GetID() string {
	return d.ID
}

func createID() model.ID[widget] {
	return model.NewID[widget](DatabaseWidget{}.NewID())
}

type Service struct {
	store store.Store[widget]
}

func NewService(store store.Store[widget]) Service {
	return Service{
		store: store,
	}
}

func NewGormStore(db *gorm.DB) store.Store[widget] {
	return gormstore.New[DatabaseWidget, widget](db)
}

func (s Service) Create(c context.Context, t WidgetTemplate) (*widget, error) {
	var w widget
	w.Name = pointers.GetWithDefault(t.Name, "Some Widget")
	w.ID = createID()

	// Persist
	res, err := s.store.Create(c, w)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to save created widget")
	}

	return res, nil
}

type WidgetStore = store.Store[widget]
