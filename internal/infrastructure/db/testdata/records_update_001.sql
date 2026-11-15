-- [REF-API-004] 書誌情報更新テスト用フィクスチャ
-- マスタ(m_record_types / m_roles / m_record_fields)は sql/seed.sql を別途読み込む前提
--
-- M_RECORD_ID / M_CONTRIBUTOR_ID はマスタ(m_records / m_contributors)へのFKを持つが、
-- seedに該当マスタがないためNULL固定とする。
--
-- 判別用のUPD_PG:
--   'untouched'       更新対象外。1行も変更されていないことを検証する行
--   'already_deleted' 論理削除済み。UPD_DTも '2020-01-01 00:00:00' で固定する
--
-- e_records は CONTAINER_ID で自己参照するため、収録物の行を先に並べている
-- (1つのINSERT文でも行単位にFKが検査されるため、順序を入れ替えると Error 1452 になる)。

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
-- 収録物(雑誌)。UPD_FULLのコンテナ。構成要素を持ち、書誌の更新で消えないことの検証に使う
("UPD_CONTAINER_00000001",1001,"RT000011","更新前の雑誌",null,null,"更新前の巻","更新前の号","更新前の出版者","2020-03-15",0,null,null,null,0,
"ins_pg","system",now(),"untouched","system","2020-01-01 00:00:00"),

-- どの書誌からも参照されていない収録物。コンテナIDの付け替え先に使う
("UPD_FREE_CONTAINER_001",1001,"RT000011","付け替え先の雑誌",null,null,null,null,null,null,null,null,null,null,0,
"ins_pg","system",now(),"untouched","system","2020-01-01 00:00:00"),

-- 全カラムに値が入った書誌。構成要素もフルに持つ(論理削除済みの行を各テーブルに1件含む)
("UPD_FULL_00000000000001",1001,"RT000012","更新前の主題","更新前の副題","1-10","更新前の巻","更新前の号","更新前の出版者","2020-03-15",0,1,"UPD_CONTAINER_00000001",null,0,
"ins_pg","system",now(),"upd_pg","system",now()),

-- 必須項目のみ(任意項目はすべてNULL)。構成要素も持たない
("UPD_MINIMAL_000000001",1001,"RT000001","更新前の単著書",null,null,null,null,null,null,null,null,null,null,0,
"ins_pg","system",now(),"upd_pg","system",now()),

-- 貢献者を共有する2件の書誌(Ph0では発生しない状態。名寄せ導入後を想定した検証用)
("UPD_SHARED_A_0000001",1001,"RT000001","貢献者共有の単著書A",null,null,null,null,null,null,null,null,null,null,0,
"ins_pg","system",now(),"upd_pg","system",now()),
("UPD_SHARED_B_0000001",1001,"RT000001","貢献者共有の単著書B",null,null,null,null,null,null,null,null,null,null,0,
"ins_pg","system",now(),"untouched","system","2020-01-01 00:00:00"),

-- 他ユーザの書誌。構成要素あり
("UPD_OTHER_USER_00001",1002,"RT000001","他ユーザの単著書",null,null,null,null,null,null,null,null,null,null,0,
"ins_pg","system",now(),"untouched","system","2020-01-01 00:00:00"),

-- すでに論理削除済みの書誌。構成要素も論理削除済み
("UPD_DELETED_00000001",1001,"RT000001","論理削除済の単著書",null,null,null,null,null,null,null,null,null,null,1,
"ins_pg","system",now(),"already_deleted","system","2020-01-01 00:00:00"),

-- 更新対象外の書誌。構成要素をフルに持ち、他の書誌の更新で影響を受けないことの検証に使う
("UPD_UNTOUCHED_000001",1001,"RT000001","無関係の単著書",null,null,null,null,null,null,null,null,null,null,0,
"ins_pg","system",now(),"untouched","system","2020-01-01 00:00:00");

