insert into m_policies(ID,VERSION,NAME,CONTENT,EFFECTIVE_DT,DEL_FLG,INS_PG,INS_ID,INS_DT,UPD_PG,UPD_ID,UPD_DT)
values
-- 正常：最新の１件
('P001','test_0.01','_test_name_0.01_','_test_content_0.01_','2026-05-10',0,'admin','admin',now(),'admin','admin',now()),
-- ダミー：最新でない
('P001','test_0.00','_test_name_0.00_','_test_content_0.00_','2026-05-09',0,'admin','admin',now(),'admin','admin',now()),
-- ダミー：現在時刻より後
('P001','test_1.01','_test_name_1.01_','_test_content_1.01_','9999-12-31',0,'admin','admin',now(),'admin','admin',now()),
-- ダミー：規約IDが違う
('P000','test_0.01','_test_name_0.01_','_test_content_0.01_','2026-05-10',0,'admin','admin',now(),'admin','admin',now()),
-- ダミー：論理削除済み
('P001','test_0.02','_test_name_0.02_','_test_content_0.02_','2026-05-11',1,'admin','admin',now(),'admin','admin',now());
