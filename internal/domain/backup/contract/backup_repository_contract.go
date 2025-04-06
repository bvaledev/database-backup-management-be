package contract

import "github.com/bvaledev/database-backup-management-be/internal/domain/backup/entity"

type IBackupRepository interface {
	GetBackups(datasourceId *string) ([]entity.Backup, error)
	GetBackup(entityID string) (entity.Backup, error)
	CreateBackup(entity entity.Backup) error
	UpdateBackup(entity entity.Backup) error
	DeleteBackup(entityID string) error
}
