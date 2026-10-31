package record

type RecordType string

const (
	RecordTypeAuthoredBook           RecordType = "authored_book"
	RecordTypeJournal                RecordType = "journal"
	RecordTypeJournalArticle         RecordType = "journal_article"
	RecordTypeContributedBook        RecordType = "contributed_book"
	RecordTypeContributedBookArticle RecordType = "contributed_book_article"
)

const (
	RecordTypeIdAuthoredBook           = "RT000001"
	RecordTypeIdJournal                = "RT000011"
	RecordTypeIdJournalArticle         = "RT000012"
	RecordTypeIdContributedBook        = "RT000021"
	RecordTypeIdContributedBookArticle = "RT000022"
)

var typeToIdMap = map[RecordType]string{
	RecordTypeAuthoredBook:           RecordTypeIdAuthoredBook,
	RecordTypeJournal:                RecordTypeIdJournal,
	RecordTypeJournalArticle:         RecordTypeIdJournalArticle,
	RecordTypeContributedBook:        RecordTypeIdContributedBook,
	RecordTypeContributedBookArticle: RecordTypeIdContributedBookArticle,
}

var idToTypeMap = map[string]RecordType{
	RecordTypeIdAuthoredBook:           RecordTypeAuthoredBook,
	RecordTypeIdJournal:                RecordTypeJournal,
	RecordTypeIdJournalArticle:         RecordTypeJournalArticle,
	RecordTypeIdContributedBook:        RecordTypeContributedBook,
	RecordTypeIdContributedBookArticle: RecordTypeContributedBookArticle,
}

var childToParentMap = map[RecordType]RecordType{
	RecordTypeJournalArticle:         RecordTypeJournal,
	RecordTypeContributedBookArticle: RecordTypeContributedBook,
}

// GetId 書誌タイプのIDを返却する
func (t RecordType) GetId() string {
	id, ok := typeToIdMap[t]
	if !ok {
		panic("invalid type")
	}

	return id
}

func (t RecordType) HasParent() bool {
	_, ok := childToParentMap[t]
	return ok
}

// GetParent コンテナに当たる書誌の書誌タイプを返却する
func (t RecordType) GetParent() RecordType {
	parent, ok := childToParentMap[t]
	if !ok {
		// フロントとの契約上、親を持たない書誌タイプからは呼ばれない前提とする
		panic("invalid type as child")
	}

	return parent
}

// RecordTypeFromId 書誌タイプIDをもとに書誌タイプを返却する
func RecordTypeFromId(id string) RecordType {
	recordType, ok := idToTypeMap[id]
	if !ok {
		panic("invalid record type id")
	}
	return recordType
}

// IsValid 書誌タイプが正しいものであればtrueを返却する
func (t RecordType) IsValid() bool {
	_, ok := typeToIdMap[t]
	return ok
}
