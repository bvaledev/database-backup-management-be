package job

import (
	"fmt"
	"log"
	"time"

	"github.com/bvaledev/go-database-backaup-management/internal/backup"
	"github.com/bvaledev/go-database-backaup-management/internal/datasource"
)

type PostgresBackupJobCommand struct {
	backupService backup.DBBackupService
}

var _ IJobCMD = (*PostgresBackupJobCommand)(nil)

func NewPostgresBackupJobCommand(backupService backup.DBBackupService) *PostgresBackupJobCommand {
	return &PostgresBackupJobCommand{backupService}
}

func (pgb *PostgresBackupJobCommand) Command(datasource datasource.Datasource) func() {
	return func() {
		log.Printf("[JOB COMMAND %s STARTED] Datasource: %s", datasource.ID, datasource.Database)
		fileName := fmt.Sprintf("%s-%d", datasource.Database, time.Now().Unix())
		decodedDataSource, err := datasource.Decoded()
		if err != nil {
			log.Printf("[JOB COMMAND ERROR] Datasource: %s, Error: %s", datasource.Database, err.Error())
			return
		}
		_, err = pgb.backupService.Backup(decodedDataSource, fileName, backup.Plain)
		if err != nil {
			log.Printf("[JOB COMMAND ERROR] Datasource: %s, Error: %s", datasource.Database, err.Error())
		}
		log.Printf("[JOB COMMAND %s FINISHED] Datasource: %s", datasource.ID, datasource.Database)
	}
}
