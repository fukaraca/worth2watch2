package util

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCheckPasswordHash(t *testing.T) {
	ret, err := HashPassword("password")
	assert.Nil(t, err)
	ok := CheckPasswordHash("password", ret)
	if !ok {
		t.Fail()
	}
}
