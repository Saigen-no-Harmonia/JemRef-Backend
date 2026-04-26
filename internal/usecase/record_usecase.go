package usecase

import (
	"jemref_go/internal/domain"
	"jemref_go/internal/interface/repository"
)

type RecordUsecase struct {
	repo repository.RecordRepository
}

func NewRecordUsecase(r repository.RecordRepository) *RecordUsecase {
	return &RecordUsecase{repo: r}
}

func (r *RecordUsecase) CreateRecord(ci CreateRecordInput) (*CreateRecordOutput, error) {

	// 引数をマッピング
	record := &domain.UserRecordEntry{
		UserId:    ci.UserId,
		MainTitle: ci.MainTitle,
	}

	// DBに登録
	if err := r.repo.Create(record); err != nil {
		return nil, err
	}

	// レスポンス
	return &CreateRecordOutput{UserId: record.UserId}, nil
}
