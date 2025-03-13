package ptr

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOf(t *testing.T) {
	t.Parallel()
	str := "hi"
	num := 123

	assert.Equal(t, &str, Of("hi"))
	assert.Equal(t, &num, Of(123))
}
