package tools

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTools(t *testing.T) {
	var id = 1

	enc := Encode(id)
	if len(enc) == 0 {
		t.Error("error with encoding")
		return
	}

	dec, err := Decode(enc)

	assert.NoError(t, err)
	assert.Equal(t, id, dec)
}
