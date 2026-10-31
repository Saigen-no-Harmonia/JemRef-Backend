package repository

import (
	"context"
	"jemref_go/internal/domain/record"
)

type RecordRepository interface {
	Create(
		ctx context.Context,
		tx DBTX,
		ur record.UserRecordEntry,
	) error
}
