package backup

import "github.com/bvaledev/go-database-backaup-management/internal/datasource"

type Mode string

const (
	Custom Mode = "c"
	Plain  Mode = "p"
)

// DBBackupService define as operações essenciais de backup e restauração para bancos de dados.
// Implementações podem suportar diferentes motores como PostgreSQL, MySQL, etc.
type DBBackupService interface {
	// TestConnection testa a conectividade com o banco de dados informado.
	//
	// Parâmetros:
	// - ds: informações de conexão com o banco.
	//
	// Retorna:
	// - Um erro caso a conexão falhe, ou nil se for bem-sucedida.
	TestConnection(ds datasource.Datasource) error

	// Backup realiza o backup completo do banco de dados no formato especificado.
	//
	// O backup é sempre gerado no diretório "./backups" e compactado com Gzip.
	//
	// Parâmetros:
	// - ds: informações de conexão com o banco.
	// - outputFile: nome base do arquivo de destino (sem extensão).
	// - format: modo de saída (Plain ou Custom). A extensão final será:
	//     - Plain  → .sql.gz
	//     - Custom → .backup.gz
	//
	// Retorna:
	// - A saída do comando pg_dump.
	// - Um erro, caso a execução falhe ou a compactação não seja concluída.
	Backup(ds datasource.Datasource, outputFile string, format Mode) (string, error)

	// Restore realiza a restauração do banco a partir de um arquivo de backup compactado ou não.
	//
	// O tipo de arquivo é detectado automaticamente com base na extensão:
	// - .sql           → executa psql
	// - .sql.gz        → descompacta e executa psql
	// - .backup        → executa pg_restore
	// - .backup.gz     → descompacta e executa pg_restore
	//
	// Antes da restauração, o banco é limpo com ClearDatabase.
	//
	// Retorna:
	// - A saída do comando de restauração.
	// - Um erro, caso o processo falhe.
	Restore(ds datasource.Datasource, inputFile string) (string, error)

	// ClearDatabase remove todos os schemas do banco, exceto os padrões, e recria o schema "public".
	ClearDatabase(ds datasource.Datasource) error

	// CreateDatabase cria um novo banco de dados com o nome informado.
	CreateDatabase(ds datasource.Datasource) error

	// DropDatabase remove um banco de dados existente com o nome informado.
	DropDatabase(ds datasource.Datasource) error
}