insert into e_contributors
(ID,RAW_NAME,USER_ID,M_CONTRIBUTOR_ID,DEL_FLG,INS_PG,INS_ID,INS_DT,UPD_PG,UPD_ID,UPD_DT)
values
-- UPD_FULL の貢献者(9203は貢献側が論理削除済み)
(9201,"更新前の著者",1001,null,0,"ins_pg","system",now(),"upd_pg","system",now()),
(9202,"更新前の訳者",1001,null,0,"ins_pg","system",now(),"upd_pg","system",now()),
(9203,"更新前の削除済貢献者",1001,null,0,"ins_pg","system",now(),"upd_pg","system",now()),
-- UPD_CONTAINER の貢献者
(9204,"雑誌の編者",1001,null,0,"ins_pg","system",now(),"untouched","system","2020-01-01 00:00:00"),
-- UPD_SHARED_A と UPD_SHARED_B が共有する貢献者
(9211,"共有される貢献者",1001,null,0,"ins_pg","system",now(),"untouched","system","2020-01-01 00:00:00"),
-- 他ユーザの貢献者
(9221,"他ユーザの著者",1002,null,0,"ins_pg","system",now(),"untouched","system","2020-01-01 00:00:00"),
-- 論理削除済み書誌の貢献者
(9231,"論理削除済の著者",1001,null,1,"ins_pg","system",now(),"already_deleted","system","2020-01-01 00:00:00"),
-- UPD_UNTOUCHED の貢献者
(9241,"無関係の著者1",1001,null,0,"ins_pg","system",now(),"untouched","system","2020-01-01 00:00:00"),
(9242,"無関係の著者2",1001,null,0,"ins_pg","system",now(),"untouched","system","2020-01-01 00:00:00");

insert into e_record_contributions
(E_RECORD_ID,E_CONTRIBUTOR_ID,ROLE_ID,DISPLAY_NUMBER,DEL_FLG,INS_PG,INS_ID,INS_DT,UPD_PG,UPD_ID,UPD_DT)
values
-- UPD_FULL: 表示順0,1が生存、2は論理削除済み(更新では物理削除されることの検証用)
("UPD_FULL_00000000000001",9201,"RL000010",0,0,"ins_pg","system",now(),"upd_pg","system",now()),
("UPD_FULL_00000000000001",9202,"RL000020",1,0,"ins_pg","system",now(),"upd_pg","system",now()),
("UPD_FULL_00000000000001",9203,"RL000010",2,1,"ins_pg","system",now(),"already_deleted","system","2020-01-01 00:00:00"),
("UPD_CONTAINER_00000001",9204,"RL000030",0,0,"ins_pg","system",now(),"untouched","system","2020-01-01 00:00:00"),
-- 同じ貢献者(9211)を2件の書誌が参照する
("UPD_SHARED_A_0000001",9211,"RL000010",0,0,"ins_pg","system",now(),"upd_pg","system",now()),
("UPD_SHARED_B_0000001",9211,"RL000010",0,0,"ins_pg","system",now(),"untouched","system","2020-01-01 00:00:00"),
("UPD_OTHER_USER_00001",9221,"RL000010",0,0,"ins_pg","system",now(),"untouched","system","2020-01-01 00:00:00"),
("UPD_DELETED_00000001",9231,"RL000010",0,1,"ins_pg","system",now(),"already_deleted","system","2020-01-01 00:00:00"),
("UPD_UNTOUCHED_000001",9241,"RL000010",0,0,"ins_pg","system",now(),"untouched","system","2020-01-01 00:00:00"),
("UPD_UNTOUCHED_000001",9242,"RL000020",1,0,"ins_pg","system",now(),"untouched","system","2020-01-01 00:00:00");

insert into e_record_urls
(E_RECORD_ID,DISPLAY_NUMBER,RAW_LABEL,RAW_URL,USER_ID,DEL_FLG,INS_PG,INS_ID,INS_DT,UPD_PG,UPD_ID,UPD_DT)
values
-- UPD_FULL: 表示順0,1が生存、2は論理削除済み
("UPD_FULL_00000000000001",0,"更新前のラベル0","https://example.com/before/0",1001,0,"ins_pg","system",now(),"upd_pg","system",now()),
("UPD_FULL_00000000000001",1,"更新前のラベル1","https://example.com/before/1",1001,0,"ins_pg","system",now(),"upd_pg","system",now()),
("UPD_FULL_00000000000001",2,"更新前の削除済ラベル","https://example.com/before/deleted",1001,1,"ins_pg","system",now(),"already_deleted","system","2020-01-01 00:00:00"),
("UPD_CONTAINER_00000001",0,"雑誌のURL","https://example.com/container",1001,0,"ins_pg","system",now(),"untouched","system","2020-01-01 00:00:00"),
("UPD_OTHER_USER_00001",0,"他ユーザのURL","https://example.com/other",1002,0,"ins_pg","system",now(),"untouched","system","2020-01-01 00:00:00"),
("UPD_DELETED_00000001",0,"論理削除済のURL","https://example.com/deleted",1001,1,"ins_pg","system",now(),"already_deleted","system","2020-01-01 00:00:00"),
("UPD_UNTOUCHED_000001",0,"無関係のURL0","https://example.com/untouched/0",1001,0,"ins_pg","system",now(),"untouched","system","2020-01-01 00:00:00"),
("UPD_UNTOUCHED_000001",1,"無関係のURL1","https://example.com/untouched/1",1001,0,"ins_pg","system",now(),"untouched","system","2020-01-01 00:00:00");

