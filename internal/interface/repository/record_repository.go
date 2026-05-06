package repository

import "jemref_go/internal/domain"

type RecordRepository interface {
	Create(ur *domain.UserRecordEntry) error
}
