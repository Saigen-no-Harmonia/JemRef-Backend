package id

import (
	"testing"

	"github.com/oklog/ulid"
	"github.com/stretchr/testify/assert"
)

func TestUlidGenerator_Generate(t *testing.T) {
	g := NewUlidGenerator()
	id := g.Generate()

	assert.NotEmpty(t, id)

	_, err := ulid.Parse(id)
	assert.NoError(t, err)
}

func TestUlidGenerator_Generate_Unique(t *testing.T) {
	g := NewUlidGenerator()
	id1 := g.Generate()
	id2 := g.Generate()

	assert.NotEqual(t, id1, id2)
}
