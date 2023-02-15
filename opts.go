package jinja2

type jinja2Options struct {
	DebugTrace   bool `json:"debugTrace"`
	NonStrict    bool `json:"nonStrict"`
	TrimBlocks   bool `json:"trimBlocks"`
	LstripBlocks bool `json:"lstripBlocks"`

	SearchDirs []string       `json:"searchDirs"`
	Globals    map[string]any `json:"globals"`

	Filters    map[string]string `json:"filters"`
	Extensions []string          `json:"extensions"`

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

// WithFilter adds a custom filter with `name` to the engine
//
// name: the name of the filter
// code: the code defines a filter function
//
// By default, name of the defined function should be same as the filter name.
// You can change this behaviour to set your filter name to 'xxx:yyy' format,
// then the real filter name is 'xxx' and the function name is 'yyy'.
//
// For example, you can use
//
//	WithFilter("add", "def add(x, y): return x + y")
//
// And also, you can use
//
//	WithFilter("add:my_add", "def my_add(x, y): return x + y")
func WithFilter(name string, code string) Jinja2Opt {
	return func(o *jinja2Options) {
		if o.Filters == nil {
			o.Filters = make(map[string]string)
		}

		o.Filters[name] = code
	}
}

// WithExtension adds a custom extension to the engine
//
// You can pass in an import path to an extension (for using some built-in extension, e.g. `jinja2.ext.debug`)
// Or you can pass in the code of an extension (there must be exactly one class inherited from `jinja2.ext.Extension`)
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
