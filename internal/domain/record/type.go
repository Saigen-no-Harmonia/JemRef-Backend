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

var articleToContainerMap = map[RecordType]RecordType{
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

// IsArticle コンテナに収録される書誌タイプかを返却する
//
// 雑誌記事・編著書記事のみtrueとなる。単著書や収録物側(雑誌・編著書)はfalse。
func (t RecordType) IsArticle() bool {
	_, ok := articleToContainerMap[t]
	return ok
}

// GetContainerType コンテナに当たる書誌の書誌タイプを返却する
func (t RecordType) GetContainerType() RecordType {
	container, ok := articleToContainerMap[t]
	if !ok {
		// フロントとの契約上、コンテナを持たない書誌タイプからは呼ばれない前提とする
		panic("invalid type as article")
	}

	return container
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
