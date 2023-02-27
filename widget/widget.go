// Package widget implements CoreObject for widget.
package widget

import (
	"context"
  "github.com/pkg/errors"
	"pckilgore/app/model"
	"pckilgore/app/pointers"
	"pckilgore/app/store"
	"strings"

	"gorm.io/gorm"
)

var name string = "widget"

// OK, so the issue I see here is that if widget is private, only the widget
// package can do business logic on it. So the store shit will have to go there.
// Honestly, what is the point of a model, if the only way it can be accessed is 
// in this package. What do I do if a service needs to do things with bot...
//
// Wait is that it?
//
// Is the role of a service composition? Does this need to send DTO's *out* too?
//
// So the database stuff is tough, we might need an interface for it, but this
// might actually start making sense.
//
// something like a function heavy design where pattern is:
//  func Name(ctx, store, ...args) so any writes are abstracted by the store,
//  and the result is a widget, which is just accepted by other functions to do
//  things to avoid weird activerecordy shit.
//
//
//  Or, maybe first version: just leave the fields on widget's public.
//
//  They can't be created invalidly at least, even if they can be read/mutated.
//
//  Or just implement getter methods everywhere but yuck.
type widget struct {
  ID model.ID[widget]
  Name string
}

type DatabaseWidget struct {
  ID string
  Name string
}  

func (DatabaseWidget) TableName() string {
  return "widgets"
}

// WidgetTemplate describes desired mutation on an Widget. Nil values indicate
// no mutation is desired.
type WidgetTemplate struct {
  ID *string
  Name *string
}

func (w widget) GetID() model.ID[widget] {
  return w.ID
}

func (w widget) ObjectName() string {
  return "widget"
}

func (w *widget) SetId(id string) { 
  var b strings.Builder
  b.WriteString(w.ObjectName())
  b.WriteString("_")
  b.WriteString(id)
  w.ID = model.CastID[widget](b.String())
}

func (widget) Serialize(w widget) DatabaseWidget {
  return DatabaseWidget{
    ID: string(w.ID),
    Name: w.Name,
  }
}

func (widget) Deserialize(d DatabaseWidget) (*widget, error) {
  var result widget
  result.SetId(d.ID)
  result.Name = d.Name
  return &result, nil
}

func New(w WidgetTemplate) widget {
  var m widget 
  m.SetId(pointers.GetWithDefault(w.ID, "123"))
  return m
}

func (d DatabaseWidget) GetID() string {
  return d.ID
}

type Service struct {
  store store.DBStore[widget, DatabaseWidget]
}

func NewService(db *gorm.DB) Service {
  return Service{
    store: store.NewDBStore[DatabaseWidget, widget](db, DatabaseWidget{}, widget{}),
  }
}

func (s Service) Create(c context.Context, t WidgetTemplate) (*widget, error) {
  var w widget
  // TODO: Pattern for requried fields
  w.Name = pointers.GetWithDefault(t.Name, "Some Widget")
  w.ID =  model.CastID[widget]("testing") //pointers.GetWithDefault(t.Name, "Some Widget")

  // Persist
  res, err := s.store.Create(c, w)
  if err != nil {
    return nil, errors.Wrap(err, "Failed to save created widget")
  }

  return res, nil
}

type WidgetStore = store.Store[widget]
