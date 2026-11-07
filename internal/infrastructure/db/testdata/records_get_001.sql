-- [REF-API-002] 書誌詳細参照テスト用フィクスチャ
-- マスタ(m_record_types / m_roles / m_record_fields)は sql/seed.sql を別途読み込む前提
--
-- M_RECORD_ID / M_CONTRIBUTOR_ID はマスタ(m_records / m_contributors)へのFKを持つが、
-- seedに該当マスタがないためNULL固定とする。
-- 論理削除済みの行は RAW_NAME / RAW_URL 等に「削除済」を含め、誤って取得された場合に判別できるようにしている。

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
-- 収録物(雑誌)。構成要素を持ち、書誌との取り違え検証に使う
("GET_CONTAINER_00000000001",1001,"RT000011","境界線の学校史",null,null,"第1巻","第11号","東京大学出版会","2022-11-30",0,1,null,null,0,
"ins_pg","system",now(),"upd_pg","system",now()),

-- 全カラムに値が入った書誌(収録物あり)
("GET_FULL_00000000000001",1001,"RT000012","第1章 「学校」制度の境界線","その形成と展開","15-46","第1巻","第11号","東京大学出版会","2022-11-30",0,1,"GET_CONTAINER_00000000001",null,0,
"ins_pg","system",now(),"upd_pg","system",now()),

-- 必須項目のみ(任意項目はすべてNULL)。構成要素も持たない
("GET_MINIMAL_0000000001",1001,"RT000001","必須のみの単著書",null,null,null,null,null,null,null,null,null,null,0,
"ins_pg","system",now(),"upd_pg","system",now()),

-- 任意項目が空文字(NULLではない)
("GET_EMPTY_STR_000000001",1001,"RT000001","空文字の単著書","","","","","",null,null,null,null,null,0,
"ins_pg","system",now(),"upd_pg","system",now()),

-- 構成要素を表示順の逆順でINSERTする(昇順で返ることの検証用)
("GET_ORDER_000000000001",1001,"RT000001","表示順検証の単著書",null,null,null,null,null,null,null,null,null,null,0,
"ins_pg","system",now(),"upd_pg","system",now()),

-- 構成要素に生存と論理削除済みが混在する
("GET_DEL_COMPONENT_00001",1001,"RT000001","構成要素一部削除の単著書",null,null,null,null,null,null,null,null,null,null,0,
"ins_pg","system",now(),"upd_pg","system",now()),

-- 貢献自体は生存だが、貢献者本体(e_contributors)が論理削除済み
("GET_DEL_CONTRIB_000001",1001,"RT000001","貢献者削除済の単著書",null,null,null,null,null,null,null,null,null,null,0,
"ins_pg","system",now(),"upd_pg","system",now()),

-- 構成要素の任意項目がNULL
("GET_NULL_COMPONENT_0001",1001,"RT000001","構成要素NULLの単著書",null,null,null,null,null,null,null,null,null,null,0,
"ins_pg","system",now(),"upd_pg","system",now()),

-- 他ユーザの書誌
("GET_OTHER_USER_0000001",1002,"RT000001","他ユーザの単著書",null,null,null,null,null,null,null,null,null,null,0,
"ins_pg","system",now(),"upd_pg","system",now()),

-- 論理削除済みの書誌
("GET_DELETED_0000000001",1001,"RT000001","論理削除済の単著書",null,null,null,null,null,null,null,null,null,null,1,
"ins_pg","system",now(),"upd_pg","system",now());

