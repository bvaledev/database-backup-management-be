package contract

import "github.com/bvaledev/database-backup-management-be/internal/domain/backup/entity"

type IDatasourceRepository interface {
	GetDatasources(enabled *bool) ([]entity.Datasource, error)
	GetDatasource(entityID string) (entity.Datasource, error)
	CreateDatasource(entity entity.Datasource) error
	UpdateDatasource(entity entity.Datasource) error
	DeleteDatasource(entityID string) error
}
