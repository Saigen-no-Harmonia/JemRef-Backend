-- [REF-API-001] 書誌一覧参照テスト用フィクスチャ(検索条件・除外条件の検証用)
-- マスタ(m_record_types / m_roles / m_record_fields)は sql/seed.sql を別途読み込む前提
--
-- 条件なしで一覧に出るのは、ユーザ1001の論理削除されていない書誌のうち、収録物を除いた16件。
-- 一覧に出てはいけない行(他ユーザ・論理削除済み・収録物)は、タイトルや氏名に「他ユーザ」「削除済」等を含め、
-- 誤って取得された場合に判別できるようにしている。
-- M_RECORD_ID / M_CONTRIBUTOR_ID は seed に該当マスタがないためNULL固定とする。

insert into m_users
(ID,PUBLIC_ID,FIREBASE_ID,EMAIL,TERMS_AGREED_DT,TERMS_VERSION,PRIVACY_POLICY_AGREED_DT,PRIVACY_POLICY_VERSION,DELETED_AT,DEL_FLG,INS_PG,INS_ID,INS_DT,UPD_PG,UPD_ID,UPD_DT)
values
(1001,"SAMPLE000000000000001001","_firebase_uid001_","sample001@example.com","2026-01-01 12:00:00","1.0","2026-02-02 14:00:00","2.0",null,0,
"ins_pg","ins_id",now(),"upd_pg","upd_id",now()),
(1002,"SAMPLE000000000000001002","_firebase_uid002_","sample002@example.com","2026-01-02 12:00:00","1.0","2026-02-03 14:00:00","2.0",null,0,
"ins_pg","ins_id",now(),"upd_pg","upd_id",now());

insert into e_records
(ID,USER_ID,TYPE_ID,MAIN_TITLE,SUB_TITLE,PAGE_RANGE,VOLUME,ISSUE,PUBLISHER,PUBLICATION_DATE,DATE_PRECISION,READ_STATUS,CONTAINER_ID,M_RECORD_ID,DEL_FLG,INS_PG,INS_ID,INS_DT,UPD_PG,UPD_ID,UPD_DT)
values
-- ---- 収録物(一覧に出ない)。書誌のCONTAINER_IDから参照されるため先に登録する(FK順) ----
-- 雑誌。主題・副題・貢献者で、収録する書誌がヒットする
("LST_CON_JOURNAL",1001,"RT000011","教育学研究","特集 学級",null,"第88巻","第1号","日本教育学会","2021-03-01",1,null,null,null,0,
"ins_pg","system",now(),"upd_pg","system",now()),
-- 編著書。主題・副題・貢献者で、収録する書誌がヒットする
("LST_CON_BOOK",1001,"RT000021","境界線の学校史","戦後日本の周縁",null,null,null,"東京大学出版会","2020-12-31",0,null,null,null,0,
"ins_pg","system",now(),"upd_pg","system",now()),
-- 論理削除済みの雑誌。主題・貢献者でヒットしてはいけない
("LST_CON_DELETED",1001,"RT000011","削除済の雑誌",null,null,null,null,null,null,null,null,null,null,1,
"ins_pg","system",now(),"upd_pg","system",now()),

-- ---- 一覧に出る書誌(16件) ----
-- 全カラムに値が入った単著書
("LST_SOLO_FULL",1001,"RT000001","音楽教育の歴史","その展開","1-300","第1巻","第2号","音楽之友社","2020-05-10",0,1,null,null,0,
"ins_pg","system",now(),"upd_pg","system",now()),
-- 雑誌記事(収録物 LST_CON_JOURNAL)
("LST_ART_JOURNAL",1001,"RT000012","戦後の学校",null,"10-20",null,null,"日本教育学会","2021-03-01",1,null,"LST_CON_JOURNAL",null,0,
"ins_pg","system",now(),"upd_pg","system",now()),
-- 編著書記事(収録物 LST_CON_BOOK)。出版年の境界(12/31)
("LST_ART_BOOK",1001,"RT000022","第1章 境界線","制度の形成","15-46",null,null,"東京大学出版会","2020-12-31",0,null,"LST_CON_BOOK",null,0,
"ins_pg","system",now(),"upd_pg","system",now()),
-- どの書誌からも参照されていない雑誌(収録物ではないので一覧に出る)
("LST_STANDALONE",1001,"RT000011","単独で登録した雑誌",null,null,null,null,null,null,null,null,null,null,0,
"ins_pg","system",now(),"upd_pg","system",now()),
-- ワイルドカード文字の検証用(「100%」で前者だけがヒットする)
("LST_PCT",1001,"RT000001","100%の理解",null,null,null,null,null,null,null,null,null,null,0,
"ins_pg","system",now(),"upd_pg","system",now()),
("LST_PCT_PLAIN",1001,"RT000001","100点の理解",null,null,null,null,null,null,null,null,null,null,0,
"ins_pg","system",now(),"upd_pg","system",now()),
("LST_USC",1001,"RT000001","a_bの研究",null,null,null,null,null,null,null,null,null,null,0,
"ins_pg","system",now(),"upd_pg","system",now()),
("LST_USC_PLAIN",1001,"RT000001","axbの研究",null,null,null,null,null,null,null,null,null,null,0,
"ins_pg","system",now(),"upd_pg","system",now()),
-- エスケープ文字(!)の検証用(「意!重」で前者だけがヒットする)
("LST_ESC",1001,"RT000001","注意!重要",null,null,null,null,null,null,null,null,null,null,0,
"ins_pg","system",now(),"upd_pg","system",now()),
("LST_ESC_PLAIN",1001,"RT000001","注意重要",null,null,null,null,null,null,null,null,null,null,0,
"ins_pg","system",now(),"upd_pg","system",now()),
-- 貢献者が2人とも「田中」で一致する
("LST_MULTI",1001,"RT000001","共著の本",null,null,null,null,null,null,null,null,null,null,0,
"ins_pg","system",now(),"upd_pg","system",now()),
-- 貢献(e_record_contributions)が論理削除済み
("LST_DEL_CB",1001,"RT000001","貢献削除の本",null,null,null,null,null,null,null,null,null,null,0,
"ins_pg","system",now(),"upd_pg","system",now()),
-- 貢献者本体(e_contributors)が論理削除済み
("LST_DEL_CR",1001,"RT000001","貢献者削除の本",null,null,null,null,null,null,null,null,null,null,0,
"ins_pg","system",now(),"upd_pg","system",now()),
-- 論理削除済みの収録物を指している雑誌記事
("LST_ART_DELCON",1001,"RT000012","孤立した記事",null,null,null,null,null,null,null,null,"LST_CON_DELETED",null,0,
"ins_pg","system",now(),"upd_pg","system",now()),
-- 出版年の境界(1/1)
("LST_YEAR_START",1001,"RT000001","年始の本",null,null,null,null,null,"2020-01-01",2,null,null,null,0,
"ins_pg","system",now(),"upd_pg","system",now()),
-- 翌年の1/1(2020年には含まれない)
("LST_YEAR_NEXT",1001,"RT000001","翌年の本",null,null,null,null,null,"2021-01-01",2,null,null,null,0,
"ins_pg","system",now(),"upd_pg","system",now()),