insert into e_contributors
(ID,RAW_NAME,USER_ID,M_CONTRIBUTOR_ID,DEL_FLG,INS_PG,INS_ID,INS_DT,UPD_PG,UPD_ID,UPD_DT)
values
-- GET_FULL の貢献者
-- M_CONTRIBUTOR_IDはm_contributorsへのFKを持つが、seedに貢献者マスタがないためNULL固定
(8001,"木村元",1001,null,0,"ins_pg","system",now(),"upd_pg","system",now()),
(8002,"訳者B",1001,null,0,"ins_pg","system",now(),"upd_pg","system",now()),
-- GET_CONTAINER の貢献者
(8003,"編者A",1001,null,0,"ins_pg","system",now(),"upd_pg","system",now()),
-- GET_ORDER の貢献者(表示順の逆順で登録する)
(8011,"三番目の貢献者",1001,null,0,"ins_pg","system",now(),"upd_pg","system",now()),
(8012,"二番目の貢献者",1001,null,0,"ins_pg","system",now(),"upd_pg","system",now()),
(8013,"一番目の貢献者",1001,null,0,"ins_pg","system",now(),"upd_pg","system",now()),
-- GET_DEL_COMPONENT の貢献者(貢献側が削除済み)
(8021,"生存の貢献者",1001,null,0,"ins_pg","system",now(),"upd_pg","system",now()),
(8022,"削除済の貢献",1001,null,0,"ins_pg","system",now(),"upd_pg","system",now()),
-- GET_DEL_CONTRIB の貢献者(貢献者本体が削除済み)
(8031,"生存の貢献者2",1001,null,0,"ins_pg","system",now(),"upd_pg","system",now()),
(8032,"削除済の貢献者",1001,null,1,"ins_pg","system",now(),"upd_pg","system",now()),
-- GET_NULL_COMPONENT の貢献者(M_CONTRIBUTOR_IDがNULL)
(8041,"マスタ未突合の貢献者",1001,null,0,"ins_pg","system",now(),"upd_pg","system",now());

insert into e_record_contributions
(E_RECORD_ID,E_CONTRIBUTOR_ID,ROLE_ID,DISPLAY_NUMBER,DEL_FLG,INS_PG,INS_ID,INS_DT,UPD_PG,UPD_ID,UPD_DT)
values
("GET_FULL_00000000000001",8001,"RL000010",0,0,"ins_pg","system",now(),"upd_pg","system",now()),
("GET_FULL_00000000000001",8002,"RL000020",1,0,"ins_pg","system",now(),"upd_pg","system",now()),
("GET_CONTAINER_00000000001",8003,"RL000030",0,0,"ins_pg","system",now(),"upd_pg","system",now()),
-- 表示順の逆順でINSERTする
("GET_ORDER_000000000001",8011,"RL000010",2,0,"ins_pg","system",now(),"upd_pg","system",now()),
("GET_ORDER_000000000001",8012,"RL000010",1,0,"ins_pg","system",now(),"upd_pg","system",now()),
("GET_ORDER_000000000001",8013,"RL000010",0,0,"ins_pg","system",now(),"upd_pg","system",now()),
-- 貢献側の論理削除
("GET_DEL_COMPONENT_00001",8021,"RL000010",0,0,"ins_pg","system",now(),"upd_pg","system",now()),
("GET_DEL_COMPONENT_00001",8022,"RL000010",1,1,"ins_pg","system",now(),"upd_pg","system",now()),
-- 貢献者本体の論理削除(貢献自体は生存)
("GET_DEL_CONTRIB_000001",8031,"RL000010",0,0,"ins_pg","system",now(),"upd_pg","system",now()),
("GET_DEL_CONTRIB_000001",8032,"RL000010",1,0,"ins_pg","system",now(),"upd_pg","system",now()),
("GET_NULL_COMPONENT_0001",8041,"RL000010",0,0,"ins_pg","system",now(),"upd_pg","system",now());

insert into e_record_urls
(E_RECORD_ID,DISPLAY_NUMBER,RAW_LABEL,RAW_URL,USER_ID,DEL_FLG,INS_PG,INS_ID,INS_DT,UPD_PG,UPD_ID,UPD_DT)
values
("GET_FULL_00000000000001",0,"オンラインPDF","https://example.com/pdf",1001,0,"ins_pg","system",now(),"upd_pg","system",now()),
("GET_FULL_00000000000001",1,"書誌情報","https://example.com/info",1001,0,"ins_pg","system",now(),"upd_pg","system",now()),
("GET_ORDER_000000000001",2,"三番目のURL","https://example.com/3",1001,0,"ins_pg","system",now(),"upd_pg","system",now()),
("GET_ORDER_000000000001",1,"二番目のURL","https://example.com/2",1001,0,"ins_pg","system",now(),"upd_pg","system",now()),
("GET_ORDER_000000000001",0,"一番目のURL","https://example.com/1",1001,0,"ins_pg","system",now(),"upd_pg","system",now()),
("GET_DEL_COMPONENT_00001",0,"生存のURL","https://example.com/alive",1001,0,"ins_pg","system",now(),"upd_pg","system",now()),
("GET_DEL_COMPONENT_00001",1,"削除済のURL","https://example.com/deleted",1001,1,"ins_pg","system",now(),"upd_pg","system",now()),
("GET_NULL_COMPONENT_0001",0,null,"https://example.com/nolabel",1001,0,"ins_pg","system",now(),"upd_pg","system",now());

