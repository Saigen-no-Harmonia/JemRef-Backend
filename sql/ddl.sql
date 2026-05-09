-- ============================================================
-- JemRef DB DDL
-- ENGINE=InnoDB DEFAULT CHARSET=utf8mb4
-- TIMEZONE: Asia/Tokyo (GMT+9)
-- ============================================================

-- ============================================================
-- マスタテーブル
-- ============================================================

CREATE TABLE m_users (
  ID           BIGINT        NOT NULL          AUTO_INCREMENT COMMENT 'ユーザID',
  PUBLIC_ID    CHAR(26)      NOT NULL                         COMMENT '公開用ユーザID',
  FIREBASE_ID  VARCHAR(128)  NOT NULL                         COMMENT 'FirebaseユーザID',
  EMAIL        VARCHAR(254)  NOT NULL                         COMMENT 'メールアドレス',
  TERMS_AGREED_DT          DATETIME(6) NOT NULL               COMMENT '利用規約同意日時',
  TERMS_VERSION            VARCHAR(20) NOT NULL               COMMENT '利用規約同意バージョン',
  PRIVACY_POLICY_AGREED_DT DATETIME(6) NOT NULL               COMMENT 'プライバシーポリシー同意日時',
  PRIVACY_POLICY_VERSION   VARCHAR(20) NOT NULL               COMMENT 'プライバシーポリシー同意バージョン',
  DEL_FLG      TINYINT     NOT NULL DEFAULT 0                 COMMENT '削除フラグ',
  INS_PG       VARCHAR(64) NOT NULL                           COMMENT '作成プログラム',
  INS_ID       VARCHAR(64) NOT NULL                           COMMENT '作成者ID',
  INS_DT       DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) COMMENT '作成日時',
  UPD_PG       VARCHAR(64) NOT NULL                           COMMENT '更新プログラム',
  UPD_ID       VARCHAR(64) NOT NULL                           COMMENT '更新者ID',
  UPD_DT       DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6) COMMENT '更新日時',
  PRIMARY KEY (ID),
  UNIQUE KEY uk_public_id (PUBLIC_ID),  
  UNIQUE KEY uk_firebase_id (FIREBASE_ID),
  UNIQUE KEY uk_email (EMAIL, DEL_FLG)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='ユーザマスタ';

-- ------------------------------------------------------------

