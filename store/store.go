// Interfaces for writing to database.
package store

import (
	"context"
  "fmt"

  "github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

type Serializer[Model any, DatabaseModel any] interface {
  Serialize(m Model) DatabaseModel
}

type Deserializer[Model any, DatabaseModel any] interface {
  // Error because the database is not necessarily a trustworthy source.
  Deserialize(d DatabaseModel) (*Model, error)
}

type Serder[Model any, DatabaseModel any] interface {
  Serializer[Model, DatabaseModel]
  Deserializer[Model, DatabaseModel]
}

type Retriever[Model any] interface {
  Retrieve(ctx context.Context, id string) (*Model, bool, error)
}

type Deleter[Model any] interface {
  Delete(ctx context.Context, id string) (error)
}

type Creator[Model any] interface {
  Create(ctx context.Context, m Model) (*Model, error)
}

type Store[Model any] interface {
  Retriever[Model]
  Creator[Model]
  Deleter[Model]
}

type Storable interface{
  schema.Tabler
  GetID() string
}

// Create serializes a Model into the database. Returns the model after it's
// written, in case the model pushes logic into the database.
func (s DBStore[Model, DatabaseModel]) Create(c context.Context, m Model) (*Model, error) {
	db := s.db.WithContext(c)
  fmt.Printf("storemodel --->\n\n%#v", s.model)
	serializable := s.model.Serialize(m)
  fmt.Printf("model --->\n\n%#v", m)
  fmt.Printf("serializable--->\n\n%#v", serializable)
	result := db.Create(&serializable)

  fmt.Printf("result--->\n\n%#v", result)

	if result.Error != nil {
		return nil, errors.Wrap(result.Error, "failed to create record")
	}

	// Re-fetch in case there are calculated fields.
	//retrieved, found, err := s.Retrieve(c, s.dbModel.GetID())
	//if err != nil {
		//return nil, errors.Wrap(err, "failed to retrieve newly-created model")
	//} else if !found {
		//return nil, errors.New("failed to find newly-created model")
	//}

	return &m, nil
}

// Retrieve a model.
func(s *DBStore[Model, DatabaseModel]) Retrieve(c context.Context, id string) (*Model, bool, error) {
	db := s.db.WithContext(c)
	query := db.Unscoped().Preload(clause.Associations)

  var d DatabaseModel
  query = query.Model(d)

	resp := query.First(
    &d,
    clause.Table{Name: s.dbModel.TableName()},
    clause.Where{Exprs: []clause.Expression{
      clause.Eq{Column: "id", Value: id},
    }},
  )

	if resp.Error != nil {
		if errors.Is(resp.Error, gorm.ErrRecordNotFound) {
			return nil, false, nil
		}
		return nil, false, errors.Wrap(resp.Error, "failed to retrieve model")
	}

  model, err := s.model.Deserialize(d)
  if err != nil {
    return nil, false, errors.Wrap(err, "deserialization failed")
  }

	return model, true, nil
}

type DBStore[Model any, DatabaseModel Storable] struct {
  db *gorm.DB
  dbModel DatabaseModel
  model Serder[Model, DatabaseModel]
}

func NewDBStore[DatabaseModel Storable, Model Serder[Model, DatabaseModel]] (
  db *gorm.DB,
  dbModel DatabaseModel,
  model Serder[Model, DatabaseModel],
) DBStore[Model, DatabaseModel] {
  return DBStore[Model, DatabaseModel]{
    db: db,
    dbModel: dbModel,
    model: model,
  }
}
