package http

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/bvaledev/database-backup-management-be/internal/domain/backup/contract"
	"github.com/bvaledev/database-backup-management-be/internal/domain/backup/dto"
	"github.com/bvaledev/database-backup-management-be/internal/domain/backup/entity"
	"github.com/bvaledev/database-backup-management-be/internal/utils"
	"github.com/go-chi/chi"
)

type BackupsController struct {
	backupRepo     contract.IBackupRepository
	datasourceRepo contract.IDatasourceRepository
	backupService  contract.IBackupService
	backupCommand  contract.ICommand
}

func NewBackupController(backupRepo contract.IBackupRepository, datasourceRepo contract.IDatasourceRepository, backupService contract.IBackupService, backupCommand contract.ICommand) *BackupsController {
	return &BackupsController{backupRepo, datasourceRepo, backupService, backupCommand}
}

func (c *BackupsController) List(w http.ResponseWriter, r *http.Request) {
	backupId := r.URL.Query().Get("datasourceId")
	var (
		backups []entity.Backup
		err     error
	)

	if backupId == "" {
		backups, err = c.backupRepo.GetBackups(nil)
	} else {
		backups, err = c.backupRepo.GetBackups(&backupId)
	}
	if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "não foi possível retornar os backups")
		return
	}

	utils.JSONResponse(w, http.StatusOK, backups)
}

func (c *BackupsController) Get(w http.ResponseWriter, r *http.Request) {
	backupId := chi.URLParam(r, "id")
	backup, err := c.backupRepo.GetBackup(backupId)
	if err != nil {
		utils.JSONError(w, http.StatusNotFound, "backup não encontrado")
		return
	}

	utils.JSONResponse(w, http.StatusOK, backup)
}

func (c *BackupsController) CreateBackup(w http.ResponseWriter, r *http.Request) {
	var input dto.CreateBackupDto
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.JSONError(w, http.StatusUnprocessableEntity, "json inválido")
		return
	}

	datasource, err := c.datasourceRepo.GetDatasource(input.DatasourceId)
	if err != nil {
		utils.JSONError(w, http.StatusNotFound, "datasource não encontrado")
		return
	}

	go c.backupCommand.Command(datasource, entity.BackupManual)()

	response := map[string]string{
		"datasource_id": datasource.ID,
		"status":        string(entity.BackupInitialized),
	}

	utils.JSONResponse(w, http.StatusOK, response)
}

func (c *BackupsController) RestoreBackup(w http.ResponseWriter, r *http.Request) {
	var (
		ds  entity.Datasource
		err error
	)

	backupId := chi.URLParam(r, "id")
	backup, err := c.backupRepo.GetBackup(backupId)
	if err != nil {
		utils.JSONError(w, http.StatusNotFound, "backup não encontrado")
		return
	}

	datasourceId := r.URL.Query().Get("datasourceId")
	if datasourceId == "" {
		ds, err = c.datasourceRepo.GetDatasource(backup.DatasourceId)
		if err != nil {
			utils.JSONError(w, http.StatusNotFound, "datasource não encontrado")
			return
		}
	} else {
		ds, err = c.datasourceRepo.GetDatasource(datasourceId)
		if err != nil {
			utils.JSONError(w, http.StatusNotFound, "datasource não encontrado")
			return
		}
	}

	go func(backupRepo contract.IBackupRepository, backup entity.Backup, ds entity.Datasource) {
		decodedDs, err := ds.Decode()
		if err != nil {
			log.Printf("erro ao decodificar datasource: %v", err)
			return
		}

		_, err = c.backupService.Restore(decodedDs, backup.FilePath)
		if err != nil {
			log.Printf("erro ao restaurar o backup: %v", err)
			return
		}
		backup.SetRestoredAt()
		backupRepo.UpdateBackup(backup)
	}(c.backupRepo, backup, ds)

	response := map[string]string{
		"message": "restauração iniciada",
	}

	utils.JSONResponse(w, http.StatusOK, response)
}

func (c *BackupsController) Delete(w http.ResponseWriter, r *http.Request) {
	backupId := chi.URLParam(r, "id")

	// Verifica se o backup existe
	backup, err := c.backupRepo.GetBackup(backupId)
	if err != nil {
		utils.JSONError(w, http.StatusNotFound, "o backup não existe")
		return
	}

	// verifica se o arquivo de backup existe
	if _, err := os.Stat(backup.FilePath); os.IsExist(err) {
		err := os.Remove(backup.FilePath)
		if err != nil {
			utils.JSONError(w, http.StatusNotFound, "não foi possível deletar o arquivo de backup")
			return
		}
	}

	err = c.backupRepo.DeleteBackup(backup.ID)
	if err != nil {
		utils.JSONError(w, http.StatusNotFound, "não foi possível deletar o backup")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
