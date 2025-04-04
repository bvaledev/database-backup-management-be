package backup

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/bvaledev/go-database-backaup-management/internal/datasource"
	"github.com/bvaledev/go-database-backaup-management/internal/pkg/compression"
)

var (
	timeOutInMinutes = time.Duration(15)
)

type PostgresBackupService struct{}

var _ DBBackupService = (*PostgresBackupService)(nil)

func NewPostgresBackupService() *PostgresBackupService {
	service := &PostgresBackupService{}
	service.generateBackupDir()
	return service
}

// generateBackupDir garante que o diretório padrão de backups ("./backups") exista.
//
// Se o diretório ainda não existir, ele será criado com permissões 0755.
// Caso ocorra um erro ao criar o diretório, a aplicação será encerrada com log fatal.
//
// Este método é utilizado internamente antes de operações de backup para garantir que o local de destino esteja preparado.
//
// Não retorna valores. Em caso de erro, finaliza a execução com log.Fatalf.
func (pbs *PostgresBackupService) generateBackupDir() {
	if _, err := os.Stat("./backups"); os.IsNotExist(err) {
		if err := os.Mkdir("./backups", 0755); err != nil {
			log.Fatalf("Erro ao criar o diretório de backups: %s", err)
		}
	}
}

// TestConnection verifica a conectividade com um banco de dados PostgreSQL utilizando o comando psql.
//
// Parâmetros:
// - ds: informações de conexão com o banco de dados (host, porta, usuário, senha, banco, sslmode).
//
// Retorna:
// - Um erro, caso a conexão falhe.
// - nil, se a conexão for bem-sucedida.
func (pbs *PostgresBackupService) TestConnection(ds datasource.Datasource) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeOutInMinutes*time.Minute)
	defer cancel()

	cmd := pbs.buildCommand(
		ds,
		ctx,
		"psql",
		"-h", ds.Host,
		"-p", fmt.Sprintf("%d", ds.Port),
		"-U", ds.Username,
		"-d", ds.Database,
		"-v",
		fmt.Sprintf("sslmode=%s", ds.SSLMode),
	)
	_, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}

	return nil
}

// Backup realiza o backup de um banco de dados PostgreSQL utilizando o utilitário pg_dump.
//
// Esta função sempre gera backups compactados com Gzip, com a extensão final definida conforme o tipo:
// - Plain: gera um arquivo .sql.gz com os comandos SQL brutos.
// - Custom: gera um arquivo .backup.gz no formato customizado do PostgreSQL (ideal para pg_restore).
//
// O backup é salvo no diretório padrão "./backups" com a extensão apropriada.
//
// Parâmetros:
// - ds: informações de conexão com o banco de dados (host, porta, usuário, senha, sslmode, nome do banco).
// - outputFile: nome base do arquivo de backup (sem extensão).
// - format: tipo de formato interno (Plain ou Custom). O resultado será sempre compactado.
//
// Retorna:
// - A saída gerada pelo comando pg_dump (string), útil para logs e debugging.
// - Um erro, caso a execução do backup falhe ou a compactação não seja concluída com sucesso.
func (pbs *PostgresBackupService) Backup(ds datasource.Datasource, outputFile string, format Mode) (string, error) {
	pbs.generateBackupDir()

	ctx, cancel := context.WithTimeout(context.Background(), timeOutInMinutes*time.Minute)
	defer cancel()

	if err := pbs.TestConnection(ds); err != nil {
		return "", err
	}

	var dumpFormat Mode
	var tmpExt, finalExt string

	switch format {
	case Plain:
		dumpFormat = Plain
		tmpExt = ".sql"
		finalExt = ".sql.gz"
	case Custom:
		dumpFormat = Custom
		tmpExt = ".backup"
		finalExt = ".backup.gz"
	default:
		return "", fmt.Errorf("formato inválido de backup: %s", format)
	}

	tmpOutput := outputFile + tmpExt
	finalOutput := outputFile + finalExt

	if !strings.HasPrefix(tmpOutput, "./backups/") {
		tmpOutput = fmt.Sprintf("./backups/%s", tmpOutput)
	}
	if !strings.HasPrefix(finalOutput, "./backups/") {
		finalOutput = fmt.Sprintf("./backups/%s", finalOutput)
	}

	cmd := pbs.buildCommand(
		ds,
		ctx,
		"pg_dump",
		"--no-owner",
		"-h", ds.Host,
		"-p", fmt.Sprintf("%d", ds.Port),
		"-U", ds.Username,
		"-d", ds.Database,
		"-F", string(dumpFormat),
		"-v",
		"-f", tmpOutput,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("erro ao executar o backup: %s\n%s", err, string(output))
	}

	// Sempre gzip
	if err := compression.CompressToGzip(tmpOutput, finalOutput); err != nil {
		return string(output), fmt.Errorf("backup realizado, mas erro ao compactar: %w", err)
	}

	return string(output), nil
}

