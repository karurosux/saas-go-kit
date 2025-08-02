package {{.ModuleName}}gorm

import (
	"gorm.io/gorm"
	{{.ModuleName}}model "{{.Project.GoModule}}/internal/{{.ModuleName}}/model"
)

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&{{.ModuleName}}model.{{.ModuleNameCap}}{},
	)
}