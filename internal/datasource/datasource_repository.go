package datasource

import "database/sql"

type IDatasourceRepository interface {
	GetDatasources(enabled *bool) ([]Datasource, error)
	GetDatasource(entityID string) (Datasource, error)
	CreateDatasource(entity Datasource) error
	UpdateDatasource(entity Datasource) error
	DeleteDatasource(entityID string) error
}

type DatasourceRepository struct {
	db *sql.DB
}

var _ IDatasourceRepository = (*DatasourceRepository)(nil)

func NewDatasourceRepository(db *sql.DB) *DatasourceRepository {
	return &DatasourceRepository{db}
}

// GetDatasource implements IDatasourceRepository.
func (repo *DatasourceRepository) GetDatasource(entityID string) (Datasource, error) {
	var datasource Datasource = Datasource{IsEncoded: true, Cron: &CronExpr{}}

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
		return Datasource{}, err
	}

	return datasource, nil
}

// GetDatasource implements IDatasourceRepository.
func (repo *DatasourceRepository) GetDatasources(enabled *bool) ([]Datasource, error) {
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

	var datasources []Datasource = make([]Datasource, 0)
	for rows.Next() {
		var datasource Datasource = Datasource{IsEncoded: true, Cron: &CronExpr{}}
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
			return []Datasource{}, err
		}
		datasources = append(datasources, datasource)
	}

	return datasources, nil
}

// CreateDatasource implements IDatasourceRepository.
func (repo *DatasourceRepository) CreateDatasource(entity Datasource) error {
	stmt, err := repo.db.Prepare(`
		INSERT INTO datasources (id, host, database, port, username, password, ssl_mode, cron_expr, description, enabled)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`)
	if err != nil {
		return err
	}

	datasource, err := entity.Encoded()
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
func (repo *DatasourceRepository) UpdateDatasource(entity Datasource) error {
	stmt, err := repo.db.Prepare(`
		UPDATE datasources
		SET host=$2, database=$3, port=$4, username=$5, password=$6, ssl_mode=$7, cron_expr=$8, description=$9, enabled=$10
		WHERE id = $1::uuid
	`)
	if err != nil {
		return err
	}

	datasource, err := entity.Encoded()
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
