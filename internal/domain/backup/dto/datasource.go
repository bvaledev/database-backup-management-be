package dto

type CronExprDto struct {
	CronExpr    string `json:"cron_expr"`
	Description string `json:"description"`
	Enabled     bool   `json:"enabled"`
}

type CreateDatasourceDto struct {
	Host     string      `json:"host"`
	Database string      `json:"database"`
	Port     int32       `json:"port"`
	Username string      `json:"username"`
	Password string      `json:"password"`
	SSLMode  string      `json:"ssl_mode"`
	Cron     CronExprDto `json:"cron"`
}

type UpdateDatasourceDto struct {
	CreateDatasourceDto
}
