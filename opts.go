package jinja2

type jinja2Options struct {
	DebugTrace   bool `json:"debugTrace"`
	NonStrict    bool `json:"nonStrict"`
	TrimBlocks   bool `json:"trimBlocks"`
	LstripBlocks bool `json:"lstripBlocks"`

	SearchDirs []string       `json:"searchDirs"`
	Globals    map[string]any `json:"globals"`

	Extensions []string `json:"extensions"`

	// not passed to renderer
	pythonPath       []string
	traceJsonSend    func(map[string]any)
	traceJsonReceive func(map[string]any)
}

type Jinja2Opt func(o *jinja2Options)

func WithDebugTrace(debugTrace bool) Jinja2Opt {
	return func(o *jinja2Options) {
		o.DebugTrace = debugTrace
	}
}

func WithPythonPath(p string) Jinja2Opt {
	return func(o *jinja2Options) {
		o.pythonPath = append(o.pythonPath, p)
	}
}

func WithStrict(strict bool) Jinja2Opt {
	return func(o *jinja2Options) {
		o.NonStrict = !strict
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

func WithExtension(e string) Jinja2Opt {
	return func(o *jinja2Options) {
		o.Extensions = append(o.Extensions, e)
	}
}

func WithTraceJsonSend(f func(map[string]any)) Jinja2Opt {
	return func(o *jinja2Options) {
		o.traceJsonSend = f
	}
}

func WithTraceJsonReceive(f func(map[string]any)) Jinja2Opt {
	return func(o *jinja2Options) {
		o.traceJsonReceive = f
	}
}
