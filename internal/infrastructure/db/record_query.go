package db

// 書誌情報(e_records系テーブル)のSQL定義
//
// 監査列は PG にAPI ID、ID に 'system' を設定する(API IDはcontextから取得する)。
// 一括INSERTは固定部分(Base)と1行分のプレースホルダ(Values)に分け、
// buildBulkInsertQuery で行数分に展開する。

const insertRecordEntryQuery = `
		INSERT INTO e_records (
			ID,
			USER_ID,
			TYPE_ID,
            MAIN_TITLE,
            SUB_TITLE,
            PAGE_RANGE,
            VOLUME,
            ISSUE,
            PUBLISHER,
            PUBLICATION_DATE,
            DATE_PRECISION,
            READ_STATUS,
            CONTAINER_ID,
            M_RECORD_ID,
            DEL_FLG,
			INS_PG,
			INS_ID,
			UPD_PG,
			UPD_ID
		) VALUES (
			?,
			?,
			?,
			?,
			?,
			?,
			?,
			?,
			?,
			?,
			?,
			?,
			?,
			?,
			?,
			?,
			'system',
			?,
			'system'
		)
	`

const insertContributorQuery = `
		INSERT INTO e_contributors (
			RAW_NAME,
			USER_ID,
            DEL_FLG,
			INS_PG,
			INS_ID,
			UPD_PG,
			UPD_ID
		) VALUES (
			?,
			?,
			?,
			?,
			'system',
			?,
			'system'
		)
	`

const insertContributionQuery = `
		INSERT INTO e_record_contributions (
			E_RECORD_ID,
			E_CONTRIBUTOR_ID,
            ROLE_ID,
			DISPLAY_NUMBER,
            DEL_FLG,
			INS_PG,
			INS_ID,
			UPD_PG,
			UPD_ID
		) VALUES (
			?,
			?,
			?,
			?,
			?,
			?,
			'system',
			?,
			'system'
		)
	`

const insertRecordUrlsQueryBase = `
        INSERT INTO e_record_urls (
            E_RECORD_ID,
            DISPLAY_NUMBER,
            RAW_LABEL,
            RAW_URL,
            USER_ID,
            DEL_FLG,
			INS_PG,
			INS_ID,
			UPD_PG,
			UPD_ID
        ) VALUES
    `

const insertRecordUrlsQueryValues = "(?, ?, ?, ?, ?, ?,?,'system',?,'system')"

const insertRecordMemosQueryBase = `
        INSERT INTO e_record_memos (
            E_RECORD_ID,
            DISPLAY_NUMBER,
            TITLE,
            CONTENT,
            USER_ID,
            DEL_FLG,
			INS_PG,
			INS_ID,
			UPD_PG,
			UPD_ID
        ) VALUES
    `

const insertRecordMemosQueryValues = "(?, ?, ?, ?, ?, ?,?,'system',?,'system')"

const insertRecordFieldValuesQueryBase = `
        INSERT INTO e_record_field_values (
            E_RECORD_ID,
            RECORD_FIELD_ID,
            DISPLAY_NUMBER,
            RAW_VALUE,
            USER_ID,
            DEL_FLG,
			INS_PG,
			INS_ID,
			UPD_PG,
			UPD_ID
        ) VALUES
    `

const insertRecordFieldValuesQueryValues = "(?, ?, ?, ?, ?, ?,?,'system',?,'system')"

// selectRecordByPkQuery 書誌を1件取得する
//
// 所有権確認を兼ねるため内部用ユーザIDを条件に含め、論理削除済みは対象外とする。
// 詳細参照・削除・更新で共用するため全カラムを取得する(監査列は参照しないので対象外)。
const selectRecordByPkQuery = `
	select
		ID,
		USER_ID,
		TYPE_ID,
		MAIN_TITLE,
		SUB_TITLE,
		PAGE_RANGE,
		VOLUME,
		ISSUE,
		PUBLISHER,
		PUBLICATION_DATE,
		DATE_PRECISION,
		READ_STATUS,
		CONTAINER_ID,
		M_RECORD_ID,
		DEL_FLG
	from
		e_records
	where
		ID = ?
	and
		USER_ID = ?
	and
		DEL_FLG = 0
`

// countActiveArticlesQuery 指定した書誌をコンテナとする、論理削除されていない書誌の件数
const countActiveArticlesQuery = `
	select
		COUNT(*)
	from
		e_records
	where
		CONTAINER_ID = ?
	and
		DEL_FLG = 0
`

