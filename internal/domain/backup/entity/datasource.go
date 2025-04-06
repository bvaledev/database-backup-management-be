package entity

import (
	"github.com/bvaledev/database-backup-management-be/internal/pkg/encryption"
	"github.com/google/uuid"
)

type CronExpr struct {
	CronExpr    string `json:"cron_expr"`
	Description string `json:"description"`
	Enabled     bool   `json:"enabled"`
}

type Datasource struct {
	ID       string    `json:"id"`
	Host     string    `json:"host"`
	Database string    `json:"database"`
	Port     int32     `json:"port"`
	Username string    `json:"username"`
	SSLMode  string    `json:"ssl_mode"`
	Password string    `json:"-"`
	Cron     *CronExpr `json:"cron"`
}

func NewDatasource(host, database, username, password, sslMode string, port int32, cronExpr, description string, enabled bool) (*Datasource, error) {
	id := uuid.New()
	return &Datasource{
		ID:       id.String(),
		Host:     host,
		Database: database,
		Username: username,
		Password: password,
		SSLMode:  sslMode,
		Port:     port,
		Cron:     &CronExpr{cronExpr, description, enabled},
	}, nil
}

func (d *Datasource) Encode() (Datasource, error) {
	datasource := d
	if d.IsEncoded() {
		return *datasource, nil
	}
	encodedPassword, err := encryption.Encrypt(datasource.Password)
	if err != nil {
		return Datasource{}, err
	}
	datasource.Password = encodedPassword
	return *datasource, nil
}

func (d *Datasource) Decode() (Datasource, error) {
	datasource := d
	if !d.IsEncoded() {
		return *datasource, nil
	}
	decodedPassword, err := encryption.Decrypt(d.Password)
	if err != nil {
		return Datasource{}, err
	}
	datasource.Password = decodedPassword
	return *datasource, nil
}

func (d *Datasource) IsEncoded() bool {
	_, err := encryption.Decrypt(d.Password)
	return err == nil
}
