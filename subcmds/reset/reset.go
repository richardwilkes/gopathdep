package reset

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/richardwilkes/gopathdep/repo"
	"github.com/richardwilkes/gopathdep/util"
	"github.com/richardwilkes/toolbox/cmdline"
	"github.com/richardwilkes/toolbox/xio/term"
)

// Cmd holds the reset command.
type Cmd struct {
}

// Name returns the name of the command as it needs to be entered on the command line.
func (cmd *Cmd) Name() string {
	return "reset"
}

// Usage returns a description of what the command does.
func (cmd *Cmd) Usage() string {
	return "Reset all repos on $GOPATH back to the master branch"
}

// Run the command.
func (cmd *Cmd) Run(cl *cmdline.CmdLine, args []string) error {
	cl.Parse(args)
	roots := make(map[string]bool)
	for _, srcRoot := range util.SrcPaths {
		if filepath.Walk(srcRoot, func(path string, info os.FileInfo, walkerErr error) error {
			if walkerErr != nil {
				return walkerErr
			}
			if info.IsDir() {
				name := info.Name()
				if name == ".git" {
					roots[filepath.Dir(path)] = true
					return filepath.SkipDir
				}
				if strings.HasPrefix(name, ".") || strings.HasPrefix(name, "_") || name == "vendor" {
					return filepath.SkipDir
				}
				if _, exists := roots[filepath.Dir(path)]; exists {
					return filepath.SkipDir
				}
			}
			return nil
		}) != nil {
			util.Ignore()
		}
	}
	out := term.NewANSI(os.Stdout)
	var wg sync.WaitGroup
	var lock sync.Mutex
	for root := range roots {
		importPath := util.StripPrefix(root, util.SrcPaths)
		if one, err := repo.NewFromImportPath(importPath, false); err == nil {
			wg.Add(1)
			go func(r *repo.Repo) {
				defer wg.Done()
				if state := r.State(); state.Exists {
					var markerColor term.Color
					var marker rune
					var description string
					var revision string
					var err error
					if state.Dirty {
						markerColor = term.Red
						marker = 'M'
						description = "is modified and will not be updated"
					} else if !state.HasBranch("master") {
						err = r.Checkout("master")
						if err == nil {
							if err = r.Pull(); err == nil {
								markerColor = term.Green
								marker = '✓'
								description = "has been updated to"
								revision = "master"
							}
						}
					}
					if err == nil {
						if marker != 0 {
							lock.Lock()
							out.Foreground(markerColor, term.Bold)
							fmt.Fprintf(out, "%c", marker)
							out.Reset()
							fmt.Fprintf(out, " %s %s", r.ImportPath, description)
							if revision != "" {
								fmt.Fprint(out, " [")
								out.Foreground(term.Blue, term.Bold)
								fmt.Fprint(out, revision)
								out.Reset()
								fmt.Fprint(out, "]")
							}
							fmt.Fprint(out, "\n")
							lock.Unlock()
						}
					} else {
						lock.Lock()
						out.Foreground(term.Red, term.Bold)
						fmt.Println(err)
						out.Reset()
						lock.Unlock()
					}
				}
			}(one)
		}
	}
	wg.Wait()
	return nil
}
