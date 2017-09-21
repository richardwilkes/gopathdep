package main

import (
	"fmt"
	"os"

	"github.com/richardwilkes/cmdline"
	"github.com/richardwilkes/gopathdep/subcmds/apply"
	"github.com/richardwilkes/gopathdep/subcmds/check"
	"github.com/richardwilkes/gopathdep/subcmds/record"
	"github.com/richardwilkes/gopathdep/subcmds/reset"
)

func main() {
	cmdline.AppName = "Go Path Dependencies"
	cmdline.AppVersion = "1.3"
	cmdline.CopyrightYears = "2017"
	cmdline.CopyrightHolder = "Richard A. Wilkes"
	cl := cmdline.New(true)
	cl.Description = "Manage $GOPATH dependencies."
	cl.UsageSuffix = "[path to repo]"
	cl.AddCommand(&apply.Cmd{})
	cl.AddCommand(&check.Cmd{})
	cl.AddCommand(&record.Cmd{})
	cl.AddCommand(&reset.Cmd{})
	if err := cl.RunCommand(cl.Parse(os.Args[1:])); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
