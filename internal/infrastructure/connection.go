package infrastructure

import (
	"database/sql"
	"fmt"
	"jemref_go/internal/config"
)

// NewDB DB接続
func NewDB(cfg config.Config) (*sql.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBName,
	)

	return sql.Open("mysql", dsn)
}
