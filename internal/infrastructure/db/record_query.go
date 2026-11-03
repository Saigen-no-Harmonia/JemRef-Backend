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
            PARENT_ID,
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