insert into e_record_memos
(E_RECORD_ID,DISPLAY_NUMBER,TITLE,CONTENT,USER_ID,SHARE_LEVEL,DEL_FLG,INS_PG,INS_ID,INS_DT,UPD_PG,UPD_ID,UPD_DT)
values
("GET_FULL_00000000000001",0,"要約","学校の戦後史。",1001,0,0,"ins_pg","system",now(),"upd_pg","system",now()),
("GET_FULL_00000000000001",1,"感想","少し難しい。",1001,0,0,"ins_pg","system",now(),"upd_pg","system",now()),
("GET_ORDER_000000000001",2,"三番目のメモ","内容3",1001,0,0,"ins_pg","system",now(),"upd_pg","system",now()),
("GET_ORDER_000000000001",1,"二番目のメモ","内容2",1001,0,0,"ins_pg","system",now(),"upd_pg","system",now()),
("GET_ORDER_000000000001",0,"一番目のメモ","内容1",1001,0,0,"ins_pg","system",now(),"upd_pg","system",now()),
("GET_DEL_COMPONENT_00001",0,"生存のメモ","生存",1001,0,0,"ins_pg","system",now(),"upd_pg","system",now()),
("GET_DEL_COMPONENT_00001",1,"削除済のメモ","削除済",1001,0,1,"ins_pg","system",now(),"upd_pg","system",now()),
("GET_NULL_COMPONENT_0001",0,null,"見出しなしのメモ",1001,0,0,"ins_pg","system",now(),"upd_pg","system",now());

insert into e_record_field_values
(E_RECORD_ID,RECORD_FIELD_ID,DISPLAY_NUMBER,RAW_VALUE,USER_ID,DEL_FLG,INS_PG,INS_ID,INS_DT,UPD_PG,UPD_ID,UPD_DT)
values
-- 3種類すべての入力項目を持たせる
("GET_FULL_00000000000001","RF000001",10,"9784130513555",1001,0,"ins_pg","system",now(),"upd_pg","system",now()),
("GET_FULL_00000000000001","RF000011",20,"第2版",1001,0,"ins_pg","system",now(),"upd_pg","system",now()),
("GET_FULL_00000000000001","RF000021",30,"School History of Boundary",1001,0,"ins_pg","system",now(),"upd_pg","system",now()),
-- 収録物にも項目値を持たせ、書誌との取り違えを検出できるようにする
("GET_CONTAINER_00000000001","RF000001",10,"9784130519999",1001,0,"ins_pg","system",now(),"upd_pg","system",now()),
-- 表示順の逆順でINSERTする
("GET_ORDER_000000000001","RF000021",30,"三番目の項目値",1001,0,"ins_pg","system",now(),"upd_pg","system",now()),
("GET_ORDER_000000000001","RF000011",20,"二番目の項目値",1001,0,"ins_pg","system",now(),"upd_pg","system",now()),
("GET_ORDER_000000000001","RF000001",10,"一番目の項目値",1001,0,"ins_pg","system",now(),"upd_pg","system",now()),
("GET_DEL_COMPONENT_00001","RF000001",10,"生存の項目値",1001,0,"ins_pg","system",now(),"upd_pg","system",now()),
("GET_DEL_COMPONENT_00001","RF000011",20,"削除済の項目値",1001,1,"ins_pg","system",now(),"upd_pg","system",now());