// 論理削除クエリ
//
// DEL_FLG = 0 を条件に含めるので、すでに削除済みの行は処理しない。
const logicalDeleteRecordFieldValuesQuery = `
	update e_record_field_values
	set
		DEL_FLG = 1,
		UPD_PG = ?,
		UPD_ID = 'system',
		UPD_DT = UTC_TIMESTAMP(6)
	where
		E_RECORD_ID = ?
	and
		DEL_FLG = 0
`

const logicalDeleteRecordMemosQuery = `
	update e_record_memos
	set
		DEL_FLG = 1,
		UPD_PG = ?,
		UPD_ID = 'system',
		UPD_DT = UTC_TIMESTAMP(6)
	where
		E_RECORD_ID = ?
	and
		DEL_FLG = 0
`

const logicalDeleteRecordUrlsQuery = `
	update e_record_urls
	set
		DEL_FLG = 1,
		UPD_PG = ?,
		UPD_ID = 'system',
		UPD_DT = UTC_TIMESTAMP(6)
	where
		E_RECORD_ID = ?
	and
		DEL_FLG = 0
`

const logicalDeleteRecordContributionsQuery = `
	update e_record_contributions
	set
		DEL_FLG = 1,
		UPD_PG = ?,
		UPD_ID = 'system',
		UPD_DT = UTC_TIMESTAMP(6)
	where
		E_RECORD_ID = ?
	and
		DEL_FLG = 0
`

const logicalDeleteRecordEntryQuery = `
	update e_records
	set
		DEL_FLG = 1,
		UPD_PG = ?,
		UPD_ID = 'system',
		UPD_DT = UTC_TIMESTAMP(6)
	where
		ID = ?
	and
		DEL_FLG = 0
`

// 構成要素の取得クエリ
//
// 対象書誌と収録物の最大2件をまとめて引くため、書誌IDはIN句で渡す(N+1回避)。
// 固定部分(Base)と末尾(Order)に分け、間のプレースホルダを buildPlaceholders で展開する。
// いずれも論理削除済みを除外し、表示順の昇順で返す。
// 書誌IDごとの振り分けに使うため、先頭カラムは必ず E_RECORD_ID とする。

const selectContributionsByRecordIdsQueryBase = `
	select
		c.E_RECORD_ID,
		c.ID,
		c.ROLE_ID,
		c.DISPLAY_NUMBER,
		co.ID,
		co.RAW_NAME,
		co.USER_ID,
		co.M_CONTRIBUTOR_ID
	from
		e_record_contributions c
	inner join
		e_contributors co
	on
		co.ID = c.E_CONTRIBUTOR_ID
	and
		co.DEL_FLG = 0
	where
		c.DEL_FLG = 0
	and
		c.E_RECORD_ID in (`

const selectContributionsByRecordIdsQueryOrder = `)
	order by
		c.E_RECORD_ID,
		c.DISPLAY_NUMBER
`

const selectUrlsByRecordIdsQueryBase = `
	select
		E_RECORD_ID,
		ID,
		DISPLAY_NUMBER,
		RAW_LABEL,
		RAW_URL,
		USER_ID
	from
		e_record_urls
	where
		DEL_FLG = 0
	and
		E_RECORD_ID in (`

const selectUrlsByRecordIdsQueryOrder = `)
	order by
		E_RECORD_ID,
		DISPLAY_NUMBER
`

const selectMemosByRecordIdsQueryBase = `
	select
		E_RECORD_ID,
		ID,
		DISPLAY_NUMBER,
		TITLE,
		CONTENT,
		USER_ID,
		SHARE_LEVEL
	from
		e_record_memos
	where
		DEL_FLG = 0
	and
		E_RECORD_ID in (`

const selectMemosByRecordIdsQueryOrder = `)
	order by
		E_RECORD_ID,
		DISPLAY_NUMBER
`

const selectFieldValuesByRecordIdsQueryBase = `
	select
		E_RECORD_ID,
		ID,
		RECORD_FIELD_ID,
		DISPLAY_NUMBER,
		RAW_VALUE,
		USER_ID
	from
		e_record_field_values
	where
		DEL_FLG = 0
	and
		E_RECORD_ID in (`

const selectFieldValuesByRecordIdsQueryOrder = `)
	order by
		E_RECORD_ID,
		DISPLAY_NUMBER
`
