package repository

import "jemref_go/internal/domain"

type RecordRepository interface {
	Create(b *domain.UserRecordEntry) error
}
