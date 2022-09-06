package jinja2

import (
	"github.com/kluctl/go-embed-python/embed_util"
	"github.com/kluctl/go-embed-python/python"
	"github.com/kluctl/go-jinja2/internal/data"
	"github.com/kluctl/go-jinja2/python_src"
	"sync"
)

type Jinja2 struct {
	ep          *python.EmbeddedPython
	jinja2Lib   *embed_util.EmbeddedFiles
	rendererSrc *embed_util.EmbeddedFiles

	parallelism int
	pj          chan *pythonJinja2Renderer
	globCache   map[string]interface{}
	mutex       sync.Mutex

	defaultOptions jinja2Options
}

type RenderJob struct {
	Template string
	Result   *string
	Error    error

	target string
}

type Jinja2Error struct {
	error string
}

func (m *Jinja2Error) Error() string {
	return m.error
}

func NewJinja2(name string, parallelism int, opts ...Jinja2Opt) (*Jinja2, error) {
	var wg sync.WaitGroup
	var mutex sync.Mutex
	var err error

	j2 := &Jinja2{
		globCache: map[string]interface{}{},
		pj:        make(chan *pythonJinja2Renderer, parallelism),
	}

	j2.ep, err = python.NewEmbeddedPython(name)
	if err != nil {
		return nil, err
	}
	j2.jinja2Lib, err = embed_util.NewEmbeddedFiles(data.Data, name)
	if err != nil {
		return nil, err
	}
	j2.ep.AddPythonPath(j2.jinja2Lib.GetExtractedPath())

	j2.rendererSrc, err = embed_util.NewEmbeddedFiles(python_src.RendererSource, name)
	if err != nil {
		return nil, err
	}

	for _, o := range opts {
		o(&j2.defaultOptions)
	}

	for i := 0; i < parallelism; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			pj, err2 := newPythonJinja2Renderer(j2)
			if err2 != nil {
				mutex.Lock()
				defer mutex.Unlock()
				err = err2
				return
			}
			j2.pj <- pj
		}()
	}
	wg.Wait()
	if err != nil {
		return nil, err
	}

	return j2, nil
}

func (j *Jinja2) Close() {
	for i := 0; i < j.parallelism; i++ {
		pj := <-j.pj
		pj.Close()
	}
}

func (j *Jinja2) Cleanup() {
	_ = j.rendererSrc.Cleanup()
	_ = j.jinja2Lib.Cleanup()
	_ = j.ep.Cleanup()
}

func (j *Jinja2) RenderStrings(jobs []*RenderJob, opts ...Jinja2Opt) error {
	pj := <-j.pj
	defer func() { j.pj <- pj }()
	return pj.renderHelper(jobs, true, opts)
}

func (j *Jinja2) RenderString(template string, opts ...Jinja2Opt) (string, error) {
	jobs := []*RenderJob{{
		Template: template,
	}}
	err := j.RenderStrings(jobs, opts...)
	if err != nil {
		return "", err
	}
	if jobs[0].Error != nil {
		return "", jobs[0].Error
	}
	return *jobs[0].Result, nil
}

func (j *Jinja2) RenderFiles(jobs []*RenderJob, opts ...Jinja2Opt) error {
	pj := <-j.pj
	defer func() { j.pj <- pj }()
	return pj.renderHelper(jobs, false, opts)
}

func (j *Jinja2) RenderFile(template string, opts ...Jinja2Opt) (string, error) {
	jobs := []*RenderJob{{
		Template: template,
	}}
	err := j.RenderFiles(jobs, opts...)
	if err != nil {
		return "", err
	}
	if jobs[0].Error != nil {
		return "", jobs[0].Error
	}
	return *jobs[0].Result, nil
}
