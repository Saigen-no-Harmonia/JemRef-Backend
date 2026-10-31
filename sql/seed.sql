insert into m_policies(ID,VERSION,NAME,CONTENT,EFFECTIVE_DT,DEL_FLG,INS_PG,INS_ID,INS_DT,UPD_PG,UPD_ID,UPD_DT)
values
('P001','0.1','ユーザ利用規約','ユーザ利用規約のサンプルです。\n第１条 hogehoge 第２条 fugafuga 2026年5月10日','2026-05-10',0,'admin','admin',now(),'admin','admin',now()),
('P001','0.2','ユーザ利用規約','ユーザ利用規約ver0.2のサンプルです。\n第１条 hogehoge 第２条 fugafuga 2026年7月10日','2026-07-10',0,'admin','admin',now(),'admin','admin',now()),
('P002','0.1','プライバシーポリシー','プライバシーポリシーのサンプルです。\n第１条 hogehoge 第２条 fugafuga 2026年5月10日','2026-05-10',0,'admin','admin',now(),'admin','admin',now()),
('P002','0.2','プライバシーポリシー','プライバシーポリシーver0.2のサンプルです。\n第１条 hogehoge 第２条 fugafuga 2026年7月10日','2026-07-10',0,'admin','admin',now(),'admin','admin',now())
;

insert into m_roles(ID,NAME,DISPLAY_NAME,DISPLAY_DESCRIPTION,DEL_FLG,INS_PG,INS_ID,INS_DT,UPD_PG,UPD_ID,UPD_DT)
values
('RL000010','author','著者','著者 単著書の執筆または共著書の分担執筆を行なった者',0,'admin','admin',now(),'admin','admin',now()),
('RL000020','interpreter','訳者','訳者 外国語書の翻訳を行なった者',0,'admin','admin',now(),'admin','admin',now()),
('RL000030','compiler','編者','編者 編著書の編集責任をもつ者',0,'admin','admin',now(),'admin','admin',now()),
('RL000031','compiler_author','編著者','編著者 編著書の編集と分担執筆をした者',0,'admin','admin',now(),'admin','admin',now())
;

insert into m_record_types(ID,NAME,DISPLAY_NAME,MARC21_TYPE_ID,MARC21_RECORD_LEVEL_ID,DEL_FLG,INS_PG,INS_ID,INS_DT,UPD_PG,UPD_ID,UPD_DT)
values
('RT000001','authored_book','単著書',null,null,0,'admin','admin',now(),'admin','admin',now()),
('RT000011','journal','雑誌（学会誌）',null,null,0,'admin','admin',now(),'admin','admin',now()),
('RT000012','journal_article','雑誌記事/論文',null,null,0,'admin','admin',now(),'admin','admin',now()),
('RT000021','contributed_book','編著書',null,null,0,'admin','admin',now(),'admin','admin',now()),
('RT000022','contributed_book_article','編著書記事/論文',null,null,0,'admin','admin',now(),'admin','admin',now())
;

insert into m_record_fields(ID,NAME,DISPLAY_NAME,DEL_FLG,INS_PG,INS_ID,INS_DT,UPD_PG,UPD_ID,UPD_DT)
values
('RF000001','isbn','ISBN',0,'admin','admin',now(),'admin','admin',now()),
('RF000011','edition','版',0,'admin','admin',now(),'admin','admin',now()),
('RF000021','original_title','原題',0,'admin','admin',now(),'admin','admin',now())
;

