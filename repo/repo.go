package repo

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/richardwilkes/gopathdep/util"
	"github.com/richardwilkes/toolbox/errs"
)

// Prefixes for the refs.
const (
	BranchPrefix                 = "refs/heads/"
	TagPrefix                    = "refs/tags/"
	maxSimultaneousShellRequests = 32
)

type response struct {
	req  *request
	data []byte
	err  error
}

type request struct {
	cmd      *exec.Cmd
	response chan *response
}

// Repo holds information for the git repo.
type Repo struct {
	ImportPath string
}

var (
	cmdQueue chan *request
)

func init() {
	cmdQueue = make(chan *request, 2*maxSimultaneousShellRequests)
	go processQueue()
}

// NewFromImportPath creates a new repo from the import path.
func NewFromImportPath(importPath string, force bool) (*Repo, error) {
	var err error
	gitRoot := fromGoPath(importPath)
	if !force {
		gitRoot, err = util.GitRoot(gitRoot)
	}
	if err == nil {
		stripped := util.StripPrefix(gitRoot, util.SrcPaths)
		if stripped != gitRoot {
			return &Repo{ImportPath: stripped}, nil
		}
		return nil, errs.New(fmt.Sprintf("%s is outside of $GOPATH %v", gitRoot, util.SrcPaths))
	}
	return nil, err
}

func fromGoPath(dir string) string {
	for _, one := range util.SrcPaths {
		one = filepath.ToSlash(filepath.Join(one, dir))
		if util.IsDir(one) {
			return one
		}
	}
	if len(util.SrcPaths) > 0 {
		return filepath.ToSlash(filepath.Join(util.SrcPaths[0], dir))
	}
	log.Fatalln("$GOPATH not set")
	return ""
}

// Root returns the root of the repo.
func (repo *Repo) Root() string {
	return fromGoPath(repo.ImportPath)
}

// Exec runs a git command against the repo.
func (repo *Repo) Exec(cmd string, arg ...string) (string, error) {
	command := exec.Command("git", append([]string{cmd}, arg...)...)
	command.Dir = repo.Root()
	return repo.runWithOutput(command)
}

func (repo *Repo) runWithOutput(cmd *exec.Cmd) (string, error) {
	rspCh := make(chan *response, 1)
	cmdQueue <- &request{
		cmd:      cmd,
		response: rspCh,
	}
	rsp := <-rspCh
	output := strings.TrimSpace(string(rsp.data))
	if rsp.err != nil {
		rsp.err = errs.NewWithCause(output, rsp.err)
	}
	return output, rsp.err
}

func processQueue() {
	var pending []*request
	var backlog []*request
	shellResponse := make(chan *response)
	for {
		select {
		case req := <-cmdQueue:
			if len(pending) < maxSimultaneousShellRequests {
				pending = append(pending, req)
				go runShell(req, shellResponse)
			} else {
				backlog = append(backlog, req)
			}
		case rsp := <-shellResponse:
			for i, one := range pending {
				if one == rsp.req {
					pending[i] = pending[len(pending)-1]
					pending[len(pending)-1] = nil
					pending = pending[:len(pending)-1]
					one.response <- rsp
					break
				}
			}
			if len(backlog) != 0 && len(pending) < maxSimultaneousShellRequests {
				req := backlog[0]
				pending = append(pending, req)
				copy(backlog, backlog[1:])
				backlog[len(backlog)-1] = nil
				backlog = backlog[:len(backlog)-1]
				go runShell(req, shellResponse)
			}
		}
	}
}

func runShell(req *request, rsp chan *response) {
	data, err := req.cmd.CombinedOutput()
	rsp <- &response{
		req:  req,
		data: data,
		err:  err,
	}
}

// State returns the state of the repo.
func (repo *Repo) State() *State {
	state := &State{Import: repo.ImportPath}
	if err := repo.Fetch(); err == nil {
		state.Exists = true
		if state.Commit, err = repo.Exec("rev-parse", "HEAD"); err == nil {
			var result string
			if result, err = repo.Exec("for-each-ref", "--points-at", state.Commit, `--format=%(refname)`); err == nil {
				for _, one := range strings.Split(result, "\n") {
					one = strings.TrimSpace(one)
					if one != "" {
						if strings.HasPrefix(one, BranchPrefix) {
							branch := strings.TrimPrefix(one, BranchPrefix)
							var remote string
							if remote, err = repo.Exec("rev-parse", "refs/remotes/origin/"+branch); err == nil && remote == state.Commit {
								state.Branches = append(state.Branches, branch)
							}
						} else if strings.HasPrefix(one, TagPrefix) {
							state.Tags = append(state.Tags, strings.TrimPrefix(one, TagPrefix))
						}
					}
				}
			}

			if result, err = repo.Exec("status", "--porcelain"); err == nil {
				state.Dirty = false
				for _, line := range strings.Split(result, "\n") {
					if line != "" && !strings.HasPrefix(line, "?? ") {
						state.Dirty = true
						break
					}
				}
			} else {
				state.Dirty = true
			}
		}
	}
	return state
}

// Remote returns the git remote URL for the repo.
func (repo *Repo) Remote() string {
	return GitRemote(repo.ImportPath)
}

// Clone runs the git clone command.
func (repo *Repo) Clone(branchOrTag string) error {
	root := repo.Root()
	dir := filepath.Dir(root)
	err := os.MkdirAll(dir, 0777)
	if err == nil {
		args := make([]string, 0, 5)
		args = append(args, "clone", "--quiet")
		if branchOrTag != "" {
			args = append(args, "--branch", branchOrTag)
		}
		args = append(args, repo.Remote(), filepath.Base(root))
		command := exec.Command("git", args...)
		command.Dir = dir
		_, err = repo.runWithOutput(command)
	}
	return err
}

// Fetch runs the git fetch command.
func (repo *Repo) Fetch() error {
	_, err := repo.Exec("fetch", "--quiet")
	return err
}

// Pull runs the git pull command.
func (repo *Repo) Pull() error {
	_, err := repo.Exec("pull", "--quiet")
	return err
}

// Checkout runs the git checkout command.
func (repo *Repo) Checkout(commit string) error {
	_, err := repo.Exec("checkout", "--quiet", commit)
	return err
}
