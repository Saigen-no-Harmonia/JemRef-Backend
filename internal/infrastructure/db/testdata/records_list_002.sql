-- [REF-API-001] 書誌一覧参照テスト用フィクスチャ(並び替え・ページングの検証用)
-- マスタ(m_record_types / m_roles / m_record_fields)は sql/seed.sql を別途読み込む前提
--
-- 並び順の期待値がIDの辞書順(SORT_A < SORT_B < ...)と一致しないよう、出版日と登録日時を選んでいる。
-- SORT_E1 / SORT_E2 は出版日・登録日時とも同じ値で、同順位がIDの昇順になることの検証に使う。
-- INS_DTはUTCで明示する(並び順を固定するため now() は使わない)。

insert into m_users
(ID,PUBLIC_ID,FIREBASE_ID,EMAIL,TERMS_AGREED_DT,TERMS_VERSION,PRIVACY_POLICY_AGREED_DT,PRIVACY_POLICY_VERSION,DELETED_AT,DEL_FLG,INS_PG,INS_ID,INS_DT,UPD_PG,UPD_ID,UPD_DT)
values
(1001,"SAMPLE000000000000001001","_firebase_uid001_","sample001@example.com","2026-01-01 12:00:00","1.0","2026-02-02 14:00:00","2.0",null,0,
"ins_pg","ins_id",now(),"upd_pg","upd_id",now());

insert into e_records
(ID,USER_ID,TYPE_ID,MAIN_TITLE,SUB_TITLE,PAGE_RANGE,VOLUME,ISSUE,PUBLISHER,PUBLICATION_DATE,DATE_PRECISION,READ_STATUS,CONTAINER_ID,M_RECORD_ID,DEL_FLG,INS_PG,INS_ID,INS_DT,UPD_PG,UPD_ID,UPD_DT)
values
("SORT_A",1001,"RT000001","並び替えA",null,null,null,null,null,"2020-06-01",0,null,null,null,0,
"ins_pg","system","2026-01-03 00:00:00","upd_pg","system",now()),
("SORT_B",1001,"RT000001","並び替えB",null,null,null,null,null,"2019-01-01",0,null,null,null,0,
"ins_pg","system","2026-01-01 00:00:00","upd_pg","system",now()),
-- 出版日なし
("SORT_C",1001,"RT000001","並び替えC",null,null,null,null,null,null,null,null,null,null,0,
"ins_pg","system","2026-01-04 00:00:00","upd_pg","system",now()),
("SORT_D",1001,"RT000001","並び替えD",null,null,null,null,null,"2021-01-01",0,null,null,null,0,
"ins_pg","system","2026-01-02 00:00:00","upd_pg","system",now()),
-- 同順位(出版日・登録日時とも同じ)
("SORT_E1",1001,"RT000001","並び替えE1",null,null,null,null,null,"2022-01-01",0,null,null,null,0,
"ins_pg","system","2026-01-05 00:00:00","upd_pg","system",now()),
("SORT_E2",1001,"RT000001","並び替えE2",null,null,null,null,null,"2022-01-01",0,null,null,null,0,
"ins_pg","system","2026-01-05 00:00:00","upd_pg","system",now()),
-- 出版日なし
("SORT_F",1001,"RT000001","並び替えF",null,null,null,null,null,null,null,null,null,null,0,
"ins_pg","system","2026-01-06 00:00:00","upd_pg","system",now());
