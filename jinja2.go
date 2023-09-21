package jinja2

import (
	"fmt"
	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
	"github.com/gobwas/glob"
	"github.com/hashicorp/go-multierror"
	"github.com/kluctl/go-embed-python/embed_util"
	"github.com/kluctl/go-embed-python/python"
	"github.com/kluctl/go-jinja2/internal/data"
	"github.com/kluctl/go-jinja2/python_src"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
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
		parallelism: parallelism,
		globCache:   map[string]interface{}{},
		pj:          make(chan *pythonJinja2Renderer, parallelism),
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

	for _, p := range j2.defaultOptions.pythonPath {
		j2.ep.AddPythonPath(p)
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

func (j *Jinja2) Load(opts ...Jinja2Opt) {
	for _, o := range opts {
		o(&j.defaultOptions)
	}
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

func (j *Jinja2) getGlob(pattern string) (glob.Glob, error) {
	j.mutex.Lock()
	defer j.mutex.Unlock()

	g, ok := j.globCache[pattern]
	if ok {
		if g2, ok := g.(glob.Glob); ok {
			return g2, nil
		} else {
			return nil, g2.(error)
		}
	}
	g, err := glob.Compile(pattern, '/')
	if err != nil {
		j.globCache[pattern] = err
		return nil, err
	}
	j.globCache[pattern] = g
	return g.(glob.Glob), nil
}

func (j *Jinja2) RenderDirectory(sourceDir string, targetDir string, excludePatterns []string, opts ...Jinja2Opt) error {
	var jobs []*RenderJob

	tmpOpts := j.defaultOptions
	for _, o := range opts {
		o(&tmpOpts)
	}

	rootDir := sourceDir
	subdir := ""
	if tmpOpts.templateIgnoreRootPath != "" {
		rootDir = tmpOpts.templateIgnoreRootPath
		abs, err := filepath.Abs(rootDir)
		if err != nil {
			return err
		}
		subdir, err = filepath.Rel(abs, sourceDir)
		if err != nil {
			return err
		}
	}

	ignore, err := j.readAllIgnoreFiles(rootDir, subdir, excludePatterns)
	if err != nil {
		return err
	}

	ignoreMatcher := gitignore.NewMatcher(ignore)

	symlinks := map[string]string{}
	err = filepath.WalkDir(sourceDir, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(sourceDir, p)
		if err != nil {
			return err
		}

		targetPath := filepath.Join(targetDir, relPath)

		if d.Type() == fs.ModeSymlink {
			lnk, err := os.Readlink(p)
			if err != nil {
				return err
			}
			symlinks[targetPath] = lnk
			return nil
		} else if d.IsDir() {
			err = os.MkdirAll(targetPath, 0o700)
			if err != nil {
				return err
			}
			return nil
		}

		pathSlice := strings.Split(filepath.Join(subdir, relPath), string(filepath.Separator))

		if ignoreMatcher.Match(pathSlice, d.IsDir()) {
			if !d.IsDir() {
				b, err := os.ReadFile(p)
				if err != nil {
					return err
				}
				err = os.WriteFile(targetPath, b, 0o600)
				if err != nil {
					return err
				}
			}
			return nil
		}

		job := &RenderJob{
			Template: filepath.ToSlash(p),
			target:   targetPath,
		}
		jobs = append(jobs, job)
		return nil
	})
	if err != nil {
		return err
	}

	err = j.RenderFiles(jobs, opts...)
	if err != nil {
		return err
	}

	var retErr *multierror.Error
	for _, job := range jobs {
		if job.Error != nil {
			retErr = multierror.Append(retErr, fmt.Errorf("failed rendering template '%s': %w", job.Template, job.Error))
			continue
		}

		err = os.WriteFile(job.target, []byte(*job.Result), 0o600)
		if err != nil {
			return err
		}
	}
	if retErr.ErrorOrNil() != nil {
		return retErr
	}

	for n, o := range symlinks {
		err = os.Symlink(o, n)
		if err != nil {
			return err
		}
	}

	return nil
}
