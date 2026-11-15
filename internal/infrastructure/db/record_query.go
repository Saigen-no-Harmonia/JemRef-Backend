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

// recordEntryColumns 書誌(e_records)を復元するための取得カラム
//
// scanRecordEntry のScan順と一致させる。書誌を返すクエリはすべてこれを使い、列順を1か所に固定する。
// 監査列は参照しないので対象外。FROMに e_records を1つだけ置くクエリで使う前提で、テーブル名は付けない。
const recordEntryColumns = `
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
`

// selectRecordByPkQuery 書誌を1件取得する
//
// 所有権確認を兼ねるため内部用ユーザIDを条件に含め、論理削除済みは対象外とする。
// 詳細参照・削除・更新で共用するため全カラムを取得する(監査列は参照しないので対象外)。
const selectRecordByPkQuery = `
	select` + recordEntryColumns + `
	from  e_records
	where ID = ?
	  and USER_ID = ?
	  and DEL_FLG = 0
`

// selectRecordsByPksQueryBase 書誌を複数件取得する(一覧で表示する収録物の一括取得用)
//
// 条件は selectRecordByPkQuery と同じ(自ユーザ・論理削除済みは対象外)。
// プレースホルダは USER_ID が先で、IN句の書誌IDが続く。
const selectRecordsByPksQueryBase = `
	select` + recordEntryColumns + `
	from  e_records
	where USER_ID = ?
	  and DEL_FLG = 0
	  and ID in (`

const selectRecordsByPksQueryEnd = `)
`

// 書誌一覧の検索クエリ
//
// 件数クエリ(Count)と一覧クエリ(Select)は同じWHEREを共有する。
// WHEREは固定の条件(recordListWhereBase)に、指定された検索条件の断片を連結して組み立てる
// (buildRecordListWhere)。断片はいずれも外側の書誌を別名 r で参照し、先頭に and を持つ。
//
// ユーザの入力はすべてプレースホルダで渡す。SQLに連結してよいのは、ここで定義した断片と
// ORDER BYのホワイトリスト(recordListOrderBy)から引いた句だけ。
// LIKEに渡す値はエスケープ済みのパターンとし、エスケープ文字は '!' に固定する
// (既定の '\' は sql_mode の NO_BACKSLASH_ESCAPES で無効になるため明示する)。

const countRecordListQueryBase = `
	select
		COUNT(*)
	from
		e_records r
`

const selectRecordListQueryBase = `
	select` + recordEntryColumns + `
	from
		e_records r
`

// recordListWhereBase 一覧の対象を自ユーザの論理削除されていない書誌に絞る固定条件
//
// 収録物(論理削除されていない書誌からCONTAINER_IDで参照されている書誌)は一覧に出さない。
// CountActiveArticles > 0 を収録物とみなす詳細参照・削除・更新のガードと同じ定義にしている。
// プレースホルダは1つ(内部用ユーザID)。
const recordListWhereBase = `
	where r.USER_ID = ?
	  and r.DEL_FLG = 0
	  and not exists (
			select 1
			from  e_records a
			where a.CONTAINER_ID = r.ID
			  and a.DEL_FLG = 0
		)
`

// recordListWhereContributorName 貢献者名の部分一致
//
// 書誌の貢献者、または論理削除されていない収録物の貢献者が対象。
// 貢献・貢献者本体とも論理削除済みは除外する(SelectComponents の取得条件と同じ)。
// EXISTSで判定するので、複数の貢献者が一致しても書誌は1件のまま。
// プレースホルダは2つ(書誌・収録物の順。いずれも同じパターン)。
const recordListWhereContributorName = `
	and (
		exists (
			select 1
			from  e_record_contributions c
			inner join e_contributors co
			  on  co.ID = c.E_CONTRIBUTOR_ID
			  and co.DEL_FLG = 0
			where c.E_RECORD_ID = r.ID
			  and c.DEL_FLG = 0
			  and co.RAW_NAME like ? escape '!'
		)
		or
		exists (
			select 1
			from  e_records ct
			inner join e_record_contributions c
			  on  c.E_RECORD_ID = ct.ID
			  and c.DEL_FLG = 0
			inner join e_contributors co
			  on  co.ID = c.E_CONTRIBUTOR_ID
			  and co.DEL_FLG = 0
			where ct.ID = r.CONTAINER_ID
			  and ct.DEL_FLG = 0
			  and co.RAW_NAME like ? escape '!'
		)
	)
`

// recordListWhereTitle タイトルの部分一致
//
// 書誌の主題・副題、または論理削除されていない収録物の主題・副題が対象。
// プレースホルダは4つ(書誌の主題・副題、収録物の主題・副題の順。いずれも同じパターン)。
const recordListWhereTitle = `
	and (
		r.MAIN_TITLE like ? escape '!'
		or
		r.SUB_TITLE like ? escape '!'
		or
		exists (
			select 1
			from e_records ct
			where ct.ID = r.CONTAINER_ID
			  and ct.DEL_FLG = 0
			  and (
				ct.MAIN_TITLE like ? escape '!'
				or
				ct.SUB_TITLE like ? escape '!'
			)
		)
	)
`

