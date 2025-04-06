package datasource

import "time"

type BackupTrigger string
type BackupStatus string

var (
	BackupManual BackupTrigger = "manual"
	BackupCron   BackupTrigger = "cron"

	BackupInitialized BackupStatus = "initialized"
	BackupCompleted   BackupStatus = "completed"
	BackupFailed      BackupStatus = "failed"
)

type Backup struct {
	ID           string        `json:"id"`
	DatasourceId string        `json:"datasource_id"`
	Trigger      BackupTrigger `json:"trigger"`
	Status       BackupStatus  `json:"status"`
	FileName     string        `json:"file_name"`
	FileSize     string        `json:"file_size"`
	StartedAt    *time.Time    `json:"started_at"`
	FinishedAt   *time.Time    `json:"finished_at"`
	RestoredAt   *time.Time    `json:"restored_at"`
}

func NewBackups(id, datasourceId, fileName, fileSize string, trigger BackupTrigger, status BackupStatus) *Backup {
	return &Backup{
		ID:           id,
		DatasourceId: datasourceId,
		Trigger:      trigger,
		Status:       status,
		FileName:     fileName,
		FileSize:     fileSize,
		StartedAt:    nil,
		FinishedAt:   nil,
		RestoredAt:   nil,
	}
}

func (b *Backup) SetStartedAt() {
	now := time.Now()
	b.StartedAt = &now
}

func (b *Backup) SetFinishedAt() {
	now := time.Now()
	b.FinishedAt = &now
}

func (b *Backup) SetRestoredAt() {
	now := time.Now()
	b.RestoredAt = &now
}

func (b *Backup) SetCompleted() {
	b.Status = BackupCompleted
}

func (b *Backup) SetFailed() {
	b.Status = BackupFailed
}

func (b *Backup) SetInitialized() {
	b.Status = BackupInitialized
}

func (b *Backup) SetTriggerCron() {
	b.Trigger = BackupCron
}

func (b *Backup) SetTriggerManual() {
	b.Trigger = BackupManual
}
