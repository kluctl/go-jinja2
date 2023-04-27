package jinja2

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestToYaml(t *testing.T) {
	type testCase struct {
		v      any
		result string
	}

	testCases := []testCase{
		{v: "a", result: "v: a"},
		{v: 1, result: "v: 1"},
		{v: "1", result: "v: '1'"},
		{v: "01", result: "v: '01'"},
		{v: "09", result: "v: '09'"},
	}
	j2 := newJinja2(t)

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%v-%s", tc.v, tc.result), func(t *testing.T) {
			s, err := j2.RenderString("{{ m | to_yaml }}", WithGlobal("m", map[string]any{"v": tc.v}))
			assert.NoError(t, err)
			assert.Equal(t, tc.result+"\n", s)
		})
	}
}

func TestSha256(t *testing.T) {
	j2 := newJinja2(t)
	s, err := j2.RenderString("{{ 'test' | sha256 }}")
	assert.NoError(t, err)
	assert.Equal(t, "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08", s)
}

func TestSha256PrefixLength(t *testing.T) {
	j2 := newJinja2(t)
	s, err := j2.RenderString("{{ 'test' | sha256(6) }}")
	assert.NoError(t, err)
	assert.Equal(t, "9f86d0", s)
}
