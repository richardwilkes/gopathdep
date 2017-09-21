package check

import (
	"fmt"
	"os"

	"github.com/richardwilkes/cmdline"
	"github.com/richardwilkes/gopathdep/imports"
	"github.com/richardwilkes/gopathdep/repo"
	"github.com/richardwilkes/term"
)

// Cmd holds the check command.
type Cmd struct {
}

// Name returns the name of the command as it needs to be entered on the command line.
func (cmd *Cmd) Name() string {
	return "check"
}

// Usage returns a description of what the command does.
func (cmd *Cmd) Usage() string {
	return "Check the current import state"
}

// Run the command.
func (cmd *Cmd) Run(cl *cmdline.CmdLine, args []string) error {
	var noColor, prune, errorsOnly bool
	cl.UsageSuffix = "[path to repo]"
	cl.NewBoolOption(&noColor).SetSingle('n').SetName("no-color").SetUsage("Use plain output that does not contain color and is suitable for parsing with scripts")
	cl.NewBoolOption(&prune).SetSingle('p').SetName("prune").SetUsage("Remove imports that are no longer needed from the configuration file")
	cl.NewBoolOption(&errorsOnly).SetSingle('e').SetName("errors-only").SetUsage("Suppress output for good imports")
	remainingArgs := cl.Parse(args)
	if len(remainingArgs) == 0 {
		remainingArgs = []string{"."}
	}
	cfg, err := repo.NewConfigFromDir(remainingArgs[0])
	if err == nil {
		deps := imports.GetDepInfo(cfg)
		if prune {
			cfg.Dependencies = make(repo.Dependencies, 0, len(deps))
			for _, dep := range deps {
				if dep.State == imports.MissingOnDisk || dep.State == imports.IncorrectVersion || dep.State == imports.Dirty || dep.State == imports.Good {
					cfg.Dependencies = append(cfg.Dependencies, dep.Dependency)
				}
			}
			err = cfg.Save()
		}
		if err == nil {
			out := term.NewANSI(os.Stdout)
			for _, dep := range deps {
				if !errorsOnly || dep.State != imports.Good {
					marker, description := dep.State.MarkerAndDescription()
					var rev string
					revColor := term.Blue
					if dep.Dependency != nil {
						if dep.Dependency.Commit != "" {
							rev = dep.Dependency.Commit
						} else if dep.Dependency.Tag != "" {
							rev = dep.Dependency.Tag
						} else if dep.Dependency.Branch != "" {
							rev = dep.Dependency.Branch
						}
					}
					if rev == "" {
						rev = "?"
						revColor = term.Red
					}
					if noColor {
						fmt.Fprintf(out, "%c %s [%s]\n", marker, dep.Import, rev)
					} else {
						var color term.Color
						if dep.State == imports.Good {
							color = term.Green
						} else {
							color = term.Red
						}
						out.Foreground(color, term.Bold)
						fmt.Fprintf(out, "%c", marker)
						out.Reset()
						fmt.Fprintf(out, " %s [", dep.Import)
						out.Foreground(revColor, term.Bold)
						fmt.Fprint(out, rev)
						out.Reset()
						fmt.Fprintf(out, "] %s", description)
						if prune && dep.State == imports.NotNeeded {
							fmt.Fprint(out, " (")
							out.Foreground(term.Green, term.Bold)
							fmt.Fprint(out, "removed")
							out.Reset()
							fmt.Fprint(out, ")")
						}
						fmt.Fprintln(out)
					}
				}
			}
		}
	}
	return err
}
