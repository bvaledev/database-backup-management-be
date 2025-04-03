package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/bvaledev/go-database-backaup-management/internal/backup"
	"github.com/bvaledev/go-database-backaup-management/internal/datasource"
)

func main() {
	//dsProd := &datasource.Datasource{
	//	ID:       "",
	//	Host:     "docbox-database-v1-do-user-12616404-0.b.db.ondigitalocean.com",
	//	Port:     25060,
	//	Database: "defaultdb",
	//	Username: "doadmin",
	//	Password: "AVNS_z3ReadFtXJ96GTlO6Kx",
	//	SSLMode:  "require",
	//}

	ds := &datasource.Datasource{
		ID:       "",
		Host:     "localhost",
		Port:     5432,
		Database: "teste-restore",
		Username: "postgres",
		Password: "root",
		SSLMode:  "disable",
	}

	service := backup.NewPostgresBackupService()

	//if _, err := service.Backup(*ds, "backup-teste-restore", backup.Gzip); err != nil {
	//	log.Fatalf("❌ Erro ao executar psql: %v", err)
	//}

	output, err := service.Restore(*ds, "./backups/backup-teste-restore.backup.gz")
	if err := saveOutputLog(*ds, output, "restore"); err != nil {
		log.Printf("⚠️ Falha ao salvar log: %v", err)
	}

	if err != nil {
		// Se tiver "ERROR" na saída, falha mesmo
		if strings.Contains(string(output), "ERROR") {
			log.Printf("falha crítica na restauração: %s\n%s", err.Error(), output)
			return
		}
		// Senão apenas alerta
		log.Printf("⚠️ Restauração finalizada com alertas: %s", err)
	}

	log.Println("✅ Backup feito")
	log.Println(output)

}

func saveOutputLog(ds datasource.Datasource, content string, prefix string) error {
	timestamp := time.Now().Format("20060102-150405")
	dir := "./backups"
	filename := fmt.Sprintf("%s/%s-%s-%s-log.txt", dir, prefix, ds.Database, timestamp)

	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("erro ao criar diretório %s: %w", dir, err)
	}

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("erro ao criar arquivo de log: %w", err)
	}
	defer file.Close()

	if _, err := file.WriteString(content); err != nil {
		return fmt.Errorf("erro ao escrever no arquivo de log: %w", err)
	}

	log.Printf("📁 Log salvo em %s", filename)
	return nil
}
