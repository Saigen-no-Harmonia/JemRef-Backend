package repository

import "jemref_go/internal/domain/record"

type RecordRepository interface {
	Create(ur *record.UserRecordEntry) error
}
