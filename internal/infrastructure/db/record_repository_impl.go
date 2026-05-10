package db

// 書誌情報Repository実装

import "database/sql"

type RecordRepositoryImpl struct {
	db *sql.DB
}

func NewRecordRepositoryImpl(db *sql.DB) *RecordRepositoryImpl {
	return &RecordRepositoryImpl{
		db: db,
	}
}
