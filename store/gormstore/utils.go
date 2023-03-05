package gormstore

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type null = string

const Null null = "~~~null~~~"

func isSigil[T ~string](maybeSigil T) bool {
	return null(maybeSigil) == Null
}

func ColumnInIDs[T ~string](columnName string, ids *[]T) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if ids == nil {
			return db
		}

		var c []clause.Expression
		var IDs []interface{}
		for _, id := range *ids {
			if isSigil(id) {
				//whereClauses = append(whereClauses, clause.n)
				c = append(c, clause.Expr{
					SQL:  "? IS NULL",
					Vars: []interface{}{clause.Column{Name: columnName}},
				})
				continue
			}

			IDs = append(IDs, id)
		}

		if len(IDs) > 0 {
			c = append(c, clause.IN{Column: columnName, Values: IDs})
		}

		if len(c) > 0 {
			db = db.Clauses(clause.Where{Exprs: []clause.Expression{clause.Or(c...)}})
		}

		return db
	}
}
