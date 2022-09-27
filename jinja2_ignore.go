package jinja2

import (
	"bufio"
	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
)

const (
	commentPrefix      = "#"
	templateIgnoreFile = ".templateignore"
)

// readIgnoreFile reads a specific git ignore file.
func (j *Jinja2) readIgnoreFile(fs fs.FS, p []string, ignoreFile string) (ps []gitignore.Pattern, err error) {
	f, err := fs.Open(filepath.Join(append(p, ignoreFile)...))
	if err == nil {
		defer f.Close()

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			s := scanner.Text()
			if !strings.HasPrefix(s, commentPrefix) && len(strings.TrimSpace(s)) > 0 {
				ps = append(ps, gitignore.ParsePattern(s, p))
			}
		}
	} else if !os.IsNotExist(err) {
		return nil, err
	}

	return
}

// readPatterns reads gitignore patterns recursively traversing through the directory
// structure. The result is in the ascending order of priority (last higher).
func (j *Jinja2) readPatterns(f fs.FS, p []string) (ps []gitignore.Pattern, err error) {
	ps, _ = j.readIgnoreFile(f, p, templateIgnoreFile)

	pd := p
	if len(p) == 0 {
		pd = []string{"."}
	}

	var fis []os.DirEntry
	fis, err = fs.ReadDir(f, path.Join(pd...))
	if err != nil {
		return
	}

	for _, fi := range fis {
		if fi.IsDir() {
			var subps []gitignore.Pattern
			subps, err = j.readPatterns(f, append(p, fi.Name()))
			if err != nil {
				return
			}

			if len(subps) > 0 {
				ps = append(ps, subps...)
			}
		}
	}

	return
}
