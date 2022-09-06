package jinja2

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
)

func TestJinja2(t *testing.T) {
	name := fmt.Sprintf("jinja2-%d", rand.Uint32())
	j2, err := NewJinja2(name, 1, WithGlobal("test_var1", 1), WithGlobal("test_var2", map[string]any{"test": 2}))
	assert.NoError(t, err)
	defer j2.Close()
	defer j2.Cleanup()

	s, err := j2.RenderString("test")
	assert.NoError(t, err)
	assert.Equal(t, "test", s)

	s, err = j2.RenderString("test - {{ test_var1 }}")
	assert.NoError(t, err)
	assert.Equal(t, "test - 1", s)

	s, err = j2.RenderString("test - {{ test_var2.test }}")
	assert.NoError(t, err)
	assert.Equal(t, "test - 2", s)

	s, err = j2.RenderString("test - {{ get_var('test_var2.test', 'd') }}", WithExtention("go_jinja2.ext.kluctl"))
	assert.NoError(t, err)
	assert.Equal(t, "test - 2", s)

	s, err = j2.RenderString("test - {{ get_var('test_var2.test1', 'd') }}", WithExtention("go_jinja2.ext.kluctl"))
	assert.NoError(t, err)
	assert.Equal(t, "test - d", s)
}