-- ---- 一覧に出ない書誌 ----
-- 他ユーザの書誌(主題・貢献者が自ユーザの書誌と一致する)
("LST_OTHER_USER",1002,"RT000001","音楽教育の他ユーザ",null,null,null,null,null,"2020-05-10",0,null,null,null,0,
"ins_pg","system",now(),"upd_pg","system",now()),
-- 論理削除済みの書誌(主題が自ユーザの書誌と一致する)
("LST_DELETED",1001,"RT000001","音楽教育の削除済",null,null,null,null,null,"2020-05-10",0,null,null,null,1,
"ins_pg","system",now(),"upd_pg","system",now());

insert into e_contributors
(ID,RAW_NAME,USER_ID,M_CONTRIBUTOR_ID,DEL_FLG,INS_PG,INS_ID,INS_DT,UPD_PG,UPD_ID,UPD_DT)
values
(9001,"山田太郎",1001,null,0,"ins_pg","system",now(),"upd_pg","system",now()),
(9002,"佐藤花子",1001,null,0,"ins_pg","system",now(),"upd_pg","system",now()),
(9003,"日本教育学会",1001,null,0,"ins_pg","system",now(),"upd_pg","system",now()),
(9004,"鈴木一郎",1001,null,0,"ins_pg","system",now(),"upd_pg","system",now()),
(9005,"木村元",1001,null,0,"ins_pg","system",now(),"upd_pg","system",now()),
(9006,"田中一",1001,null,0,"ins_pg","system",now(),"upd_pg","system",now()),
(9007,"田中二",1001,null,0,"ins_pg","system",now(),"upd_pg","system",now()),
-- 貢献側が論理削除済み
(9008,"消された貢献",1001,null,0,"ins_pg","system",now(),"upd_pg","system",now()),
-- 貢献者本体が論理削除済み
(9009,"消された貢献者",1001,null,1,"ins_pg","system",now(),"upd_pg","system",now()),
-- 論理削除済みの収録物の貢献者(貢献・貢献者本体は生存。収録物のDEL_FLGだけで除外されることを検証する)
(9010,"削除雑誌の編者",1001,null,0,"ins_pg","system",now(),"upd_pg","system",now()),
-- 他ユーザの貢献者
(9011,"山田太郎",1002,null,0,"ins_pg","system",now(),"upd_pg","system",now());

insert into e_record_contributions
(E_RECORD_ID,E_CONTRIBUTOR_ID,ROLE_ID,DISPLAY_NUMBER,DEL_FLG,INS_PG,INS_ID,INS_DT,UPD_PG,UPD_ID,UPD_DT)
values
("LST_SOLO_FULL",9001,"RL000010",0,0,"ins_pg","system",now(),"upd_pg","system",now()),
("LST_ART_JOURNAL",9002,"RL000010",0,0,"ins_pg","system",now(),"upd_pg","system",now()),
("LST_CON_JOURNAL",9003,"RL000030",0,0,"ins_pg","system",now(),"upd_pg","system",now()),
("LST_ART_BOOK",9004,"RL000010",0,0,"ins_pg","system",now(),"upd_pg","system",now()),
("LST_CON_BOOK",9005,"RL000030",0,0,"ins_pg","system",now(),"upd_pg","system",now()),
("LST_MULTI",9006,"RL000010",0,0,"ins_pg","system",now(),"upd_pg","system",now()),
("LST_MULTI",9007,"RL000010",1,0,"ins_pg","system",now(),"upd_pg","system",now()),
("LST_DEL_CB",9008,"RL000010",0,1,"ins_pg","system",now(),"upd_pg","system",now()),
("LST_DEL_CR",9009,"RL000010",0,0,"ins_pg","system",now(),"upd_pg","system",now()),
("LST_CON_DELETED",9010,"RL000030",0,0,"ins_pg","system",now(),"upd_pg","system",now()),
("LST_OTHER_USER",9011,"RL000010",0,0,"ins_pg","system",now(),"upd_pg","system",now());
