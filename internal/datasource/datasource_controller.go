package datasource

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/bvaledev/go-database-backaup-management/internal/utils"
	"github.com/go-chi/chi"
)

type DatasourceController struct {
	datasourceRepo IDatasourceRepository
}

func NewDatasourceController(datasourceRepo IDatasourceRepository) *DatasourceController {
	return &DatasourceController{datasourceRepo}
}

func (c *DatasourceController) List(w http.ResponseWriter, r *http.Request) {
	enabledStr := r.URL.Query().Get("enabled")
	var (
		datasources []Datasource
		err         error
	)

	if enabledStr == "" {
		datasources, err = c.datasourceRepo.GetDatasources(nil)
	} else {
		enabled, errConv := strconv.ParseBool(enabledStr)
		if errConv != nil {
			utils.JSONError(w, http.StatusBadRequest, "parametro 'enabled' inválido")
			return
		}
		datasources, err = c.datasourceRepo.GetDatasources(&enabled)
	}
	if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "não foi possível retornar os datasources")
		return
	}

	utils.JSONResponse(w, http.StatusOK, datasources)
}

func (c *DatasourceController) Get(w http.ResponseWriter, r *http.Request) {
	datasourceId := chi.URLParam(r, "id")
	datasource, err := c.datasourceRepo.GetDatasource(datasourceId)
	if err != nil {
		utils.JSONError(w, http.StatusNotFound, "datasource não encontrado")
		return
	}

	utils.JSONResponse(w, http.StatusOK, datasource)
}

func (c *DatasourceController) Create(w http.ResponseWriter, r *http.Request) {
	var input CreateDatasourceDto
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.JSONError(w, http.StatusUnprocessableEntity, "json inválido")
		return
	}
	datasource, err := NewDatasource(input.Host, input.Database, input.Username, input.Password, input.SSLMode, input.Port, input.Cron.CronExpr, input.Cron.Description, input.Cron.Enabled)
	if err != nil {
		utils.JSONError(w, http.StatusUnprocessableEntity, "datasource inválido")
		return
	}
	err = c.datasourceRepo.CreateDatasource(*datasource)
	if err != nil {
		utils.JSONError(w, http.StatusUnprocessableEntity, "não foi possivel cadastrar o datasource")
		return
	}
	response := map[string]string{
		"id": datasource.ID,
	}

	utils.JSONResponse(w, http.StatusCreated, response)
}

func (c *DatasourceController) Update(w http.ResponseWriter, r *http.Request) {
	datasourceId := chi.URLParam(r, "id")
	var input UpdateDatasourceDto
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.JSONError(w, http.StatusBadRequest, "json inválido")
		return
	}
	datasource, err := c.datasourceRepo.GetDatasource(datasourceId)
	if err != nil {
		utils.JSONError(w, http.StatusNotFound, "o datasource não existe")
		return
	}

	datasource.Host = input.Host
	datasource.Port = input.Port
	datasource.Database = input.Database
	datasource.Username = input.Username
	datasource.Password = input.Password
	datasource.SSLMode = input.SSLMode
	datasource.Cron.CronExpr = input.Cron.CronExpr
	datasource.Cron.Description = input.Cron.Description
	datasource.Cron.Enabled = input.Cron.Enabled

	err = c.datasourceRepo.UpdateDatasource(datasource)
	if err != nil {
		utils.JSONError(w, http.StatusUnprocessableEntity, "não foi possível atualizar o datasource")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (c *DatasourceController) Delete(w http.ResponseWriter, r *http.Request) {
	datasourceId := chi.URLParam(r, "id")
	err := c.datasourceRepo.DeleteDatasource(datasourceId)
	if err != nil {
		utils.JSONError(w, http.StatusNotFound, "o datasource não existe")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
