insert into m_policies(ID,VERSION,NAME,CONTENT,EFFECTIVE_DT,DEL_FLG,INS_PG,INS_ID,INS_DT,UPD_PG,UPD_ID,UPD_DT)
values
-- 正常
('P001','test_0.01','_test_name_0.01_','_test_content_0.01_','2026-05-10',0,'admin','admin',now(),'admin','admin',now()),
-- ダミー：バージョンが異なる
('P001','test_0.02','_test_name_0.02_','_test_content_0.02_','2026-05-12',0,'admin','admin',now(),'admin','admin',now()),
-- ダミー：規約IDが異なる
('P002','test_0.01','_test_name_P002-0.01_','_test_content_P002-0.01_','2026-05-13',0,'admin','admin',now(),'admin','admin',now());
