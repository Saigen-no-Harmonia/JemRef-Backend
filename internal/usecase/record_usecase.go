package usecase

import (
	"jemref_go/internal/domain/record"
	"jemref_go/internal/repository"
	"jemref_go/internal/usecase/dto"
)

type RecordUsecase struct {
	repo repository.RecordRepository
	// txManager repository.TxManager
}

func NewRecordUsecase(r repository.RecordRepository) *RecordUsecase {
	return &RecordUsecase{repo: r}
}

func (r *RecordUsecase) CreateRecord(ci dto.CreateRecordInput) (*dto.CreateRecordOutput, error) {

	// 引数をマッピング
	record := &record.UserRecordEntry{
		PublicUserId: ci.UserId,
		MainTitle:    ci.MainTitle,
	}

	// DBに登録
	err := r.repo.Create(record)

	if err != nil {
		return nil, err
	}

	// レスポンス
	return &dto.CreateRecordOutput{UserId: record.PublicUserId}, nil
}
