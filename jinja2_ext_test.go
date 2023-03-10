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
