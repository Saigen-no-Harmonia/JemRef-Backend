-- [REF-API-003] 書誌情報登録テスト用フィクスチャ
-- マスタ(m_roles / m_record_types / m_record_fields)は sql/seed.sql を別途読み込む前提

insert into m_users
(ID,PUBLIC_ID,FIREBASE_ID,EMAIL,TERMS_AGREED_DT,TERMS_VERSION,PRIVACY_POLICY_AGREED_DT,PRIVACY_POLICY_VERSION,DELETED_AT,DEL_FLG,INS_PG,INS_ID,INS_DT,UPD_PG,UPD_ID,UPD_DT)
values
-- 登録対象ユーザ
(1001,"SAMPLE000000000000001001","_firebase_uid001_","sample001@example.com","2026-01-01 12:00:00","1.0","2026-02-02 14:00:00","2.0",null,0,
"ins_pg","ins_id",now(),"upd_pg","upd_id",now()),

-- 別ユーザ(他ユーザのデータに影響しないことの確認用)
(1002,"SAMPLE000000000000001002","_firebase_uid002_","sample002@example.com","2026-01-02 12:00:00","1.0","2026-02-03 14:00:00","2.0",null,0,
"ins_pg","ins_id",now(),"upd_pg","upd_id",now());

insert into e_records
(ID,USER_ID,TYPE_ID,MAIN_TITLE,SUB_TITLE,PAGE_RANGE,VOLUME,ISSUE,PUBLISHER,PUBLICATION_DATE,DATE_PRECISION,READ_STATUS,PARENT_ID,M_RECORD_ID,DEL_FLG,INS_PG,INS_ID,INS_DT,UPD_PG,UPD_ID,UPD_DT)
values
-- 既存の収録物(雑誌)。ID重複テストおよび親ID参照テストで使用する
("EXISTINGCONTAINER000000001",1001,"RT000011","既存の雑誌",null,null,"第1巻","第2号","既存出版社","2020-04-01",0,2,null,null,0,
"ins_pg","ins_id",now(),"upd_pg","upd_id",now());
