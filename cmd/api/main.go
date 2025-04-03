package main

import (
	"log"

	"github.com/bvaledev/go-database-backaup-management/internal/backup"
	"github.com/bvaledev/go-database-backaup-management/internal/database"
	"github.com/bvaledev/go-database-backaup-management/internal/datasource"
	"github.com/bvaledev/go-database-backaup-management/internal/job"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Erro ao carregar .env")
	}
}

func main() {
	dbConn, err := database.NewConnection()
	if err != nil {
		log.Fatal("Erro na configuração do banco de dados")
	}

	datasourceRepo := datasource.NewDatasourceRepository(dbConn.DB)
	datasourceController := datasource.NewDatasourceController(datasourceRepo)

	postgresBackupService := backup.NewPostgresBackupService()
	postgresBackupJobCommand := job.NewPostgresBackupJobCommand(postgresBackupService)
	jobManager := job.NewJobManager(datasourceRepo, postgresBackupJobCommand)
	jobManager.Start()

	log.Println("Scheduler iniciado.")
	defer jobManager.Stop()

	router := gin.Default()
	router.SetTrustedProxies(nil)

	v1 := router.Group("/v1")

	datasourceRoute := v1.Group("/datasources")
	datasourceRoute.GET("/", datasourceController.List)
	datasourceRoute.GET("/:id", datasourceController.Get)
	datasourceRoute.POST("/", datasourceController.Create)
	datasourceRoute.PUT("/:id", datasourceController.Update)
	datasourceRoute.DELETE("/:id", datasourceController.Delete)

	router.Run()
}
