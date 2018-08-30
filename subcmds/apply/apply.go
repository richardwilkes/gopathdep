package apply

import (
	"bytes"
	"errors"
	"fmt"
	"sync"

	"github.com/richardwilkes/gopathdep/imports"
	"github.com/richardwilkes/gopathdep/repo"
	"github.com/richardwilkes/toolbox/cmdline"
	"github.com/richardwilkes/toolbox/errs"
)

// Cmd holds the apply command.
type Cmd struct {
}

// Name returns the name of the command as it needs to be entered on the command line.
func (cmd *Cmd) Name() string {
	return "apply"
}

// Usage returns a description of what the command does.
func (cmd *Cmd) Usage() string {
	return "Apply the configured import state"
}

// Run the command.
func (cmd *Cmd) Run(cl *cmdline.CmdLine, args []string) error {
	cl.UsageSuffix = "[path to repo]"
	remainingArgs := cl.Parse(args)
	if len(remainingArgs) == 0 {
		remainingArgs = []string{"."}
	}
	cfg, err := repo.NewConfigFromDir(remainingArgs[0])
	if err == nil {
		deps := imports.GetDepInfo(cfg)
		buffer := &bytes.Buffer{}
		var lock sync.Mutex
		var wg sync.WaitGroup
		for _, dep := range deps {
			wg.Add(1)
			go process(dep.Dependency, dep.State, buffer, &lock, &wg)
		}
		wg.Wait()
		if buffer.Len() > 0 {
			err = errors.New(buffer.String())
		}
	}
	return err
}

func process(dep *repo.Dependency, depState imports.DepState, buffer *bytes.Buffer, lock sync.Locker, wg *sync.WaitGroup) {
	defer wg.Done()
	var r *repo.Repo
	var err error
	switch depState {
	case imports.MissingOnDisk:
		if r, err = repo.NewFromImportPath(dep.Import, true); err == nil {
			var branchOrTag string
			if dep.Commit == "" {
				branchOrTag = dep.Tag
				if branchOrTag == "" {
					branchOrTag = dep.Branch
				}
			}
			if err = r.Clone(branchOrTag); err == nil {
				if branchOrTag == "" && dep.Commit != "" {
					var existing string
					if state := r.State(); state != nil {
						existing = state.Commit
					}
					if existing != dep.Commit {
						err = r.Checkout(dep.Commit)
					}
				}
				if err == nil {
					lock.Lock()
					fmt.Printf("Cloned %s and checked out ", dep.Import)
					describe(dep)
					lock.Unlock()
				}
			}
		}
		if err != nil {
			lock.Lock()
			fmt.Fprintln(buffer, errs.NewfWithCause(err, "Error: Unable to checkout %s", dep.Import))
			lock.Unlock()
		}
	case imports.NotNeeded:
		if r, err = repo.NewFromImportPath(dep.Import, false); err == nil {
			if rs := r.State(); rs != nil {
				wg.Add(1)
				process(dep, imports.GetDepState(dep, rs), buffer, lock, wg)
			}
		} else {
			wg.Add(1)
			process(dep, imports.MissingOnDisk, buffer, lock, wg)
		}
	case imports.IncorrectVersion:
		if r, err = repo.NewFromImportPath(dep.Import, false); err == nil {
			target := dep.Commit
			if target == "" {
				target = dep.Tag
				if target == "" {
					target = dep.Branch
					if target == "" {
						target = "master"
					}
				}
			}
			if err = r.Fetch(); err == nil {
				if err = r.Checkout(target); err == nil {
					if target == "master" || target == dep.Branch {
						err = r.Pull()
					}
					if err == nil {
						lock.Lock()
						fmt.Printf("Updated %s to ", dep.Import)
						describe(dep)
						lock.Unlock()
					}
				}
			}
		}
		if err != nil {
			lock.Lock()
			fmt.Fprintln(buffer, errs.NewfWithCause(err, "Error: Unable to update %s", dep.Import))
			lock.Unlock()
		}
	case imports.Dirty:
		_, description := depState.MarkerAndDescription()
		lock.Lock()
		fmt.Fprintf(buffer, "Error: %s %s\n", dep.Import, description)
		lock.Unlock()
	}
}

func describe(dep *repo.Dependency) {
	if dep.Commit != "" {
		fmt.Printf("commit %s\n", dep.Commit)
	} else if dep.Tag != "" {
		fmt.Printf("tag %s\n", dep.Tag)
	} else if dep.Branch != "" {
		fmt.Printf("branch %s\n", dep.Branch)
	} else {
		fmt.Println("branch master")
	}
}
