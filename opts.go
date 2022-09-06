package jinja2

type jinja2Options struct {
	Strict       bool `json:"strict"`
	TrimBlocks   bool `json:"trimBlocks"`
	LstripBlocks bool `json:"lstripBlocks"`

	SearchDirs []string       `json:"searchDirs"`
	Globals    map[string]any `json:"globals"`

	Extensions []string `json:"extensions"`
}

type Jinja2Opt func(o *jinja2Options)

func WithStrict(strict bool) Jinja2Opt {
	return func(o *jinja2Options) {
		o.Strict = strict
	}
}

func WithTrimBlocks(trimBlocks bool) Jinja2Opt {
	return func(o *jinja2Options) {
		o.TrimBlocks = trimBlocks
	}
}

func WithLStripBlocks(lstripBlocks bool) Jinja2Opt {
	return func(o *jinja2Options) {
		o.LstripBlocks = lstripBlocks
	}
}

func WithSearchDir(dir string) Jinja2Opt {
	return func(o *jinja2Options) {
		o.SearchDirs = append(o.SearchDirs, dir)
	}
}

func WithSearchDirs(dirs []string) Jinja2Opt {
	return func(o *jinja2Options) {
		o.SearchDirs = append(o.SearchDirs, dirs...)
	}
}

func WithGlobal(k string, v any) Jinja2Opt {
	return func(o *jinja2Options) {
		if o.Globals == nil {
			o.Globals = make(map[string]any)
		}
		o.Globals[k] = v
	}
}

func WithGlobals(globals map[string]any) Jinja2Opt {
	return func(o *jinja2Options) {
		if o.Globals == nil {
			o.Globals = make(map[string]any)
		}
		for k, v := range globals {
			o.Globals[k] = v
		}
	}
}

func WithExtention(e string) Jinja2Opt {
	return func(o *jinja2Options) {
		o.Extensions = append(o.Extensions, e)
	}
}
