package jinja2

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/jinzhu/copier"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type pythonJinja2Renderer struct {
	j2 *Jinja2

	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser

	stdoutReader *bufio.Reader
}

func newPythonJinja2Renderer(j2 *Jinja2) (*pythonJinja2Renderer, error) {
	isOk := false
	j2r := &pythonJinja2Renderer{
		j2: j2,
	}
	defer func() {
		if !isOk {
			j2r.Close()
		}
	}()

	args := []string{filepath.Join(j2.rendererSrc.GetExtractedPath(), "main.py")}
	j2r.cmd = j2.ep.PythonCmd(args...)
	j2r.cmd.Stderr = os.Stderr

	stdout, err := j2r.cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	j2r.stdout = stdout

	stdin, err := j2r.cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	j2r.stdin = stdin

	err = j2r.cmd.Start()
	if err != nil {
		return nil, err
	}

	j2r.stdoutReader = bufio.NewReader(j2r.stdout)

	isOk = true

	return j2r, nil
}

func (j *pythonJinja2Renderer) Close() {
	if j.stdin != nil {
		args := jinja2Args{Cmd: "exit"}
		_ = json.NewEncoder(j.stdin).Encode(args)

		_ = j.stdin.Close()
		j.stdin = nil
	}
	if j.stdout != nil {
		_ = j.stdout.Close()
		j.stdout = nil
	}
	if j.cmd != nil {
		if j.cmd.Process != nil {
			timer := time.AfterFunc(5*time.Second, func() {
				_ = j.cmd.Process.Kill()
			})
			_ = j.cmd.Wait()
			timer.Stop()
		}
		j.cmd = nil
	}
}

func isMaybeTemplateString(template string) bool {
	return strings.IndexRune(template, '{') != -1
}

func isMaybeTemplateBytes(template []byte) bool {
	return bytes.IndexRune(template, '{') != -1
}

func isMaybeTemplate(template string, searchDirs []string, isString bool) (bool, *string) {
	if isString {
		if !isMaybeTemplateString(template) {
			return false, &template
		}
	} else {
		for _, s := range searchDirs {
			b, err := ioutil.ReadFile(filepath.Join(s, template))
			if err != nil {
				continue
			}
			if !isMaybeTemplateBytes(b) {
				x := string(b)
				return false, &x
			} else {
				return true, nil
			}
		}
	}
	return true, nil
}

type jinja2Args struct {
	Cmd       string   `json:"cmd"`
	Templates []string `json:"templates"`

	Opts *jinja2Options `json:"opts"`
}

type jinja2Result struct {
	Result *string `json:"result,omitempty"`
	Error  *string `json:"error,omitempty"`
}

func (j *pythonJinja2Renderer) renderHelper(jobs []*RenderJob, isString bool, opts []Jinja2Opt) error {
	var jargs jinja2Args
	if isString {
		jargs.Cmd = "render-strings"
	} else {
		jargs.Cmd = "render-files"
	}

	jargs.Opts = &jinja2Options{}
	err := copier.CopyWithOption(jargs.Opts, &j.j2.defaultOptions, copier.Option{
		DeepCopy: true,
	})
	if err != nil {
		return err
	}

	for _, o := range opts {
		o(jargs.Opts)
	}

	var processedJobs []*RenderJob

	for _, job := range jobs {
		if ist, r := isMaybeTemplate(job.Template, jargs.Opts.SearchDirs, isString); !ist {
			job.Result = r
			continue
		}
		processedJobs = append(processedJobs, job)
		jargs.Templates = append(jargs.Templates, job.Template)
	}
	if len(processedJobs) == 0 {
		return nil
	}

	b, err := json.Marshal(jargs)
	if err != nil {
		j.Close()
		return err
	}
	b = append(b, '\n')

	if jargs.Opts.traceJsonSend != nil {
		var m map[string]any
		_ = json.Unmarshal(b, &m)
		jargs.Opts.traceJsonSend(m)
	}

	_, err = j.stdin.Write(b)
	if err != nil {
		j.Close()
		return err
	}

	line := bytes.NewBuffer(nil)
	for true {
		l, p, err := j.stdoutReader.ReadLine()
		if err != nil {
			return err
		}
		line.Write(l)
		if !p {
			break
		}
	}

	if jargs.Opts.traceJsonReceive != nil {
		var m map[string]any
		_ = json.Unmarshal(line.Bytes(), &m)
		jargs.Opts.traceJsonReceive(m)
	}

	var result []jinja2Result
	err = json.Unmarshal(line.Bytes(), &result)
	if err != nil {
		return err
	}

	for i, item := range result {
		if item.Result != nil {
			processedJobs[i].Result = item.Result
		} else {
			if item.Error == nil {
				return fmt.Errorf("missing result and error from item at index %d", i)
			}
			processedJobs[i].Error = &Jinja2Error{*item.Error}
		}
	}

	return nil
}
