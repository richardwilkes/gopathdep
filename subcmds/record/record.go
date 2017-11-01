package record

import (
	"bytes"
	"errors"

	"github.com/richardwilkes/gokit/cmdline"
	"github.com/richardwilkes/gopathdep/imports"
	"github.com/richardwilkes/gopathdep/repo"
	"github.com/richardwilkes/gopathdep/util"
)

// Cmd holds the record command.
type Cmd struct {
}

// Name returns the name of the command as it needs to be entered on the command line.
func (cmd *Cmd) Name() string {
	return "record"
}

// Usage returns a description of what the command does.
func (cmd *Cmd) Usage() string {
	return "Record the current import state"
}

// Run the command.
func (cmd *Cmd) Run(cl *cmdline.CmdLine, args []string) error {
	var notags, useMasterWhenMissing, preserve bool
	cl.UsageSuffix = "[path to repo]"
	cl.NewBoolOption(&notags).SetSingle('n').SetName("notags").SetUsage("Disables recording of tags matching the current repo state")
	cl.NewBoolOption(&useMasterWhenMissing).SetSingle('m').SetName("master").SetUsage("Forces recording of missing repos as being tied to the master branch, rather than omitting them from the configuration")
	cl.NewBoolOption(&preserve).SetSingle('p').SetName("preserve").SetUsage("Preserve existing dependencies and only add new ones")
	remainingArgs := cl.Parse(args)
	if len(remainingArgs) == 0 {
		remainingArgs = []string{"."}
	}
	var missingCount int
	cfg := &repo.Config{Dir: util.MustGitRootOrDir(remainingArgs[0])}
	newMap := make(map[string]*repo.Dependency)
	states := imports.GetRepoStates(remainingArgs[0])
	for _, state := range states {
		var tag string
		var branch string
		var commit string
		if state.Exists {
			if !notags && len(state.Tags) > 0 {
				tag = state.Tags[0]
			}
			if tag == "" {
				commit = state.Commit
			}
		} else if useMasterWhenMissing {
			branch = "master"
		} else {
			missingCount++
		}
		if commit != "" || tag != "" || branch != "" {
			newMap[state.Import] = &repo.Dependency{
				Import: state.Import,
				Commit: commit,
				Tag:    tag,
				Branch: branch,
			}
		}
	}
	if preserve {
		if existingCfg, loadErr := repo.NewConfigFromDir(remainingArgs[0]); loadErr == nil {
			for _, dep := range existingCfg.Dependencies {
				newMap[dep.Import] = dep
			}
		}
	}
	cfg.Dependencies = make(repo.Dependencies, 0, len(newMap))
	for _, dep := range newMap {
		cfg.Dependencies = append(cfg.Dependencies, dep)
	}
	err := cfg.Save()
	if err == nil {
		if missingCount > 0 {
			buffer := bytes.Buffer{}
			buffer.WriteString("The following repos cannot be found and were not added:\n")
			for _, state := range states {
				if !state.Exists {
					buffer.WriteString("    ")
					buffer.WriteString(state.Import)
					buffer.WriteString("\n")
				}
			}
			err = errors.New(buffer.String())
		}
	}
	return err
}