insert into e_record_memos
(E_RECORD_ID,DISPLAY_NUMBER,TITLE,CONTENT,USER_ID,SHARE_LEVEL,DEL_FLG,INS_PG,INS_ID,INS_DT,UPD_PG,UPD_ID,UPD_DT)
values
-- UPD_FULL: 表示順0,1が生存、2は論理削除済み
("UPD_FULL_00000000000001",0,"更新前の見出し0","更新前の内容0",1001,0,0,"ins_pg","system",now(),"upd_pg","system",now()),
("UPD_FULL_00000000000001",1,"更新前の見出し1","更新前の内容1",1001,0,0,"ins_pg","system",now(),"upd_pg","system",now()),
("UPD_FULL_00000000000001",2,"更新前の削除済見出し","更新前の削除済内容",1001,0,1,"ins_pg","system",now(),"already_deleted","system","2020-01-01 00:00:00"),
("UPD_CONTAINER_00000001",0,"雑誌のメモ","雑誌の内容",1001,0,0,"ins_pg","system",now(),"untouched","system","2020-01-01 00:00:00"),
("UPD_OTHER_USER_00001",0,"他ユーザのメモ","他ユーザの内容",1002,0,0,"ins_pg","system",now(),"untouched","system","2020-01-01 00:00:00"),
("UPD_DELETED_00000001",0,"論理削除済のメモ","論理削除済の内容",1001,0,1,"ins_pg","system",now(),"already_deleted","system","2020-01-01 00:00:00"),
("UPD_UNTOUCHED_000001",0,"無関係のメモ0","無関係の内容0",1001,0,0,"ins_pg","system",now(),"untouched","system","2020-01-01 00:00:00"),
("UPD_UNTOUCHED_000001",1,"無関係のメモ1","無関係の内容1",1001,0,0,"ins_pg","system",now(),"untouched","system","2020-01-01 00:00:00");

insert into e_record_field_values
(E_RECORD_ID,RECORD_FIELD_ID,DISPLAY_NUMBER,RAW_VALUE,USER_ID,DEL_FLG,INS_PG,INS_ID,INS_DT,UPD_PG,UPD_ID,UPD_DT)
values
-- UPD_FULL: 3件が生存、1件(RF000001の表示順11)は論理削除済み
("UPD_FULL_00000000000001","RF000001",0,"更新前のISBN",1001,0,"ins_pg","system",now(),"upd_pg","system",now()),
("UPD_FULL_00000000000001","RF000011",1,"更新前の版",1001,0,"ins_pg","system",now(),"upd_pg","system",now()),
("UPD_FULL_00000000000001","RF000021",2,"更新前の原題",1001,0,"ins_pg","system",now(),"upd_pg","system",now()),
("UPD_FULL_00000000000001","RF000001",11,"更新前の削除済ISBN",1001,1,"ins_pg","system",now(),"already_deleted","system","2020-01-01 00:00:00"),
("UPD_CONTAINER_00000001","RF000001",0,"雑誌のISBN",1001,0,"ins_pg","system",now(),"untouched","system","2020-01-01 00:00:00"),
("UPD_OTHER_USER_00001","RF000001",0,"他ユーザのISBN",1002,0,"ins_pg","system",now(),"untouched","system","2020-01-01 00:00:00"),
("UPD_DELETED_00000001","RF000001",0,"論理削除済のISBN",1001,1,"ins_pg","system",now(),"already_deleted","system","2020-01-01 00:00:00"),
("UPD_UNTOUCHED_000001","RF000001",0,"無関係のISBN",1001,0,"ins_pg","system",now(),"untouched","system","2020-01-01 00:00:00"),
("UPD_UNTOUCHED_000001","RF000011",1,"無関係の版",1001,0,"ins_pg","system",now(),"untouched","system","2020-01-01 00:00:00");
