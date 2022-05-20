package util

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestStriper(t *testing.T) {
	fn := func(inStr string, expStr *string, msg string) {
		out := Striper(inStr)
		if out == nil {
			assert.Equal(t, expStr, out, msg)
			return
		}
		assert.Equal(t, *expStr, *out, msg)

	}
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "regular", input: "test", expected: "test"},
		{name: "empty", input: "", expected: ""},
		{name: "white spaced afront", input: "     test", expected: "test"},
		{name: "white spaced behind", input: "test     ", expected: "test"},
		{name: "between", input: "te   st", expected: "te   st"},
	}
	for _, test := range tests {
		if test.expected == "" {
			fn(test.input, nil, test.name)
			continue
		}
		fn(test.input, &test.expected, test.name)
	}
}

func TestPassworders(t *testing.T) {
	pass := []string{"pass1", "  pass2", "23214124", "%&^&+^&", ""}
	u := &utl{}

	for _, p := range pass {
		hashed, _ := u.HashPassword(p)
		checked := u.CheckPasswordHash(p, hashed)

		assert.True(t, checked)
	}

}