// recordListWherePublishYear 出版年の一致
//
// カラムに関数(YEAR())をかけず、[その年の1月1日, 翌年の1月1日) の範囲で比較する。
// 出版日が未登録(NULL)の書誌は含まれない。
// プレースホルダは2つ(範囲の開始日・終了日)。
const recordListWherePublishYear = `
	and
		r.PUBLICATION_DATE >= ?
	and
		r.PUBLICATION_DATE < ?
`

// 一覧の並び順(ORDER BY)
//
// プレースホルダを使えないため、並び替えキー×並び順ごとに固定の句を定義し、
// recordListOrderBy のホワイトリストから引く。
// 同順位はIDの昇順で並べ、offsetによるページングで重複・取りこぼしが起きないようにする。
// 出版日が未登録の書誌は、昇順・降順とも末尾に置く。
const (
	recordListOrderByPublishYearAsc = `
	order by
		r.PUBLICATION_DATE is null,
		r.PUBLICATION_DATE asc,
		r.ID asc
`
	recordListOrderByPublishYearDesc = `
	order by
		r.PUBLICATION_DATE is null,
		r.PUBLICATION_DATE desc,
		r.ID asc
`
	recordListOrderByInsertDateAsc = `
	order by
		r.INS_DT asc,
		r.ID asc
`
	recordListOrderByInsertDateDesc = `
	order by
		r.INS_DT desc,
		r.ID asc
`
)

// recordListQueryLimit 一覧の取得件数と取得開始位置
// プレースホルダは2つ(limit・offsetの順)。
const recordListQueryLimit = `
	limit ?
	offset ?
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

// updateRecordEntryQuery 書誌を1件更新する
//
// リクエストの内容でそのまま上書きする(NULL許容カラムはNULLで上書きされる)。
// 所有権確認を兼ねるため内部用ユーザIDを条件に含め、論理削除済みは対象外とする。
// M_RECORD_ID はPh0未使用、DEL_FLG は削除APIの責務なので更新対象にしない。
//
// WHERE の USER_ID が守るのは「更新される行」の所有権だけで、SET する CONTAINER_ID が
// 同じユーザの書誌かは検証しない(fk_er_container も e_records への存在しか見ない)。
// コンテナIDの所有権確認は呼び出し側(usecase)の責務とする。
// Ph0ではコンテナIDは所有権確認済みの行から読むかサーバ側で生成するかのどちらかで、
// クライアントから受け取る経路がないため問題にならないが、リクエストでコンテナIDを
// 受け取る仕様を追加する場合は、usecase側で SelectByPk(id, uid) を通してから渡すこと。
const updateRecordEntryQuery = `
	update e_records
	set
		TYPE_ID = ?,
		MAIN_TITLE = ?,
		SUB_TITLE = ?,
		PAGE_RANGE = ?,
		VOLUME = ?,
		ISSUE = ?,
		PUBLISHER = ?,
		PUBLICATION_DATE = ?,
		DATE_PRECISION = ?,
		READ_STATUS = ?,
		CONTAINER_ID = ?,
		UPD_PG = ?,
		UPD_ID = 'system',
		UPD_DT = UTC_TIMESTAMP(6)
	where
		ID = ?
	and
		USER_ID = ?
	and
		DEL_FLG = 0
`

// 構成要素の物理削除クエリ(更新時の洗い替え用)
//
// 子テーブルのUNIQUE制約(E_RECORD_ID + DISPLAY_NUMBER 等)に DEL_FLG が含まれないため、
// 論理削除して同じ表示順で再INSERTすると重複エラーになる。更新では物理削除する。
// 貢献(e_record_contributions)は貢献者(e_contributors)より先に消す(FK順)。

const deleteRecordFieldValuesQuery = `
	delete from e_record_field_values
	where
		E_RECORD_ID = ?
`

const deleteRecordMemosQuery = `
	delete from e_record_memos
	where
		E_RECORD_ID = ?
`

const deleteRecordUrlsQuery = `
	delete from e_record_urls
	where
		E_RECORD_ID = ?
`

const deleteRecordContributionsQuery = `
	delete from e_record_contributions
	where
		E_RECORD_ID = ?
`

// selectContributorIdsByRecordIdQuery 書誌が参照している貢献者IDを取得する
//
// 貢献を物理削除すると参照の手がかりが失われるため、削除前に取得しておく。
const selectContributorIdsByRecordIdQuery = `
	select
		E_CONTRIBUTOR_ID
	from
		e_record_contributions
	where
		E_RECORD_ID = ?
`

// deleteUnreferencedContributorsQueryBase 他の書誌から参照されていない貢献者を物理削除する
//
// Ph0では貢献者と書誌は1対1なので常に削除されるが、Ph1で名寄せ(貢献者の共有)を
// 導入したときに他の書誌の貢献者を消さないよう、参照の有無を条件に入れている。
const deleteUnreferencedContributorsQueryBase = `
	delete from e_contributors
	where
		not exists (
			select
				1
			from
				e_record_contributions
			where
				e_record_contributions.E_CONTRIBUTOR_ID = e_contributors.ID
		)
	and
		ID in (`

const deleteUnreferencedContributorsQueryEnd = `)
`
