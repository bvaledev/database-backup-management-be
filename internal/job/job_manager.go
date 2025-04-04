package job

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/bvaledev/go-database-backaup-management/internal/datasource"
	"github.com/robfig/cron/v3"
)

type IJobCMD interface {
	Command(datasource datasource.Datasource) func()
}

type JobManager struct {
	cron           *cron.Cron
	jobs           map[string]cron.EntryID
	jobExprs       map[string]string
	jobLock        sync.Mutex
	datasourceRepo datasource.IDatasourceRepository
	ctx            context.Context
	cancelCtx      context.CancelFunc
	jobCommand     IJobCMD
}

func NewJobManager(datasourceRepo datasource.IDatasourceRepository, jobCommand IJobCMD) *JobManager {
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

	activeTasks := make(map[string]datasource.Datasource)

	for _, datasource := range datasources {
		activeTasks[datasource.ID] = datasource
		existingID, exists := jm.jobs[datasource.ID]
		if exists {
			isCronChanged := jm.jobExprs[datasource.ID] != datasource.Cron.CronExpr
			if isCronChanged {
				jm.cron.Remove(existingID)
			} else {
				continue
			}
		}

		// Adiciona (ou re-adiciona) a tarefa
		entryID, err := jm.cron.AddFunc(datasource.Cron.CronExpr, jm.jobCommand.Command(datasource))

		if err != nil {
			log.Printf("Erro ao adicionar tarefa ID %s: %v", datasource.ID, err)
			continue
		}
		jm.jobs[datasource.ID] = entryID
		jm.jobExprs[datasource.ID] = datasource.Cron.CronExpr
	}

	for id, entryID := range jm.jobs {
		if _, stillActive := activeTasks[id]; !stillActive {
			jm.cron.Remove(entryID)
			delete(jm.jobs, id)
			delete(jm.jobExprs, id)
		}
	}
}
