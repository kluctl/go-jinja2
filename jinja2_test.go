package jinja2

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
)

func newJinja2(t *testing.T) *Jinja2 {
	name := fmt.Sprintf("jinja2-%d", rand.Uint32())
	j2, err := NewJinja2(name, 1, WithGlobal("test_var1", 1), WithGlobal("test_var2", map[string]any{"test": 2}))
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(j2.Close)
	t.Cleanup(j2.Cleanup)

	return j2
}
func TestJinja2(t *testing.T) {
	j2 := newJinja2(t)

	s, err := j2.RenderString("test")
	assert.NoError(t, err)
	assert.Equal(t, "test", s)

	s, err = j2.RenderString("test - {{ test_var1 }}")
	assert.NoError(t, err)
	assert.Equal(t, "test - 1", s)

	s, err = j2.RenderString("test - {{ test_var2.test }}")
	assert.NoError(t, err)
	assert.Equal(t, "test - 2", s)

	s, err = j2.RenderString("test - {{ get_var('test_var2.test', 'd') }}", WithExtension("go_jinja2.ext.kluctl"))
	assert.NoError(t, err)
	assert.Equal(t, "test - 2", s)

	s, err = j2.RenderString("test - {{ get_var('test_var2.test1', 'd') }}", WithExtension("go_jinja2.ext.kluctl"))
	assert.NoError(t, err)
	assert.Equal(t, "test - d", s)
}

type testStruct struct {
	V1 string         `json:"v1"`
	S1 testStruct2    `json:"s1"`
	M1 map[string]any `json:"m1"`
}

type testStruct2 struct {
	V2 string `json:"v2"`
}

func TestRenderStruct(t *testing.T) {
	j2 := newJinja2(t)

	s := testStruct{
		V1: `{{ "1" }}`,
		S1: testStruct2{
			V2: `{{ "2" }}`,
		},
		M1: map[string]any{
			"a": map[string]any{
				"b": `{{ "3" }}`,
			},
			`{{ "c" }}`: "4",
			`{{ "d" }}`: `{{ "5" }}`,
			`{{ "e" }}`: map[string]any{
				"f":         `{{ "6" }}`,
				`{{ "g" }}`: `{{ "7" }}`,
			},
		},
	}
	e := testStruct{
		V1: "1",
		S1: testStruct2{
			V2: "2",
		},
		M1: map[string]any{
			"a": map[string]any{
				"b": "3",
			},
			"c": "4",
			"d": "5",
			"e": map[string]any{
				"f": "6",
				"g": "7",
			},
		},
	}

	changed, err := j2.RenderStruct(&s)
	assert.NoError(t, err)
	assert.True(t, changed)
	assert.Equal(t, e, s)

	changed, err = j2.RenderStruct(&s)
	assert.NoError(t, err)
	assert.False(t, changed)
	assert.Equal(t, e, s)
}