// ClearDatabase remove todos os schemas customizados de um banco de dados PostgreSQL,
// recriando apenas o schema público padrão.
//
// Parâmetros:
// - ds: informações de conexão com o banco de dados (host, porta, usuário, senha, banco, sslmode).
//
// Este método executa um bloco PL/pgSQL que:
// - Remove todos os schemas (exceto 'pg_catalog', 'information_schema', 'public', 'pg_toast%', 'pg_temp%').
// - Remove o schema 'public'.
// - Recria o schema 'public' limpo.
//
// ⚠️ Observação:
// Se o banco de dados contiver um número elevado de objetos (tabelas, views, funções, etc.),
// o PostgreSQL pode atingir o limite de locks por transação, gerando erro do tipo:
//
//	"ERROR: out of shared memory"
//	"HINT: You might need to increase max_locks_per_transaction"
//
// Para prevenir esse erro, você pode ajustar a configuração no servidor:
//
//	ALTER SYSTEM SET max_locks_per_transaction = 256; -- Altera o valor via SQL
//	sudo nano /etc/postgresql/<versão>/main/postgresql.conf -- Ou edite manualmente o arquivo
//	max_locks_per_transaction = 256 -- Linha de configuração a ser adicionada/ajustada
//	sudo systemctl restart postgresql -- Reinicie o serviço para aplicar as mudanças
//
// Retorna:
// - Um erro, caso a execução do comando SQL falhe.
// - nil, se o banco de dados for limpo com sucesso.
func (pbs *PostgresBackupService) ClearDatabase(ds datasource.Datasource) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeOutInMinutes*time.Minute)
	defer cancel()

	clearSQL := `
		DO $$ DECLARE
		    schema_name text;
		BEGIN
		    FOR schema_name IN
		        SELECT nspname FROM pg_namespace
		        WHERE nspname NOT IN ('pg_catalog', 'information_schema', 'public')
		        AND nspname NOT LIKE 'pg_toast%'
		        AND nspname NOT LIKE 'pg_temp%'
		    LOOP
		        EXECUTE format('DROP SCHEMA IF EXISTS %I CASCADE', schema_name);
		    END LOOP;
		END $$;

		DROP SCHEMA IF EXISTS public CASCADE;
		CREATE SCHEMA public;
	`

	cmd := pbs.buildCommand(
		ds,
		ctx,
		"psql",
		"-h", ds.Host,
		"-p", fmt.Sprintf("%d", ds.Port),
		"-U", ds.Username,
		"-d", ds.Database,
		"-c", clearSQL,
		fmt.Sprintf("sslmode=%s", ds.SSLMode),
	)

	log.Println("Limpando o banco de dados...")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("erro ao limpar o banco de dados: %s\n%s", err, string(output))
	}
	log.Println("Banco de dados limpo com sucesso.")
	return nil
}

// Restore restaura um banco de dados PostgreSQL a partir de um arquivo de backup nos formatos:
// .sql, .sql.gz, .backup ou .backup.gz.
//
// O tipo de restauração é detectado automaticamente com base na extensão do arquivo:
// - .sql           → executa o comando `psql` com o script SQL.
// - .sql.gz        → descompacta e executa `psql` com o script SQL.
// - .backup        → executa `pg_restore` com o formato custom do PostgreSQL.
// - .backup.gz     → descompacta e executa `pg_restore`.
//
// ⚠️ Somente arquivos com as extensões .sql.gz e .backup.gz são aceitos como válidos para restauração compactada.
//
// Antes da restauração, o banco de dados é limpo (todos os schemas são removidos, exceto os padrões).
func (pbs *PostgresBackupService) Restore(ds datasource.Datasource, inputFile string) (string, error) {
	if _, err := os.Stat(inputFile); os.IsNotExist(err) {
		return "", fmt.Errorf("arquivo de backup não encontrado: %s", inputFile)
	}

	// Detecta tipo de backup
	var usePgRestore bool
	var isGzipped bool

	switch {
	case strings.HasSuffix(inputFile, ".sql"):
		usePgRestore = false
		isGzipped = false
	case strings.HasSuffix(inputFile, ".sql.gz"):
		isGzipped = true
		usePgRestore = false
	case strings.HasSuffix(inputFile, ".backup"):
		usePgRestore = true
		isGzipped = false
	case strings.HasSuffix(inputFile, ".backup.gz"):
		isGzipped = true
		usePgRestore = true

	default:
		return "", fmt.Errorf("extensão do arquivo não reconhecida: %s", inputFile)
	}

	if err := pbs.ClearDatabase(ds); err != nil {
		return "", fmt.Errorf("falha ao limpar o banco de dados: %w", err)
	}

	originalInput := inputFile
	if isGzipped {
		tmp, err := compression.DecompressGzip(inputFile)
		if err != nil {
			return "", fmt.Errorf("erro ao descompactar %s: %w", inputFile, err)
		}
		inputFile = tmp
		defer os.Remove(tmp)
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeOutInMinutes*time.Minute)
	defer cancel()

	var cmd *exec.Cmd
	if usePgRestore {
		cmd = pbs.buildCommand(
			ds, ctx, "pg_restore",
			"--no-owner",
			"-h", ds.Host,
			"-p", fmt.Sprintf("%d", ds.Port),
			"-U", ds.Username,
			"-d", ds.Database,
			"-v",
			inputFile,
		)
	} else {
		cmd = pbs.buildCommand(
			ds, ctx, "psql",
			"-h", ds.Host,
			"-p", fmt.Sprintf("%d", ds.Port),
			"-U", ds.Username,
			"-d", ds.Database,
			fmt.Sprintf("sslmode=%s", ds.SSLMode),
			"-f", inputFile,
		)
	}
	log.Println("Restaurando banco de dados.")
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Se tiver "ERROR" na saída, falha mesmo
		if strings.Contains(string(output), "ERROR") {
			return string(output), fmt.Errorf("falha crítica na restauração (%s): %w\n%s", originalInput, err, output)
		}
		// Senão apenas alerta
		log.Printf("⚠️ Restauração finalizada com alertas: %s", err)
	}
	return string(output), nil
}

