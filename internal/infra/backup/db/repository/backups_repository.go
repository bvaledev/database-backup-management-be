package repository

import (
	"database/sql"

	"github.com/bvaledev/database-backup-management-be/internal/domain/backup/contract"
	"github.com/bvaledev/database-backup-management-be/internal/domain/backup/entity"
)

type BackupRepository struct {
	db *sql.DB
}

var _ contract.IBackupRepository = (*BackupRepository)(nil)

func NewBackupRepository(db *sql.DB) *BackupRepository {
	return &BackupRepository{db}
}

func (b *BackupRepository) GetBackup(entityID string) (entity.Backup, error) {
	backup := entity.Backup{}

	row := b.db.QueryRow(`
		SELECT id, datasource_id, trigger, status, file_path, file_original_name, file_size, started_at, finished_at, restored_at
		FROM backups
		WHERE id = $1::uuid
	`, entityID)
	err := row.Scan(
		&backup.ID,
		&backup.DatasourceId,
		&backup.Trigger,
		&backup.Status,
		&backup.FilePath,
		&backup.FileOriginalName,
		&backup.FileSize,
		&backup.StartedAt,
		&backup.FinishedAt,
		&backup.RestoredAt,
	)
	if err != nil {
		return entity.Backup{}, err
	}
	return backup, nil
}

func (b *BackupRepository) GetBackups(datasourceId *string) ([]entity.Backup, error) {
	var (
		rows *sql.Rows
		err  error
	)

	if datasourceId == nil {
		rows, err = b.db.Query(`
		SELECT id, datasource_id, trigger, status, file_path, file_original_name, file_size, started_at, finished_at, restored_at
		FROM backups
		ORDER BY finished_at DESC;
	`)
	} else {
		rows, err = b.db.Query(`
		SELECT id, datasource_id, trigger, status, file_path, file_original_name, file_size, started_at, finished_at, restored_at
		FROM backups
		WHERE datasource_id = $1::uuid
		ORDER BY finished_at DESC;
	`, *datasourceId,
		)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	backups := make([]entity.Backup, 0)
	for rows.Next() {
		var backup entity.Backup
		err := rows.Scan(
			&backup.ID,
			&backup.DatasourceId,
			&backup.Trigger,
			&backup.Status,
			&backup.FilePath,
			&backup.FileOriginalName,
			&backup.FileSize,
			&backup.StartedAt,
			&backup.FinishedAt,
			&backup.RestoredAt,
		)
		if err != nil {
			return nil, err
		}
		backups = append(backups, backup)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return backups, nil
}

func (b *BackupRepository) CreateBackup(entity entity.Backup) error {
	stmt, err := b.db.Prepare(`
		INSERT INTO backups (id, datasource_id, trigger, status, file_path, file_original_name, file_size, started_at, finished_at, restored_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`)
	if err != nil {
		return err
	}
	_, err = stmt.Exec(
		entity.ID,
		entity.DatasourceId,
		entity.Trigger,
		entity.Status,
		entity.FilePath,
		entity.FileOriginalName,
		entity.FileSize,
		entity.StartedAt,
		entity.FinishedAt,
		entity.RestoredAt,
	)
	if err != nil {
		return err
	}
	return nil
}

func (b *BackupRepository) UpdateBackup(entity entity.Backup) error {
	stmt, err := b.db.Prepare(`
		UPDATE backups
		SET trigger = $1, status = $2, file_path = $3, file_original_name = $4, file_size = $5, started_at = $6, finished_at = $7, restored_at = $8
		WHERE id = $9::uuid
	`)
	if err != nil {
		return err
	}
	_, err = stmt.Exec(
		entity.Trigger,
		entity.Status,
		entity.FilePath,
		entity.FileOriginalName,
		entity.FileSize,
		entity.StartedAt,
		entity.FinishedAt,
		entity.RestoredAt,
		entity.ID,
	)
	if err != nil {
		return err
	}
	return nil
}

func (b *BackupRepository) DeleteBackup(entityID string) error {
	_, err := b.db.Exec(`
		DELETE FROM backups
		WHERE id = $1::uuid
	`, entityID)
	return err
}
