insert into m_users
(ID,PUBLIC_ID,FIREBASE_ID,EMAIL,TERMS_AGREED_DT,TERMS_VERSION,PRIVACY_POLICY_AGREED_DT,PRIVACY_POLICY_VERSION,DELETED_AT,DEL_FLG,INS_PG,INS_ID,INS_DT,UPD_PG,UPD_ID,UPD_DT)
values
-- 正常データ
(1001, "SAMPLE000000000000000099","_firebase_uid001_","sample001@example.com","2026-01-01 12:00:00","1.0","2026-02-02 14:00:00","2.0","2999-12-31 23:59:59",0,
"ins_pg","ins_id",now(),"upd_pg","upd_id",now()),

-- 正常データ2（ID等が異なるダミー）
(1002, "SAMPLE000000000000000098","_firebase_uid002_","sample002@example.com","2026-01-02 12:00:00","1.1","2026-02-03 14:00:00","2.1","2999-12-31 23:59:58",0,
"ins_pg","ins_id",now(),"upd_pg","upd_id",now()),

-- 論理削除済み
(1003, "SAMPLE000000000000000097","_firebase_uid003_","sample003@example.com","2026-01-03 12:00:00","1.2","2026-02-04 14:00:00","2.2","2999-12-31 23:59:57",1,
"ins_pg","ins_id",now(),"upd_pg","upd_id",now());
