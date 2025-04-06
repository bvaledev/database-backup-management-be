package main

import (
	"log"

	"github.com/bvaledev/go-database-backaup-management/internal/backup"
	"github.com/bvaledev/go-database-backaup-management/internal/datasource"
)

func main() {
	postgresBackupService := backup.NewPostgresBackupService()
	datasource := &datasource.Datasource{
		Host:     "localhost",
		Database: "teste-restore",
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
