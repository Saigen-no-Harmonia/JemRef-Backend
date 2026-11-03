package infrastructure

import (
	"database/sql"
	"fmt"
	"jemref_go/internal/config"
	"strings"
)

// DsnParams DB接続時に必ず指定するDSNパラメータ
//   - loc=UTC       : Goのtime.TimeをUTCとして送受信する(ドライバのデフォルトだが明示する)
//   - time_zone     : 接続ごとに SET time_zone='+00:00' が発行され、MySQL側の時刻関数もUTCになる
//     (MySQLサーバのsystem_time_zoneがJST等であっても接続側で上書きできる)
//
// テスト(internal/infrastructure/db)も同じパラメータで接続し、本番と挙動を揃える。
var DsnParams = []string{
	"parseTime=true",
	"loc=UTC",
	"time_zone=%27%2B00%3A00%27", // '+00:00'
}

// NewDB DB接続
func NewDB(cfg config.Config) (*sql.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?%s",
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBName,
		strings.Join(DsnParams, "&"),
	)

	return sql.Open("mysql", dsn)
}
