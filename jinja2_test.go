package jinja2

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
)

func newJinja2(t *testing.T, opts ...Jinja2Opt) *Jinja2 {
	name := fmt.Sprintf("jinja2-%d", rand.Uint32())
	opts2 := append(opts, WithExtension("go_jinja2.ext.kluctl"))
	j2, err := NewJinja2(name, 1, opts2...)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(j2.Close)
	t.Cleanup(j2.Cleanup)

	return j2
}

func newTemplateFile(t *testing.T, content string) string {
	tmpFile, err := os.CreateTemp("", "test-template-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		tmpFile.Close()
		_ = os.Remove(tmpFile.Name())
	})
	_, err = tmpFile.Write([]byte(content))
	if err != nil {
		t.Fatal(err)
	}
	return tmpFile.Name()
}

func newTemplateDir(t *testing.T, contents map[string]string) string {
	dir := t.TempDir()
	for p, c := range contents {
		x := filepath.Join(dir, p)
		err := os.MkdirAll(filepath.Dir(x), 0o700)
		if err != nil {
			t.Fatal(err)
		}
		err = os.WriteFile(x, []byte(c), 0o600)
		if err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func TestJinja2(t *testing.T) {
	j2 := newJinja2(t, WithGlobals(map[string]any{
		"test_var1": "1",
		"test_var2": map[string]any{
			"test": "2",
		},
	}))

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

func TestRenderFiles(t *testing.T) {
	j2 := newJinja2(t)

	t1 := newTemplateFile(t, `{{ "a" }}`)
	t2 := newTemplateFile(t, `{{ "b" }}`)

	r1, err := j2.RenderFile(t1)
	assert.NoError(t, err)
	assert.Equal(t, "a", r1)

	r2, err := j2.RenderFile(t2)
	assert.NoError(t, err)
	assert.Equal(t, "b", r2)
}

func TestRenderFiles_Includes(t *testing.T) {
	includeDir := newTemplateDir(t, map[string]string{
		"include.yaml":        "test",
		"include-dot.yaml":    `{% include "./include.yaml" %}`,
		"include-caller.yaml": `{% include "include-caller2.yaml" %}`,
	})

	type testCase struct {
		name  string
		files map[string]string
		t     string   // templateNane
		sd    []string // searchDirs
		r     string   // result
		err   string
	}

	tests := []testCase{
		{
			name: "include with absolute file fails",
			files: map[string]string{
				"f.yaml": fmt.Sprintf(`{%% include "%s/include.yaml" %%}`, includeDir),
			},
			t:   "f.yaml",
			err: fmt.Sprintf("template %s/include.yaml not found", includeDir),
		},
		{
			name: "load_template with absolute file fails",
			files: map[string]string{
				"f.yaml": fmt.Sprintf(`{{ load_template("%s/include.yaml") }}`, includeDir),
			},
			t:   "f.yaml",
			err: fmt.Sprintf("template %s/include.yaml not found", includeDir),
		},
		{
			name: "include without searchdir fails",
			files: map[string]string{
				"f.yaml": `{% include "include.yaml" %}`,
			},
			t:   "f.yaml",
			err: "template include.yaml not found",
		},
		{
			name: "include with searchdir succeeds",
			files: map[string]string{
				"f.yaml": `{% include "include.yaml" %}`,
			},
			t:  "f.yaml",
			r:  "test",
			sd: []string{includeDir},
		},
		{
			name: "load_template without searchdir fails",
			files: map[string]string{
				"f.yaml": `{{ load_template("include.yaml") }}`,
			},
			t:   "f.yaml",
			err: "template include.yaml not found",
		},
		{
			name: "load_template with searchdir succeeds",
			files: map[string]string{
				"f.yaml": `{{ load_template("include.yaml") }}`,
			},
			t:  "f.yaml",
			r:  "test",
			sd: []string{includeDir},
		},
		{
			name: "relative include fails without dot",
			files: map[string]string{
				"d1/include1.yaml": `{% include "include2.yaml" %}`,
				"d1/include2.yaml": "test2",
				"f.yaml":           `{% include "d1/include1.yaml" %}`,
			},
			t:   "f.yaml",
			sd:  []string{"self"},
			err: "template include2.yaml not found",
		},
		{
			name: "relative include succeeds with dot",
			files: map[string]string{
				"d1/include1.yaml": `{% include "./include2.yaml" %}`,
				"d1/include2.yaml": "test2",
				"f.yaml":           `{% include "d1/include1.yaml" %}`,
			},
			t:  "f.yaml",
			sd: []string{"self"},
			r:  "test2",
		},
		{
			name: "recursive include succeeds",
			files: map[string]string{
				"include2.yaml": `{% include "./include3.yaml" %}`,
				"include3.yaml": "test3",
				"f.yaml":        `{% include "./include2.yaml" %}`,
			},
			t:  "f.yaml",
			sd: []string{"self"},
			r:  "test3",
		},
		{
			name: "recursive include from searchdir succeeds",
			files: map[string]string{
				"f.yaml": `{% include "include-dot.yaml" %}`,
			},
			t:  "f.yaml",
			sd: []string{includeDir},
			r:  "test",
		},
		{
			name: "recursive include caller from searchdir succeeds",
			files: map[string]string{
				"include-caller2.yaml": "test-caller",
				"f.yaml":               `{% include "include-caller.yaml" %}`,
			},
			t:  "f.yaml",
			sd: []string{"self", includeDir},
			r:  "test-caller",
		},
	}

	j2 := newJinja2(t)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dir := newTemplateDir(t, tc.files)
			var sd []string
			for _, x := range tc.sd {
				if x == "self" {
					sd = append(sd, dir)
				} else {
					sd = append(sd, x)
				}
			}
			r, err := j2.RenderFile(filepath.Join(dir, tc.t), WithSearchDirs(sd))
			if tc.err == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.err)
			}
			assert.Equal(t, tc.r, r)
		})
	}
}
