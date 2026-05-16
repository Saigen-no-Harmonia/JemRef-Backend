package id

import (
	"math/rand"
	"time"

	"github.com/oklog/ulid"
)

type UlidGenerator struct {
	now func() time.Time
}

// コンストラクタ
func NewUlidGenerator() *UlidGenerator {
	return &UlidGenerator{
		now: time.Now,
	}
}

// ULIDを生成し返却する
func (g *UlidGenerator) Generate() string {

	t := g.now()
	entropy := ulid.Monotonic(rand.New(rand.NewSource(t.UnixNano())), 0)
	id := ulid.MustNew(ulid.Timestamp(t), entropy)

	return id.String()
}
