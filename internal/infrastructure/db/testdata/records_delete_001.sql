-- [REF-API-005] 書誌情報削除テスト用フィクスチャ
-- マスタ(m_roles / m_record_types / m_record_fields)は sql/seed.sql を別途読み込む前提
--
-- 論理削除済みの行は UPD_PG = 'already_deleted' / UPD_DT = '2020-01-01 00:00:00' とし、
-- 「再削除しても更新されない」ことを検証できるようにしている。

insert into m_users
(ID,PUBLIC_ID,FIREBASE_ID,EMAIL,TERMS_AGREED_DT,TERMS_VERSION,PRIVACY_POLICY_AGREED_DT,PRIVACY_POLICY_VERSION,DELETED_AT,DEL_FLG,INS_PG,INS_ID,INS_DT,UPD_PG,UPD_ID,UPD_DT)
values
(1001,"SAMPLE000000000000001001","_firebase_uid001_","sample001@example.com","2026-01-01 12:00:00","1.0","2026-02-02 14:00:00","2.0",null,0,
"ins_pg","ins_id",now(),"upd_pg","upd_id",now()),
(1002,"SAMPLE000000000000001002","_firebase_uid002_","sample002@example.com","2026-01-02 12:00:00","1.0","2026-02-03 14:00:00","2.0",null,0,
"ins_pg","ins_id",now(),"upd_pg","upd_id",now());

insert into e_records
(ID,USER_ID,TYPE_ID,MAIN_TITLE,PARENT_ID,DEL_FLG,INS_PG,INS_ID,INS_DT,UPD_PG,UPD_ID,UPD_DT)
values
-- 親を持たない単著書。子要素をフルに持つ(論理削除の主対象)
("DEL_SOLO_0000000000000001",1001,"RT000001","単独の単著書",null,0,
"ins_pg","system",now(),"upd_pg","system",now()),

-- 子1件を持つ収録物と、その子(他の書誌の子要素に影響しないことの確認に使う)
("DEL_CONTAINER_000000001",1001,"RT000011","子1件の雑誌",null,0,
"ins_pg","system",now(),"upd_pg","system",now()),
("DEL_CHILD_A_00000000001",1001,"RT000012","子A",'DEL_CONTAINER_000000001',0,
"ins_pg","system",now(),"upd_pg","system",now()),

-- 子2件(生1/削除済1)を持つ収録物。子要素行は持たない
("DEL_CONTAINER_000000002",1001,"RT000011","子2件の雑誌",null,0,
"ins_pg","system",now(),"upd_pg","system",now()),
("DEL_CHILD_B_00000000001",1001,"RT000012","子B1(生存)",'DEL_CONTAINER_000000002',0,
"ins_pg","system",now(),"upd_pg","system",now()),
("DEL_CHILD_B_00000000002",1001,"RT000012","子B2(論理削除済)",'DEL_CONTAINER_000000002',1,
"ins_pg","system",now(),"already_deleted","system","2020-01-01 00:00:00"),

-- 他ユーザの書誌
("DEL_OTHER_USER_00000001",1002,"RT000001","他ユーザの単著書",null,0,
"ins_pg","system",now(),"upd_pg","system",now()),

-- すでに論理削除済みの書誌
("DEL_DELETED_00000000001",1001,"RT000001","論理削除済の単著書",null,1,
"ins_pg","system",now(),"already_deleted","system","2020-01-01 00:00:00"),

-- 子要素の一部が論理削除済みの書誌
("DEL_PARTIAL_00000000001",1001,"RT000001","子要素が一部削除済の単著書",null,0,
"ins_pg","system",now(),"upd_pg","system",now());

insert into e_contributors
(ID,RAW_NAME,USER_ID,M_CONTRIBUTOR_ID,DEL_FLG,INS_PG,INS_ID,INS_DT,UPD_PG,UPD_ID,UPD_DT)
values
(9001,"貢献者A",1001,null,0,"ins_pg","system",now(),"upd_pg","system",now()),
(9002,"貢献者B",1001,null,0,"ins_pg","system",now(),"upd_pg","system",now()),
(9003,"収録物の編者",1001,null,0,"ins_pg","system",now(),"upd_pg","system",now()),
(9004,"一部削除の貢献者1",1001,null,0,"ins_pg","system",now(),"upd_pg","system",now()),
(9005,"一部削除の貢献者2",1001,null,0,"ins_pg","system",now(),"upd_pg","system",now());

