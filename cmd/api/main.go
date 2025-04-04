package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bvaledev/go-database-backaup-management/internal/backup"
	"github.com/bvaledev/go-database-backaup-management/internal/datasource"
	"github.com/bvaledev/go-database-backaup-management/internal/db"
	"github.com/bvaledev/go-database-backaup-management/internal/job"
	"github.com/bvaledev/go-database-backaup-management/internal/pkg/encryption"
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
}

func main() {
	dbConn, err := db.NewConnection()
	if err != nil {
		log.Fatal("Erro na configuração do banco de dados")
	}

	datasourceRepo := datasource.NewDatasourceRepository(dbConn.DB)
	datasourceController := datasource.NewDatasourceController(datasourceRepo)

	postgresBackupService := backup.NewPostgresBackupService()
	postgresBackupJobCommand := job.NewPostgresBackupJobCommand(postgresBackupService)
	jobManager := job.NewJobManager(datasourceRepo, postgresBackupJobCommand)
	jobManager.Start()
	defer jobManager.Stop()

	appPort := os.Getenv("PORT")
	server := &http.Server{Addr: fmt.Sprintf("0.0.0.0:%s", appPort), Handler: appRouters(datasourceController)}
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
	if err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}

	<-serverCtx.Done()
}

func appRouters(c *datasource.DatasourceController) http.Handler {
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

	r.Get("/v1/datasources", c.List)
	r.Get("/v1/datasources/{id}", c.Get)
	r.Post("/v1/datasources", c.Create)
	r.Put("/v1/datasources/{id}", c.Update)
	r.Delete("/v1/datasources/{id}", c.Delete)

	return r
}
