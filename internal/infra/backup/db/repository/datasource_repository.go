package repository

import (
	"database/sql"
	"log"

	"github.com/bvaledev/database-backup-management-be/internal/domain/backup/contract"
	"github.com/bvaledev/database-backup-management-be/internal/domain/backup/entity"
)

type DatasourceRepository struct {
	db *sql.DB
}

var _ contract.IDatasourceRepository = (*DatasourceRepository)(nil)

func NewDatasourceRepository(db *sql.DB) *DatasourceRepository {
	return &DatasourceRepository{db}
}

// GetDatasource implements IDatasourceRepository.
func (repo *DatasourceRepository) GetDatasource(entityID string) (entity.Datasource, error) {
	var datasource entity.Datasource = entity.Datasource{Cron: &entity.CronExpr{}}

	row := repo.db.QueryRow(`
		SELECT id, host, database, port, username, password, ssl_mode, cron_expr, description, enabled
		FROM datasources
		WHERE id = $1::uuid
	`, entityID)

	err := row.Scan(
		&datasource.ID,
		&datasource.Host,
		&datasource.Database,
		&datasource.Port,
		&datasource.Username,
		&datasource.Password,
		&datasource.SSLMode,
		&datasource.Cron.CronExpr,
		&datasource.Cron.Description,
		&datasource.Cron.Enabled,
	)
	if err != nil {
		return entity.Datasource{}, err
	}

	return datasource, nil
}

// GetDatasource implements IDatasourceRepository.
func (repo *DatasourceRepository) GetDatasources(enabled *bool) ([]entity.Datasource, error) {
	var (
		rows *sql.Rows
		err  error
	)

	if enabled == nil {
		rows, err = repo.db.Query(`
			SELECT id, host, database, port, username, password, ssl_mode, cron_expr, description, enabled
			FROM datasources
		`)
	} else {
		rows, err = repo.db.Query(`
			SELECT id, host, database, port, username, password, ssl_mode, cron_expr, description, enabled
			FROM datasources
			WHERE enabled = true
		`)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var datasources []entity.Datasource = make([]entity.Datasource, 0)
	for rows.Next() {
		var datasource entity.Datasource = entity.Datasource{Cron: &entity.CronExpr{}}
		err := rows.Scan(
			&datasource.ID,
			&datasource.Host,
			&datasource.Database,
			&datasource.Port,
			&datasource.Username,
			&datasource.Password,
			&datasource.SSLMode,
			&datasource.Cron.CronExpr,
			&datasource.Cron.Description,
			&datasource.Cron.Enabled,
		)
		if err != nil {
			return []entity.Datasource{}, err
		}
		datasources = append(datasources, datasource)
	}

	return datasources, nil
}

// CreateDatasource implements IDatasourceRepository.
func (repo *DatasourceRepository) CreateDatasource(entity entity.Datasource) error {
	stmt, err := repo.db.Prepare(`
		INSERT INTO datasources (id, host, database, port, username, password, ssl_mode, cron_expr, description, enabled)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`)
	if err != nil {
		return err
	}

	datasource, err := entity.Encode()
	if err != nil {
		return err
	}
	_, err = stmt.Exec(
		datasource.ID,
		datasource.Host,
		datasource.Database,
		datasource.Port,
		datasource.Username,
		datasource.Password,
		datasource.SSLMode,
		datasource.Cron.CronExpr,
		datasource.Cron.Description,
		datasource.Cron.Enabled,
	)
	if err != nil {
		return err
	}
	return nil
}

// UpdateDatasource implements IDatasourceRepository.
func (repo *DatasourceRepository) UpdateDatasource(entity entity.Datasource) error {
	datasource, err := entity.Encode()
	if err != nil {
		return err
	}

	log.Printf("Updating -> %+v", datasource)

	stmt, err := repo.db.Prepare(`
		UPDATE datasources
		SET host=$2, database=$3, port=$4, username=$5, password=$6, ssl_mode=$7, cron_expr=$8, description=$9, enabled=$10
		WHERE id = $1::uuid
	`)
	if err != nil {
		return err
	}
	_, err = stmt.Exec(
		datasource.ID,
		datasource.Host,
		datasource.Database,
		datasource.Port,
		datasource.Username,
		datasource.Password,
		datasource.SSLMode,
		datasource.Cron.CronExpr,
		datasource.Cron.Description,
		datasource.Cron.Enabled,
	)
	if err != nil {
		return err
	}
	return nil
}

// DeleteDatasource implements IDatasourceRepository.
func (repo *DatasourceRepository) DeleteDatasource(entityID string) error {
	_, err := repo.db.Exec(`
		DELETE FROM datasources
		WHERE id = $1::uuid
	`, entityID)

	return err
}
