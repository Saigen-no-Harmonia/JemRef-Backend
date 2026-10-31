package record

type RecordField string

const (
	RecordFieldIsbn          RecordField = "isbn"
	RecordFieldEdition       RecordField = "edition"
	RecordFieldOriginalTitle RecordField = "original_title"
)

const (
	RecordFieldIdIsbn          = "RF000001"
	RecordFieldIdEdition       = "RF000011"
	RecordFieldIdOriginalTitle = "RF000021"
)

var fieldToIdMap = map[RecordField]string{
	RecordFieldIsbn:          RecordFieldIdIsbn,
	RecordFieldEdition:       RecordFieldIdEdition,
	RecordFieldOriginalTitle: RecordFieldIdOriginalTitle,
}

var idToFieldMap = map[string]RecordField{
	RecordFieldIdIsbn:          RecordFieldIsbn,
	RecordFieldIdEdition:       RecordFieldEdition,
	RecordFieldIdOriginalTitle: RecordFieldOriginalTitle,
}

func (f RecordField) GetId() string {
	id, ok := fieldToIdMap[f]
	if !ok {
		panic("invalid status")
	}

	return id
}

// RecordFieldFromId 書誌項目IDをもとに書誌項目名を返却する
func RecordFieldFromId(id string) RecordField {
	field, ok := idToFieldMap[id]
	if !ok {
		panic("invalid field id")
	}

	return field
}

// IsValid 書誌項目IDが正しいものであればtrueを返却する
func (f RecordField) IsValid() bool {
	_, ok := fieldToIdMap[f]
	return ok
}
