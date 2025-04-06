package backup

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/bvaledev/database-backup-management-be/internal/domain/backup/contract"
	"github.com/bvaledev/database-backup-management-be/internal/domain/backup/entity"
	"github.com/robfig/cron/v3"
)

type JobManager struct {
	cron           *cron.Cron
	jobs           map[string]cron.EntryID
	jobExprs       map[string]string
	jobLock        sync.Mutex
	datasourceRepo contract.IDatasourceRepository
	ctx            context.Context
	cancelCtx      context.CancelFunc
	jobCommand     contract.ICommand
}

func NewJobManager(datasourceRepo contract.IDatasourceRepository, jobCommand contract.ICommand) *JobManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &JobManager{
		cron:           cron.New(cron.WithSeconds()),
		jobs:           make(map[string]cron.EntryID),
		jobExprs:       make(map[string]string),
		datasourceRepo: datasourceRepo,
		ctx:            ctx,
		cancelCtx:      cancel,
		jobCommand:     jobCommand,
	}
}

func (jm *JobManager) Start() {
	log.Println("Scheduler iniciado.")

	jm.LoadJobsFromDB()
	jm.cron.Start()

	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				jm.LoadJobsFromDB()
			case <-jm.ctx.Done():
				return
			}
		}
	}()
}

func (jm *JobManager) Stop() {
	jm.cron.Stop()
	jm.cancelCtx()
}

func (jm *JobManager) LoadJobsFromDB() {
	jm.jobLock.Lock()
	defer jm.jobLock.Unlock()

	enabled := true
	datasources, err := jm.datasourceRepo.GetDatasources(&enabled)
	if err != nil {
		log.Printf("Erro ao carregar tarefas: %v", err)
		return
	}

	activeTasks := make(map[string]entity.Datasource)

	for _, ds := range datasources {
		activeTasks[ds.ID] = ds
		existingID, exists := jm.jobs[ds.ID]
		if exists {
			isCronChanged := jm.jobExprs[ds.ID] != ds.Cron.CronExpr
			if isCronChanged {
				jm.cron.Remove(existingID)
			} else {
				continue
			}
		}

		// Adiciona (ou re-adiciona) a tarefa
		entryID, err := jm.cron.AddFunc(ds.Cron.CronExpr, jm.jobCommand.Command(ds, entity.BackupCron))

		if err != nil {
			log.Printf("Erro ao adicionar tarefa ID %s: %v", ds.ID, err)
			continue
		}
		jm.jobs[ds.ID] = entryID
		jm.jobExprs[ds.ID] = ds.Cron.CronExpr
	}

	for id, entryID := range jm.jobs {
		if _, stillActive := activeTasks[id]; !stillActive {
			jm.cron.Remove(entryID)
			delete(jm.jobs, id)
			delete(jm.jobExprs, id)
		}
	}
}
