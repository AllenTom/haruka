package gormh

import "gorm.io/gorm"

type DataModel interface {
	DeleteById(id uint) error
	UpdateById(id uint, values map[string]interface{}) (interface{}, error)
}

// UpdateModels update models with update data mapping
func UpdateModels(db *gorm.DB, updateModel interface{}, updateModels []interface{}, allowFields ...string) error {
	var err error
	tx := db.Begin()
	for _, updateMapInterface := range updateModels {
		rawUpdateMap := updateMapInterface.(map[string]interface{})
		updateMap := make(map[string]interface{}, 0)
		for _, key := range allowFields {
			updateMap[key] = rawUpdateMap[key]
		}
		err := db.Model(updateModel).Where("id = ?", rawUpdateMap["id"]).Updates(updateMap).Error
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	tx.Commit()
	return err
}
