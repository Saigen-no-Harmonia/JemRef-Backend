package db

// DATETIME型はタイムゾーンを持たないため、Goが渡す値とMySQLが生成する値の
// 壁時計がUTCに揃っていることを検証する。
// DSNの loc / time_zone 指定(infrastructure.DsnParams)が外れると、
// MySQLサーバのタイムゾーン設定次第で両者がずれるが、
// そのズレは「DBに入った値が不正」という形でしか現れずテストをすり抜ける。

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTimezone_接続のタイムゾーンがUTCに固定されている(t *testing.T) {
	var sessionTz string
	err := testDb.QueryRow("select @@session.time_zone").Scan(&sessionTz)
	require.NoError(t, err)

	// DSNの time_zone='+00:00' が接続ごとに適用されること
	// (MySQLサーバのsystem_time_zoneがJST等でもこれが優先される)
	assert.Equal(t, "+00:00", sessionTz)
}

func TestTimezone_Go由来とDB由来の日時の壁時計が一致する(t *testing.T) {
	ctx := context.Background()

	const table = "zz_timezone_check"
	_, err := testDb.ExecContext(ctx, "drop table if exists "+table)
	require.NoError(t, err)

	_, err = testDb.ExecContext(ctx, `create table `+table+` (
		GO_DT       DATETIME(6),
		SQL_UTC_DT  DATETIME(6),
		DEFAULT_DT  DATETIME(6) DEFAULT CURRENT_TIMESTAMP(6)
	)`)
	require.NoError(t, err)
	defer testDb.ExecContext(ctx, "drop table "+table)

	goNow := time.Now()
	_, err = testDb.ExecContext(ctx,
		"insert into "+table+" (GO_DT, SQL_UTC_DT) values (?, UTC_TIMESTAMP(6))", goNow)
	require.NoError(t, err)

	var goDt, sqlUtcDt, defaultDt time.Time
	err = testDb.QueryRowContext(ctx,
		"select GO_DT, SQL_UTC_DT, DEFAULT_DT from "+table).Scan(&goDt, &sqlUtcDt, &defaultDt)
	require.NoError(t, err)

	// 3者はいずれも同一時点を指すこと
	// (ずれるとしたら9時間等の単位になるため、1分の許容で十分に検出できる)
	assert.WithinDuration(t, goNow, goDt, time.Minute)
	assert.WithinDuration(t, goNow, sqlUtcDt, time.Minute)
	assert.WithinDuration(t, goNow, defaultDt, time.Minute)

	// 読み戻した値がUTCとして解釈されること(loc=UTC)
	assert.Equal(t, time.UTC, goDt.Location())
	assert.Equal(t, time.UTC, defaultDt.Location())
}
