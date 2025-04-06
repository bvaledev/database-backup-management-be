package entity

import (
	"time"

	"github.com/google/uuid"
)

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
	ID               string        `json:"id"`
	DatasourceId     string        `json:"datasource_id"`
	Trigger          BackupTrigger `json:"trigger"`
	Status           BackupStatus  `json:"status"`
	FilePath         string        `json:"file_path"`
	FileOriginalName string        `json:"file_original_name"`
	FileSize         int64         `json:"file_size"`
	StartedAt        *time.Time    `json:"started_at"`
	FinishedAt       *time.Time    `json:"finished_at"`
	RestoredAt       *time.Time    `json:"restored_at"`
}

func NewBackup(datasourceId string, trigger BackupTrigger) *Backup {
	return &Backup{
		ID:               uuid.New().String(),
		DatasourceId:     datasourceId,
		Trigger:          trigger,
		Status:           BackupInitialized,
		FilePath:         "",
		FileOriginalName: "",
		FileSize:         0,
		StartedAt:        nil,
		FinishedAt:       nil,
		RestoredAt:       nil,
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
