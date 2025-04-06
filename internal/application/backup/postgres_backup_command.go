package backup

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/bvaledev/database-backup-management-be/internal/domain/backup/contract"
	"github.com/bvaledev/database-backup-management-be/internal/domain/backup/entity"
)

type PostgresBackupCommand struct {
	backupService contract.IBackupService
	backupRepo    contract.IBackupRepository
}

var _ contract.ICommand = (*PostgresBackupCommand)(nil)

func NewPostgresBackupCommand(backupService contract.IBackupService, backupRepo contract.IBackupRepository) *PostgresBackupCommand {
	return &PostgresBackupCommand{backupService, backupRepo}
}

func (pgb *PostgresBackupCommand) Command(ds entity.Datasource, trigger entity.BackupTrigger) func() {
	return func() {
		currenteBackup, err := pgb.onBackupInitialized(ds, trigger)
		if err != nil {
			log.Printf("[JOB ON BACKUP INITIALIZED ERROR] Datasource: %s, Error: %s", ds.Database, err.Error())
			return
		}

		log.Printf("[JOB COMMAND STARTED] Datasource: %s", ds.Database)
		decodedDataSource, err := ds.Decode()
		if err != nil {
			log.Printf("[JOB COMMAND ERROR] Datasource: %s, Error: %s", ds.Database, err.Error())
			if err := pgb.onBackupFailed(currenteBackup); err != nil {
				log.Printf("[JOB ON BACKUP FAILED ERROR] Datasource: %s, Error: %s", ds.Database, err.Error())
				return
			}
			return
		}

		fileName := fmt.Sprintf("./backups/%s-%d.sql.gz", ds.Database, time.Now().Unix())
		_, fileOutput, err := pgb.backupService.Backup(decodedDataSource, fileName, contract.Plain)
		if err != nil {
			log.Printf("[JOB COMMAND ERROR] Datasource: %s, Error: %s", ds.Database, err.Error())
			if err := pgb.onBackupFailed(currenteBackup); err != nil {
				log.Printf("[JOB ON BACKUP FAILED ERROR] Datasource: %s, Error: %s", ds.Database, err.Error())
				return
			}
		}

		if err := pgb.onBackupCompleted(currenteBackup, fileOutput); err != nil {
			log.Printf("[JOB ON BACKUP COMPLETED ERROR] Datasource: %s, Error: %s", ds.Database, err.Error())
			return
		}

		log.Printf("[JOB COMMAND FINISHED] Datasource: %s Backup File: %s", ds.Database, fileName)
	}
}

func (pgb *PostgresBackupCommand) onBackupInitialized(ds entity.Datasource, trigger entity.BackupTrigger) (*entity.Backup, error) {
	currenteBackup := entity.NewBackup(ds.ID, trigger)
	currenteBackup.SetStartedAt()
	if err := pgb.backupRepo.CreateBackup(*currenteBackup); err != nil {
		return &entity.Backup{}, err
	}
	return currenteBackup, nil
}

func (pgb *PostgresBackupCommand) onBackupFailed(currenteBackup *entity.Backup) error {
	currenteBackup.SetFailed()
	if currenteBackup.FinishedAt == nil {
		currenteBackup.SetFinishedAt()
	}
	if err := pgb.backupRepo.UpdateBackup(*currenteBackup); err != nil {
		return err
	}
	return nil
}

func (pgb *PostgresBackupCommand) onBackupCompleted(currenteBackup *entity.Backup, fileOutput string) error {
	fileInfo, err := os.Stat(fileOutput)
	if err != nil {
		if err := pgb.onBackupFailed(currenteBackup); err != nil {
			return err
		}
		return err
	}

	currenteBackup.SetCompleted()
	currenteBackup.FilePath = fileOutput
	currenteBackup.FileSize = fileInfo.Size()
	currenteBackup.FileOriginalName = filepath.Base(fileOutput)

	if currenteBackup.FinishedAt == nil {
		currenteBackup.SetFinishedAt()
	}

	if err := pgb.backupRepo.UpdateBackup(*currenteBackup); err != nil {
		return err
	}
	return nil
}