// CreateDatabase cria um novo banco de dados PostgreSQL utilizando o comando psql.
//
// Este método conecta-se ao servidor PostgreSQL e executa um comando SQL para criar o banco de dados informado.
// É necessário que o usuário tenha permissão para executar CREATE DATABASE.
//
// Parâmetros:
// - ds: informações de conexão com o banco de dados (host, porta, usuário, senha, nome do banco, sslmode).
//
// Retorna:
// - Um erro, caso o comando falhe (ex: banco já exista ou falta de permissão).
// - nil, se o banco for criado com sucesso.
func (pbs *PostgresBackupService) CreateDatabase(ds datasource.Datasource) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeOutInMinutes*time.Minute)
	defer cancel()

	cmd := pbs.buildCommand(
		ds,
		ctx,
		"psql",
		"-h", ds.Host,
		"-p", fmt.Sprintf("%d", ds.Port),
		"-U", ds.Username,
		"-c", fmt.Sprintf("CREATE DATABASE %s;", ds.Database),
		fmt.Sprintf("sslmode=%s", ds.SSLMode),
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("erro ao remover o banco de dados: %s\n%s", err, output)
	}
	return nil
}

// DropDatabase remove um banco de dados PostgreSQL utilizando o comando psql.
//
// Este método conecta-se ao servidor PostgreSQL e executa o comando DROP DATABASE para excluir o banco informado.
// É necessário que o usuário tenha permissão para executar DROP DATABASE.
// O banco não pode estar em uso por outras conexões no momento da exclusão.
//
// Parâmetros:
// - ds: informações de conexão com o banco de dados (host, porta, usuário, senha, nome do banco, sslmode).
//
// Retorna:
// - Um erro, caso o comando falhe (ex: banco inexistente, falta de permissão, ou conexões ativas).
// - nil, se o banco for removido com sucesso.
func (pbs *PostgresBackupService) DropDatabase(ds datasource.Datasource) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeOutInMinutes*time.Minute)
	defer cancel()

	cmd := pbs.buildCommand(
		ds,
		ctx,
		"psql",
		"-h", ds.Host,
		"-p", fmt.Sprintf("%d", ds.Port),
		"-U", ds.Username,
		"-c", fmt.Sprintf("DROP DATABASE %s;", ds.Database),
		fmt.Sprintf("sslmode=%s", ds.SSLMode),
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("erro ao criar o banco de dados: %s\n%s", err, output)
	}
	return nil
}

// buildCommand cria um comando executável (exec.Cmd) com contexto, argumentos e variáveis de ambiente específicas para PostgreSQL.
//
// Este método é utilizado internamente para montar comandos como psql, pg_dump ou pg_restore com os parâmetros corretos.
// Ele também injeta as variáveis de ambiente PGPASSWORD e PGSSLMODE para autenticação e configuração de segurança.
//
// Parâmetros:
// - ds: informações de conexão com o banco de dados (host, porta, usuário, senha, sslmode).
// - ctx: contexto que permite controle de timeout e cancelamento da execução.
// - executable: nome do comando a ser executado (ex: "psql", "pg_dump", "pg_restore").
// - args: argumentos adicionais para o comando.
//
// Retorna:
// - Um ponteiro para exec.Cmd pronto para execução com ambiente e contexto configurados.
func (*PostgresBackupService) buildCommand(ds datasource.Datasource, ctx context.Context, executable string, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, executable, args...)
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("PGPASSWORD=%s", ds.Password),
		fmt.Sprintf("PGSSLMODE=%s", ds.SSLMode),
	)
	return cmd
}
