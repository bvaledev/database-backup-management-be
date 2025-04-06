package datasource

import "database/sql"

type IBackupRepository interface {
	GetBackups(datasourceId *string) ([]Backup, error)
	GetBackup(entityID string) (Backup, error)
	CreateBackup(entity Backup) error
	UpdateBackup(entity Backup) error
	DeleteBackup(entityID string) error
}

type BackupRepository struct {
	db *sql.DB
}

var _ IBackupRepository = (*BackupRepository)(nil)

func NewBackupRepository(db *sql.DB) *BackupRepository {
	return &BackupRepository{db}
}

func (b *BackupRepository) GetBackup(entityID string) (Backup, error) {
	var backup Backup = Backup{}

	row := b.db.QueryRow(`
		SELECT id, datasource_id, trigger, status, file_name, file_size, started_at, finished_at, restored_at
		FROM backups
		WHERE id = $1::uuid
	`, entityID)
	err := row.Scan(
		&backup.ID,
		&backup.DatasourceId,
		&backup.Trigger,
		&backup.Status,
		&backup.FileName,
		&backup.FileSize,
		&backup.StartedAt,
		&backup.FinishedAt,
		&backup.RestoredAt,
	)
	if err != nil {
		return Backup{}, err
	}
	return backup, nil
}

func (b *BackupRepository) GetBackups(datasourceId *string) ([]Backup, error) {
	var (
		rows *sql.Rows
		err  error
	)

	if datasourceId == nil {
		rows, err = b.db.Query(`
		SELECT id, datasource_id, trigger, status, file_name, file_size, started_at, finished_at, restored_at
		FROM backups
	`)
	} else {
		rows, err = b.db.Query(`
		SELECT id, datasource_id, trigger, status, file_name, file_size, started_at, finished_at, restored_at
		FROM backups
		WHERE datasource_id = $1::uuid
	`, *datasourceId,
		)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var backups []Backup
	for rows.Next() {
		var backup Backup
		err := rows.Scan(
			&backup.ID,
			&backup.DatasourceId,
			&backup.Trigger,
			&backup.Status,
			&backup.FileName,
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

func (b *BackupRepository) CreateBackup(entity Backup) error {
	stmt, err := b.db.Prepare(`
		INSERT INTO backups (id, datasource_id, trigger, status, file_name, file_size, started_at, finished_at, restored_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`)
	if err != nil {
		return err
	}
	_, err = stmt.Exec(
		entity.ID,
		entity.DatasourceId,
		entity.Trigger,
		entity.Status,
		entity.FileName,
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

func (b *BackupRepository) UpdateBackup(entity Backup) error {
	stmt, err := b.db.Prepare(`
		UPDATE backups
		SET trigger = $1, status = $2, file_name = $3, file_size = $4, started_at = $5, finished_at = $6, restored_at = $7
		WHERE id = $8::uuid
	`)
	if err != nil {
		return err
	}
	_, err = stmt.Exec(
		entity.Trigger,
		entity.Status,
		entity.FileName,
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
