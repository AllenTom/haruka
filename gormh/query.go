package gormh

import (
	"gorm.io/gorm"
	"reflect"
)

type GORMFilter interface {
	ApplyQuery(db *gorm.DB) *gorm.DB
}

//apply query builder inner filters that implement GORMFilter
func ApplyFilters(queryBuilder interface{}, db *gorm.DB) *gorm.DB {
	query := db
	builderRef := reflect.ValueOf(queryBuilder).Elem()
	builderFieldCount := builderRef.NumField()
	//walk through all field of builder
	for fieldIdx := 0; fieldIdx < builderFieldCount; fieldIdx++ {
		field := builderRef.Field(fieldIdx).Interface()
		gormFilter, isImplement := field.(GORMFilter)
		if isImplement {
			query = gormFilter.ApplyQuery(query)
		}
	}
	return query
}

//Ids filter
type IdQueryFilter struct {
	Ids []interface{}
}

func (f IdQueryFilter) ApplyQuery(db *gorm.DB) *gorm.DB {
	if f.Ids != nil && len(f.Ids) != 0 {
		return db.Where("id in (?)", f.Ids)
	}
	return db
}

func (f *IdQueryFilter) InId(ids ...interface{}) {
	for _, id := range ids {
		if !isZeroVal(id) {
			f.Ids = append(f.Ids, id)
		}

	}

}

//order filter
type OrderQueryFilter struct {
	Order string
}

func (f OrderQueryFilter) ApplyQuery(db *gorm.DB) *gorm.DB {
	if len(f.Order) != 0 {
		return db.Order(f.Order)
	}
	return db
}

func (f *OrderQueryFilter) SetOrderFilter(order string) {
	f.Order = order
}

//read models from database
type ModelsReader interface {
	ReadModels() (int64, interface{}, error)
}
