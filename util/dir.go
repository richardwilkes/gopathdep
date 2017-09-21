package util

import (
	"fmt"
	"go/build"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/richardwilkes/errs"
)

// SrcPaths holds the $GOPATH source paths.
var SrcPaths []string

func init() {
	for _, path := range filepath.SplitList(build.Default.GOPATH) {
		SrcPaths = append(SrcPaths, filepath.ToSlash(fmt.Sprintf("%s%c", filepath.Join(path, "src"), filepath.Separator)))
	}
}

// IsRoot returns true if the path is a root directory.
func IsRoot(path string) bool {
	return filepath.IsAbs(path) && path == fmt.Sprintf("%s%c", filepath.VolumeName(path), filepath.Separator)
}

// IsDir returns true if the path is a directory.
func IsDir(path string) bool {
	fi, err := os.Stat(path)
	return err == nil && fi.IsDir()
}

// GitRoot searches for the git root directory for the path.
func GitRoot(path string) (string, error) {
	var err error
	original := path
	if path, err = filepath.Abs(path); err == nil {
		for {
			if IsDir(filepath.ToSlash(filepath.Join(path, ".git"))) {
				return filepath.ToSlash(path), nil
			}
			path = filepath.Dir(path)
			if IsRoot(path) {
				return "", errs.New(fmt.Sprintf("%s is not part of a git repo", original))
			}
		}
	} else {
		return "", errs.NewWithCause(fmt.Sprintf("%s is not a valid path", original), err)
	}
}

// MustGitRootOrDir returns the git root or the directory if it doesn't have one.
func MustGitRootOrDir(path string) string {
	gitRoot, err := GitRoot(path)
	if err != nil {
		if IsDir(path) {
			gitRoot = filepath.ToSlash(path)
		} else if gitRoot, err = os.Getwd(); err != nil {
			log.Fatalln(errs.NewWithCause("Unable to determine the current working directory", err))
		}
	}
	return gitRoot
}

// StripPrefix strips the first prefix that matches from the path.
func StripPrefix(path string, prefixes []string) string {
	for _, prefix := range prefixes {
		if strings.HasPrefix(path, prefix) {
			return strings.TrimPrefix(path, prefix)
		}
	}
	return path
}
