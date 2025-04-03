package datasource

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type DatasourceController struct {
	datasourceRepo IDatasourceRepository
}

func NewDatasourceController(datasourceRepo IDatasourceRepository) *DatasourceController {
	return &DatasourceController{datasourceRepo}
}

func (c *DatasourceController) List(ctx *gin.Context) {
	enabledStr := ctx.Query("enabled")
	var (
		datasources []Datasource
		err         error
	)
	if enabledStr == "" {
		datasources, err = c.datasourceRepo.GetDatasources(nil)
	} else {
		enabled, errConv := strconv.ParseBool(enabledStr)
		if errConv != nil {
			ctx.JSON(400, gin.H{"error": "param 'enabled' inválido"})
			return
		}
		datasources, err = c.datasourceRepo.GetDatasources(&enabled)
	}
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"data": datasources,
	})
}

func (c *DatasourceController) Get(ctx *gin.Context) {
	datasourceId := ctx.Param("id")
	datasource, err := c.datasourceRepo.GetDatasource(datasourceId)
	if err != nil {
		ctx.JSON(404, gin.H{"error": "Datasource não existe", "details": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": datasource,
	})
}

func (c *DatasourceController) Create(ctx *gin.Context) {
	var input CreateDatasourceDto
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(400, gin.H{"error": "JSON inválido", "details": err.Error()})
		return
	}
	datasource, err := NewDatasource(input.Host, input.Database, input.Username, input.Password, input.SSLMode, input.Port, input.Cron.CronExpr, input.Cron.Description, input.Cron.Enabled)
	if err != nil {
		ctx.JSON(400, gin.H{"error": "Dados inválido", "details": err.Error()})
		return
	}
	err = c.datasourceRepo.CreateDatasource(*datasource)
	if err != nil {
		ctx.JSON(400, gin.H{"error": "Não foi possível criar o datasource", "details": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"id": datasource.ID,
		},
	})
}

func (c *DatasourceController) Update(ctx *gin.Context) {
	datasourceId := ctx.Param("id")
	var input UpdateDatasourceDto
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(400, gin.H{"error": "JSON inválido", "details": err.Error()})
		return
	}
	datasource, err := c.datasourceRepo.GetDatasource(datasourceId)
	if err != nil {
		ctx.JSON(404, gin.H{"error": "Datasource não existe", "details": err.Error()})
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
		ctx.JSON(400, gin.H{"error": "Schedule não foi salvo", "details": err.Error()})
		return
	}

	ctx.JSON(http.StatusNoContent, nil)
}

func (c *DatasourceController) Delete(ctx *gin.Context) {
	datasourceId := ctx.Param("id")
	err := c.datasourceRepo.DeleteDatasource(datasourceId)
	if err != nil {
		ctx.JSON(404, gin.H{"error": "Schedule não existe", "details": err.Error()})
		return
	}
	ctx.JSON(http.StatusNoContent, nil)
}