CREATE TABLE m_terms (
  ID           CHAR(4)       NOT NULL COMMENT '規約ID (TM01 等)',
  VERSION      VARCHAR(20)   NOT NULL COMMENT 'バージョン',
  NAME         VARCHAR(255)  NOT NULL COMMENT '規約名',
  CONTENT      TEXT          NOT NULL COMMENT '内容',
  EFFECTIVE_DT DATE          NOT NULL COMMENT '発効日',
  DEL_FLG      TINYINT       NOT NULL DEFAULT 0,
  INS_PG       VARCHAR(64)   NOT NULL,
  INS_ID       VARCHAR(64)   NOT NULL,
  INS_DT       DATETIME(6)   NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  UPD_PG       VARCHAR(64)   NOT NULL,
  UPD_ID       VARCHAR(64)   NOT NULL,
  UPD_DT       DATETIME(6)   NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  PRIMARY KEY (ID, VERSION)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='規約マスタ';

-- ------------------------------------------------------------

CREATE TABLE m_records (
  ID                    CHAR(36) NOT NULL COMMENT '書誌マスタID',
  EXTERNAL_KEY_JSON_MAP TEXT     NULL     COMMENT '外部連携用IDのJSONマップ (Ph0未使用)',
  DEL_FLG  TINYINT        NOT NULL DEFAULT 0,
  INS_PG   VARCHAR(64)    NOT NULL,
  INS_ID   VARCHAR(64)    NOT NULL,
  INS_DT   DATETIME(6)    NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  UPD_PG   VARCHAR(64)    NOT NULL,
  UPD_ID   VARCHAR(64)    NOT NULL,
  UPD_DT   DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  PRIMARY KEY (ID)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='書誌マスタ';

-- ------------------------------------------------------------

CREATE TABLE m_contributors (
  ID                  CHAR(64)     NOT NULL COMMENT '貢献者ID',
  NAME                VARCHAR(1024) NOT NULL COMMENT '名前',
  NAME_KANA           VARCHAR(1024) NULL     COMMENT '名前読み (区切り: ", ")',
  BIRTH_DATE          DATE          NULL     COMMENT '生年 (Ph0未使用)',
  DEATH_DATE          DATE          NULL     COMMENT '没年 (Ph0未使用)',
  NII_AUTHOR_ID       CHAR(64)      NULL     COMMENT 'NII著者ID (Ph0未使用)',
  NACSIS_CAT_ID       CHAR(64)      NULL     COMMENT '著者名典拠ID (Ph0未使用)',
  KAKEN_ID            CHAR(64)      NULL     COMMENT '研究者番号 (Ph0未使用)',
  SAME_NAME_EXIST_FLG TINYINT        NOT NULL DEFAULT 0 COMMENT '同名者存在フラグ',
  DEL_FLG  TINYINT      NOT NULL DEFAULT 0,
  INS_PG   VARCHAR(64)    NOT NULL,
  INS_ID   VARCHAR(64)    NOT NULL,
  INS_DT   DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  UPD_PG   VARCHAR(64)    NOT NULL,
  UPD_ID   VARCHAR(64)    NOT NULL,
  UPD_DT   DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  PRIMARY KEY (ID)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='貢献者マスタ';

-- ------------------------------------------------------------

CREATE TABLE m_roles (
  ID                  CHAR(8)       NOT NULL COMMENT '役割ID (RL000001 等)',
  NAME                VARCHAR(64) NOT NULL COMMENT '役割名 (author, compiler 等)',
  DISPLAY_NAME        VARCHAR(1024) NOT NULL COMMENT '役割表示名 (著者, 編者 等)',
  DISPLAY_DESCRIPTION VARCHAR(1024) NULL     COMMENT '表示説明',
  DEL_FLG  TINYINT      NOT NULL DEFAULT 0,
  INS_PG   VARCHAR(64)    NOT NULL,
  INS_ID   VARCHAR(64)    NOT NULL,
  INS_DT   DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  UPD_PG   VARCHAR(64)    NOT NULL,
  UPD_ID   VARCHAR(64)    NOT NULL,
  UPD_DT   DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  PRIMARY KEY (ID),
  UNIQUE KEY uk_name (NAME)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='役割マスタ';

-- ------------------------------------------------------------

CREATE TABLE m_record_types (
  ID                    CHAR(8)  NOT NULL COMMENT '書誌形態ID (BK000001 等)',
  NAME                  CHAR(64) NOT NULL COMMENT '書誌形態名称 (monograph 等)',
  DISPLAY_NAME          CHAR(64) NOT NULL COMMENT '表示名 (著書, 編著 等)',
  MARC21_TYPE_ID        CHAR(64) NULL     COMMENT 'MARC21レコード種別 (将来用)',
  MARC21_RECORD_LEVEL_ID CHAR(64) NULL    COMMENT 'MARC21書誌レベル (将来用)',
  DEL_FLG  TINYINT      NOT NULL DEFAULT 0,
  INS_PG   VARCHAR(64)    NOT NULL,
  INS_ID   VARCHAR(64)    NOT NULL,
  INS_DT   DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  UPD_PG   VARCHAR(64)    NOT NULL,
  UPD_ID   VARCHAR(64)    NOT NULL,
  UPD_DT   DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  PRIMARY KEY (ID),
  UNIQUE KEY uk_name (NAME),
  UNIQUE KEY uk_display_name (DISPLAY_NAME)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='書誌形態マスタ';

-- ------------------------------------------------------------

CREATE TABLE m_record_fields (
  ID           CHAR(8)  NOT NULL COMMENT '書誌入力項目ID (RF000001 等)',
  NAME         CHAR(64) NOT NULL COMMENT '項目名 (データ管理用)',
  DISPLAY_NAME CHAR(64) NOT NULL COMMENT '表示名 (日本語)',
  DEL_FLG  TINYINT      NOT NULL DEFAULT 0,
  INS_PG   VARCHAR(64)    NOT NULL,
  INS_ID   VARCHAR(64)    NOT NULL,
  INS_DT   DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  UPD_PG   VARCHAR(64)    NOT NULL,
  UPD_ID   VARCHAR(64)    NOT NULL,
  UPD_DT   DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  PRIMARY KEY (ID),
  UNIQUE KEY uk_name (NAME),
  UNIQUE KEY uk_display_name (DISPLAY_NAME)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='書誌入力項目マスタ';

-- ------------------------------------------------------------

CREATE TABLE m_record_contributions (
  ID               BIGINT   NOT NULL AUTO_INCREMENT COMMENT '書誌貢献ID',
  M_RECORD_ID      CHAR(36) NOT NULL               COMMENT '書誌ID',
  M_CONTRIBUTOR_ID CHAR(64) NOT NULL               COMMENT '貢献者ID',
  ROLE_ID          CHAR(8)  NOT NULL               COMMENT '役割ID',
  DISPLAY_NUMBER   TINYINT  NOT NULL               COMMENT '表示順',
  DEL_FLG  TINYINT      NOT NULL DEFAULT 0,
  INS_PG   VARCHAR(64)    NOT NULL,
  INS_ID   VARCHAR(64)    NOT NULL,
  INS_DT   DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  UPD_PG   VARCHAR(64)    NOT NULL,
  UPD_ID   VARCHAR(64)    NOT NULL,
  UPD_DT   DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  PRIMARY KEY (ID),
  CONSTRAINT fk_mrc_record      FOREIGN KEY (M_RECORD_ID)      REFERENCES m_records      (ID),
  CONSTRAINT fk_mrc_contributor FOREIGN KEY (M_CONTRIBUTOR_ID) REFERENCES m_contributors (ID),
  CONSTRAINT fk_mrc_role        FOREIGN KEY (ROLE_ID)          REFERENCES m_roles        (ID)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='書誌貢献マスタ';

-- ------------------------------------------------------------

CREATE TABLE m_record_field_values (
  ID              BIGINT        NOT NULL AUTO_INCREMENT COMMENT '書誌項目値ID',
  M_RECORD_ID     CHAR(36)      NOT NULL               COMMENT '書誌ID',
  RECORD_FIELD_ID CHAR(8)       NOT NULL               COMMENT '入力項目ID',
  DISPLAY_NUMBER  TINYINT       NOT NULL DEFAULT 1     COMMENT '表示順 (複数値の場合 2 以上)',
  VALUE           VARCHAR(1024) NOT NULL               COMMENT '入力値 (テキスト。型変換はアプリ側)',
  DEL_FLG  TINYINT      NOT NULL DEFAULT 0,
  INS_PG   VARCHAR(64)    NOT NULL,
  INS_ID   VARCHAR(64)    NOT NULL,
  INS_DT   DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  UPD_PG   VARCHAR(64)    NOT NULL,
  UPD_ID   VARCHAR(64)    NOT NULL,
  UPD_DT   DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  PRIMARY KEY (ID),
  CONSTRAINT fk_mrfv_record FOREIGN KEY (M_RECORD_ID)     REFERENCES m_records       (ID),
  CONSTRAINT fk_mrfv_field  FOREIGN KEY (RECORD_FIELD_ID) REFERENCES m_record_fields (ID)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='書誌項目値マスタ';

-- ============================================================
-- エントリテーブル
-- ============================================================

CREATE TABLE e_records (
  ID               CHAR(36)      NOT NULL               COMMENT 'ユーザ書誌ID',
  USER_ID          BIGINT        NOT NULL               COMMENT '内部用ユーザID',
  TYPE_ID          CHAR(8)       NOT NULL               COMMENT '書誌形態ID',
  MAIN_TITLE       VARCHAR(1024) NOT NULL               COMMENT '主題',
  SUB_TITLE        VARCHAR(1024) NULL                   COMMENT '副題',
  PAGE_RANGE       VARCHAR(64)   NULL                   COMMENT '収録ページ (pp.i-i, 12-13 等)',
  VOLUME           VARCHAR(20)   NULL                   COMMENT '巻',
  ISSUE            VARCHAR(20)   NULL                   COMMENT '号',
  PUBLISHER        VARCHAR(1024) NULL                   COMMENT '出版社',
  PUBLICATION_DATE DATE          NULL                   COMMENT '出版年月日',
  READ_STATUS      TINYINT       NULL                   COMMENT '読了ステータス: 0=未読 1=既読 2=一部既読',
  PARENT_ID        CHAR(36)      NULL                   COMMENT 'コンテナ書誌ID (循環参照はサーバ側でチェック)',
  M_RECORD_ID      CHAR(36)      NULL                   COMMENT '書誌マスタID (突合後に登録。Ph0未使用)',
  DEL_FLG  TINYINT        NOT NULL DEFAULT 0,
  INS_PG   VARCHAR(64)    NOT NULL,
  INS_ID   VARCHAR(64)    NOT NULL,
  INS_DT   DATETIME(6)    NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  UPD_PG   VARCHAR(64)    NOT NULL,
  UPD_ID   VARCHAR(64)    NOT NULL,
  UPD_DT   DATETIME(6)    NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  PRIMARY KEY (ID),
  CONSTRAINT fk_er_user   FOREIGN KEY (USER_ID)     REFERENCES m_users        (ID),
  CONSTRAINT fk_er_type   FOREIGN KEY (TYPE_ID)     REFERENCES m_record_types (ID),
  CONSTRAINT fk_er_master FOREIGN KEY (M_RECORD_ID) REFERENCES m_records      (ID),
  CONSTRAINT fk_er_parent FOREIGN KEY (PARENT_ID)   REFERENCES e_records      (ID),
  INDEX idx_user      (USER_ID, DEL_FLG),
  INDEX idx_user_type (USER_ID, TYPE_ID, DEL_FLG)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='ユーザ入力書誌エントリ';

-- ------------------------------------------------------------

CREATE TABLE e_contributors (
  ID               BIGINT        NOT NULL AUTO_INCREMENT COMMENT '貢献者ID',
  RAW_NAME         VARCHAR(1024) NOT NULL               COMMENT '名前 (ユーザ入力)',
  USER_ID          BIGINT     NOT NULL               COMMENT '内部用ユーザID',
  M_CONTRIBUTOR_ID CHAR(64)      NULL                   COMMENT 'マスタ貢献者ID (突合後に登録)',
  DEL_FLG  TINYINT      NOT NULL DEFAULT 0,
  INS_PG   VARCHAR(64)    NOT NULL,
  INS_ID   VARCHAR(64)    NOT NULL,
  INS_DT   DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  UPD_PG   VARCHAR(64)    NOT NULL,
  UPD_ID   VARCHAR(64)    NOT NULL,
  UPD_DT   DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  PRIMARY KEY (ID),
  CONSTRAINT fk_ec_user   FOREIGN KEY (USER_ID)          REFERENCES m_users       (ID),
  CONSTRAINT fk_ec_master FOREIGN KEY (M_CONTRIBUTOR_ID) REFERENCES m_contributors(ID),
  INDEX idx_raw_name (RAW_NAME(100), DEL_FLG)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='ユーザ入力貢献者エントリ';

-- ------------------------------------------------------------

CREATE TABLE e_record_contributions (
  ID              BIGINT   NOT NULL AUTO_INCREMENT COMMENT 'ユーザ入力書誌貢献ID',
  E_RECORD_ID     CHAR(36) NOT NULL               COMMENT 'ユーザ入力書誌ID',
  E_CONTRIBUTOR_ID BIGINT  NOT NULL               COMMENT 'ユーザ入力貢献者ID',
  ROLE_ID         CHAR(8)  NOT NULL               COMMENT '役割ID',
  DISPLAY_NUMBER  TINYINT  NOT NULL               COMMENT '表示順',
  DEL_FLG      TINYINT        NOT NULL DEFAULT 0,
  INS_PG   VARCHAR(64)    NOT NULL,
  INS_ID   VARCHAR(64)    NOT NULL,
  INS_DT   DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  UPD_PG   VARCHAR(64)    NOT NULL,
  UPD_ID   VARCHAR(64)    NOT NULL,
  UPD_DT   DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  PRIMARY KEY (ID),
  UNIQUE KEY uk_contribution (E_RECORD_ID, E_CONTRIBUTOR_ID, ROLE_ID, DISPLAY_NUMBER),
  CONSTRAINT fk_erc_record      FOREIGN KEY (E_RECORD_ID)      REFERENCES e_records     (ID),
  CONSTRAINT fk_erc_contributor FOREIGN KEY (E_CONTRIBUTOR_ID) REFERENCES e_contributors(ID),
  CONSTRAINT fk_erc_role        FOREIGN KEY (ROLE_ID)          REFERENCES m_roles       (ID)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='ユーザ入力書誌貢献エントリ';

-- ------------------------------------------------------------

CREATE TABLE e_record_urls (
  ID             BIGINT        NOT NULL AUTO_INCREMENT COMMENT 'ユーザ入力URL ID',
  E_RECORD_ID    CHAR(36)      NOT NULL               COMMENT 'ユーザ書誌ID',
  DISPLAY_NUMBER TINYINT       NOT NULL               COMMENT '表示順',
  RAW_LABEL      VARCHAR(1024)    NULL                   COMMENT 'URLラベル名称',
  RAW_URL        VARCHAR(1024) NOT NULL               COMMENT 'URL (ローカルパスも可)',
  USER_ID        BIGINT NOT NULL               COMMENT '内部用ユーザID',
  DEL_FLG      TINYINT        NOT NULL DEFAULT 0,
  INS_PG   VARCHAR(64)    NOT NULL,
  INS_ID   VARCHAR(64)    NOT NULL,
  INS_DT   DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  UPD_PG   VARCHAR(64)    NOT NULL,
  UPD_ID   VARCHAR(64)    NOT NULL,
  UPD_DT   DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  PRIMARY KEY (ID),
  UNIQUE KEY uk_url_order (E_RECORD_ID, DISPLAY_NUMBER),
  CONSTRAINT fk_eru_record FOREIGN KEY (E_RECORD_ID) REFERENCES e_records (ID),
  CONSTRAINT fk_eru_user FOREIGN KEY (USER_ID) REFERENCES m_users (ID)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='ユーザ入力書誌情報URLエントリ';

-- ------------------------------------------------------------

CREATE TABLE e_record_memos (
  ID             BIGINT   NOT NULL AUTO_INCREMENT COMMENT 'ユーザ入力メモID',
  E_RECORD_ID    CHAR(36) NOT NULL               COMMENT 'ユーザ入力書誌ID',
  DISPLAY_NUMBER TINYINT  NOT NULL               COMMENT '表示順',
  TITLE          VARCHAR(1024) NULL              COMMENT 'タイトル',
  CONTENT        TEXT     NOT NULL               COMMENT '内容',
  USER_ID        BIGINT NOT NULL               COMMENT '内部用ユーザID',
  SHARE_LEVEL    TINYINT  NOT NULL DEFAULT 0     COMMENT '公開許諾ステータス: 0=非公開 (Ph0固定)',
  DEL_FLG  TINYINT      NOT NULL DEFAULT 0,
  INS_PG   VARCHAR(64)    NOT NULL,
  INS_ID   VARCHAR(64)    NOT NULL,
  INS_DT   DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  UPD_PG   VARCHAR(64)    NOT NULL,
  UPD_ID   VARCHAR(64)    NOT NULL,
  UPD_DT   DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  PRIMARY KEY (ID),
  UNIQUE KEY uk_memo_order (E_RECORD_ID, DISPLAY_NUMBER),
  CONSTRAINT fk_erm_record FOREIGN KEY (E_RECORD_ID) REFERENCES e_records (ID),
  CONSTRAINT fk_erm_user   FOREIGN KEY (USER_ID)     REFERENCES m_users   (ID)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='ユーザ入力メモエントリ';

-- ------------------------------------------------------------

CREATE TABLE e_record_field_values (
  ID              BIGINT        NOT NULL AUTO_INCREMENT COMMENT 'ユーザ入力項目値ID',
  E_RECORD_ID     CHAR(36)      NOT NULL               COMMENT 'ユーザ書誌ID',
  RECORD_FIELD_ID CHAR(8)       NOT NULL               COMMENT '入力項目ID',
  DISPLAY_NUMBER  TINYINT       NOT NULL DEFAULT 1     COMMENT '表示順 (複数値の場合 2 以上)',
  RAW_VALUE       VARCHAR(1024) NOT NULL               COMMENT 'ユーザ入力値',
  USER_ID         BIGINT      NOT NULL               COMMENT '内部用ユーザID',
  DEL_FLG  TINYINT      NOT NULL DEFAULT 0,
  INS_PG   VARCHAR(64)    NOT NULL,
  INS_ID   VARCHAR(64)    NOT NULL,
  INS_DT   DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  UPD_PG   VARCHAR(64)    NOT NULL,
  UPD_ID   VARCHAR(64)    NOT NULL,
  UPD_DT   DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  PRIMARY KEY (ID),
  UNIQUE KEY uk_field_value (E_RECORD_ID, RECORD_FIELD_ID, DISPLAY_NUMBER),
  CONSTRAINT fk_erfv_record FOREIGN KEY (E_RECORD_ID)     REFERENCES e_records      (ID),
  CONSTRAINT fk_erfv_field  FOREIGN KEY (RECORD_FIELD_ID) REFERENCES m_record_fields(ID),
  -- 項目値の部分一致検索用 (prefix index)
  INDEX idx_user_field_value (USER_ID, RAW_VALUE(100), DEL_FLG),
  -- 全文検索用 (RAW_VALUE のみ。USER_IDはアプリ側でWHEREに追加)
  FULLTEXT INDEX idx_user_fulltext (RAW_VALUE)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='ユーザ書誌項目入力値エントリ';

