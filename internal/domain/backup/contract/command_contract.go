package contract

import "github.com/bvaledev/database-backup-management-be/internal/domain/backup/entity"

type ICommand interface {
	Command(ds entity.Datasource, trigger entity.BackupTrigger) func()
}