insert into e_record_contributions
(E_RECORD_ID,E_CONTRIBUTOR_ID,ROLE_ID,DISPLAY_NUMBER,DEL_FLG,INS_PG,INS_ID,INS_DT,UPD_PG,UPD_ID,UPD_DT)
values
-- DEL_SOLO: 2件とも生存
("DEL_SOLO_0000000000000001",9001,"RL000010",0,0,"ins_pg","system",now(),"upd_pg","system",now()),
("DEL_SOLO_0000000000000001",9002,"RL000020",1,0,"ins_pg","system",now(),"upd_pg","system",now()),
-- DEL_CONTAINER_000000001: 影響を受けないことの確認用
("DEL_CONTAINER_000000001",9003,"RL000030",0,0,"ins_pg","system",now(),"upd_pg","system",now()),
-- DEL_PARTIAL: 表示順0が生存、1が論理削除済
("DEL_PARTIAL_00000000001",9004,"RL000010",0,0,"ins_pg","system",now(),"upd_pg","system",now()),
("DEL_PARTIAL_00000000001",9005,"RL000010",1,1,"ins_pg","system",now(),"already_deleted","system","2020-01-01 00:00:00");

insert into e_record_urls
(E_RECORD_ID,DISPLAY_NUMBER,RAW_LABEL,RAW_URL,USER_ID,DEL_FLG,INS_PG,INS_ID,INS_DT,UPD_PG,UPD_ID,UPD_DT)
values
("DEL_SOLO_0000000000000001",0,"PDF","https://example.com/1",1001,0,"ins_pg","system",now(),"upd_pg","system",now()),
("DEL_SOLO_0000000000000001",1,null,"https://example.com/2",1001,0,"ins_pg","system",now(),"upd_pg","system",now()),
("DEL_PARTIAL_00000000001",0,"生存","https://example.com/alive",1001,0,"ins_pg","system",now(),"upd_pg","system",now()),
("DEL_PARTIAL_00000000001",1,"削除済","https://example.com/deleted",1001,1,"ins_pg","system",now(),"already_deleted","system","2020-01-01 00:00:00");

insert into e_record_memos
(E_RECORD_ID,DISPLAY_NUMBER,TITLE,CONTENT,USER_ID,SHARE_LEVEL,DEL_FLG,INS_PG,INS_ID,INS_DT,UPD_PG,UPD_ID,UPD_DT)
values
("DEL_SOLO_0000000000000001",0,"要約","メモ内容1",1001,0,0,"ins_pg","system",now(),"upd_pg","system",now()),
("DEL_SOLO_0000000000000001",1,null,"メモ内容2",1001,0,0,"ins_pg","system",now(),"upd_pg","system",now()),
("DEL_PARTIAL_00000000001",0,"生存","メモ生存",1001,0,0,"ins_pg","system",now(),"upd_pg","system",now()),
("DEL_PARTIAL_00000000001",1,"削除済","メモ削除済",1001,0,1,"ins_pg","system",now(),"already_deleted","system","2020-01-01 00:00:00");

insert into e_record_field_values
(E_RECORD_ID,RECORD_FIELD_ID,DISPLAY_NUMBER,RAW_VALUE,USER_ID,DEL_FLG,INS_PG,INS_ID,INS_DT,UPD_PG,UPD_ID,UPD_DT)
values
("DEL_SOLO_0000000000000001","RF000001",10,"9784130513555",1001,0,"ins_pg","system",now(),"upd_pg","system",now()),
("DEL_SOLO_0000000000000001","RF000011",20,"第2版",1001,0,"ins_pg","system",now(),"upd_pg","system",now()),
("DEL_CONTAINER_000000001","RF000001",10,"9784130513556",1001,0,"ins_pg","system",now(),"upd_pg","system",now()),
("DEL_PARTIAL_00000000001","RF000001",10,"生存",1001,0,"ins_pg","system",now(),"upd_pg","system",now()),
("DEL_PARTIAL_00000000001","RF000011",20,"削除済",1001,1,"ins_pg","system",now(),"already_deleted","system","2020-01-01 00:00:00");
