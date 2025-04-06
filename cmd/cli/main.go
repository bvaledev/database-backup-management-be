package main

import (
	"log"
	"os"

	"github.com/bvaledev/database-backup-management-be/internal/application/backup"
	"github.com/bvaledev/database-backup-management-be/internal/domain/backup/contract"
	"github.com/bvaledev/database-backup-management-be/internal/domain/backup/entity"
	"github.com/bvaledev/database-backup-management-be/internal/infra/backup/db"
	"github.com/bvaledev/database-backup-management-be/internal/infra/backup/db/repository"
	"github.com/bvaledev/database-backup-management-be/internal/pkg/encryption"
	"github.com/joho/godotenv"
)

var postgresBackupService contract.IBackupService
var backupRepo contract.IBackupRepository

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

	backupRepo = repository.NewBackupRepository(dbConn.DB)
	postgresBackupService = backup.NewPostgresBackupService()

	createBackup()
}

func createBackup() {
	ds := &entity.Datasource{
		ID:       "6aed1767-af62-4601-bf6c-5db9f6e74104",
		Host:     "localhost",
		Database: "fincycle",
		Port:     5432,
		Username: "postgres",
		Password: "root",
		SSLMode:  "disable",
	}

	PostgresBackupCommand := backup.NewPostgresBackupCommand(postgresBackupService, backupRepo)

	backaupCommand := PostgresBackupCommand.Command(*ds, entity.BackupManual)

	backaupCommand()
}

func restore() {
	datasource := &entity.Datasource{
		ID:       "6aed1767-af62-4601-bf6c-5db9f6e74104",
		Host:     "localhost",
		Database: "fincycle",
		Port:     5432,
		Username: "postgres",
		Password: "root",
		SSLMode:  "disable",
	}
	output, err := postgresBackupService.Restore(*datasource, "./backups/defaultdb-1743828420.sql.gz")
	if err != nil {
		panic(err)
	}
	log.Println("Restore output:", output)
}
