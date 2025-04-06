package main

import (
	"context"
	"fmt"
	"log"
	netHttp "net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bvaledev/database-backup-management-be/internal/application/backup"
	"github.com/bvaledev/database-backup-management-be/internal/infra/backup/db"
	"github.com/bvaledev/database-backup-management-be/internal/infra/backup/db/repository"
	"github.com/bvaledev/database-backup-management-be/internal/infra/backup/handler/http"

	"github.com/bvaledev/database-backup-management-be/internal/pkg/encryption"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
)

func init() {
	log.Println("Carregando variáveis de ambiente do .env")
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Erro ao carregar .env")
	}
	if err := encryption.InitEncryptionKey(); err != nil {
		log.Fatal("Erro ao inicializar chave de criptografia:", err)
	}

	if _, err := os.Stat("./backups"); os.IsNotExist(err) {
		if err := os.Mkdir("./backups", 0755); err != nil {
			log.Fatalf("Erro ao criar o diretório de backups: %s", err)
		}
	}
}

func main() {
	dbConn, err := db.NewConnection()
	if err != nil {
		log.Fatal("Erro na configuração do banco de dados")
	}

	backupRepo := repository.NewBackupRepository(dbConn.DB)
	datasourceRepo := repository.NewDatasourceRepository(dbConn.DB)

	postgresBackupService := backup.NewPostgresBackupService()
	PostgresBackupCommand := backup.NewPostgresBackupCommand(postgresBackupService, backupRepo)

	backupController := http.NewBackupController(backupRepo, datasourceRepo, postgresBackupService, PostgresBackupCommand)
	datasourceController := http.NewDatasourceController(datasourceRepo)

	jobManager := backup.NewJobManager(datasourceRepo, PostgresBackupCommand)
	jobManager.Start()
	defer jobManager.Stop()

	appPort := os.Getenv("PORT")
	server := &netHttp.Server{Addr: fmt.Sprintf("0.0.0.0:%s", appPort), Handler: appRouters(datasourceController, backupController)}
	serverCtx, serverStopCtx := context.WithCancel(context.Background())

	// Listen for syscall signals for process to interrupt/quit
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-sig

		// Shutdown signal with grace period of 30 seconds
		shutdownCtx, cancelShutdownCtx := context.WithTimeout(serverCtx, 30*time.Second)
		defer cancelShutdownCtx()

		go func() {
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				log.Fatal("graceful shutdown timed out.. forcing exit.")
			}
		}()

		// Trigger graceful shutdown
		err := server.Shutdown(shutdownCtx)
		if err != nil {
			log.Fatal(err)
		}
		serverStopCtx()
	}()

	err = server.ListenAndServe()
	if err != nil && err != netHttp.ErrServerClosed {
		log.Fatal(err)
	}

	<-serverCtx.Done()
}

func appRouters(dsc *http.DatasourceController, bkp *http.BackupsController) netHttp.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Compress(5))
	r.Use(middleware.AllowContentEncoding("deflate", "gzip"))
	r.Use(middleware.CleanPath)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"https://*", "http://*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
	}))

	r.Get("/v1/datasources", dsc.List)
	r.Get("/v1/datasources/{id}", dsc.Get)
	r.Post("/v1/datasources", dsc.Create)
	r.Put("/v1/datasources/{id}", dsc.Update)
	r.Delete("/v1/datasources/{id}", dsc.Delete)

	r.Get("/v1/backups", bkp.List)
	r.Get("/v1/backups/{id}", bkp.Get)
	r.Post("/v1/backups", bkp.CreateBackup)
	r.Post("/v1/backups/{id}/restore-backup", bkp.RestoreBackup)
	r.Delete("/v1/backups/{id}", bkp.Delete)

	return r
}
