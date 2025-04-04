package datasource

import (
	"github.com/bvaledev/go-database-backaup-management/internal/pkg/encryption"
	"github.com/google/uuid"
)

type CronExpr struct {
	CronExpr    string `json:"cron_expr"`
	Description string `json:"description"`
	Enabled     bool   `json:"enabled"`
}

type Datasource struct {
	ID        string    `json:"id"`
	Host      string    `json:"host"`
	Database  string    `json:"database"`
	Port      int32     `json:"port"`
	Username  string    `json:"username"`
	SSLMode   string    `json:"ssl_mode"`
	Password  string    `json:"-"`
	IsEncoded bool      `json:"-"`
	Cron      *CronExpr `json:"cron"`
}

func NewDatasource(host, database, username, password, sslMode string, port int32, cronExpr, description string, enabled bool) (*Datasource, error) {
	id := uuid.New()
	encodedPassword, err := encryption.Encrypt(password)
	if err != nil {
		return nil, err
	}
	return &Datasource{
		ID:        id.String(),
		Host:      host,
		Database:  database,
		Username:  username,
		Password:  encodedPassword,
		SSLMode:   sslMode,
		Port:      port,
		IsEncoded: true,
		Cron:      &CronExpr{cronExpr, description, enabled},
	}, nil
}

func (d Datasource) Encoded() (Datasource, error) {
	datasource := d
	if datasource.IsEncoded {
		return datasource, nil
	}
	encodedPassword, err := encryption.Encrypt(datasource.Password)
	if err != nil {
		return Datasource{}, err
	}
	datasource.Password = encodedPassword
	datasource.IsEncoded = true
	return datasource, nil
}

func (d Datasource) Decoded() (Datasource, error) {
	datasource := d
	if !datasource.IsEncoded {
		return datasource, nil
	}
	decodedPassword, err := encryption.Decrypt(d.Password)
	if err != nil {
		return Datasource{}, err
	}
	datasource.Password = decodedPassword
	datasource.IsEncoded = false
	return datasource, nil
}
