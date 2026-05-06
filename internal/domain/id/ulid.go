package id

import (
	"math/rand"
	"time"

	"github.com/oklog/ulid"
)

type UlidGenerator struct{}

// コンストラクタ
func NewUlidGenerator() *UlidGenerator {
	return &UlidGenerator{}
}

// ULIDを生成し返却する
func (g *UlidGenerator) Generate() string {

	t := time.Now()
	entropy := ulid.Monotonic(rand.New(rand.NewSource(t.UnixNano())), 0)
	id := ulid.MustNew(ulid.Timestamp(t), entropy)

	return id.String()
}
