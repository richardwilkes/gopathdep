package imports

import (
	"go/build"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/richardwilkes/gopathdep/repo"
	"github.com/richardwilkes/gopathdep/util"
)

// GetRepoStates returns the repo states for each import found in the directory.
func GetRepoStates(dir string) []*repo.State {
	rootPkgs := CollectRootPackageNames(dir)
	states := make([]*repo.State, len(rootPkgs))
	wg := sync.WaitGroup{}
	for i, pkg := range rootPkgs {
		wg.Add(1)
		go func(idx int, pkgName string) {
			defer wg.Done()
			if r, err := repo.NewFromImportPath(pkgName, false); err == nil {
				states[idx] = r.State()
			} else {
				states[idx] = &repo.State{Import: pkgName}
			}
		}(i, pkg)
	}
	wg.Wait()
	return states
}

// CollectRootPackageNames collects the root package names of imports.
func CollectRootPackageNames(dir string) []string {
	// Collect the primary package directories
	var err error
	set := make(map[string]string)
	if dir, err = filepath.Abs(dir); err == nil {
		if filepath.Walk(dir, func(path string, info os.FileInfo, walkerErr error) error {
			if walkerErr != nil {
				return walkerErr
			}
			if info.IsDir() {
				name := info.Name()
				if strings.HasPrefix(name, ".") || strings.HasPrefix(name, "_") || name == "vendor" {
					return filepath.SkipDir
				}
				set[path] = path
			}
			return nil
		}) != nil {
			util.Ignore()
		}
	}
	dirs := make([]string, 0, len(set))
	for k := range set {
		dirs = append(dirs, filepath.ToSlash(k))
	}

	// Collect the package names
	set = make(map[string]string)
	for _, one := range dirs {
		if pkg, err := build.Default.Import(util.StripPrefix(one, util.SrcPaths), dir, 0); err == nil {
			if !pkg.Goroot {
				collectPackageNamesFromSlice(pkg.Imports, dir, set)
				// For the original source dirs only, collect imports from tests
				collectPackageNamesFromSlice(pkg.TestImports, dir, set)
				collectPackageNamesFromSlice(pkg.XTestImports, dir, set)
			}
		}
	}

	// Remove any from the primary package directories
	for _, one := range dirs {
		delete(set, util.StripPrefix(one, util.SrcPaths))
	}
	delete(set, "C")

	// Transform them into root pkg names
	pkgs := make(map[string]bool)
	for orig, revised := range set {
		if orig != "" && !strings.HasSuffix(orig, "_test") {
			pkgs[revised] = true
		}
	}
	names := make([]string, 0, len(pkgs))
	for one := range pkgs {
		names = append(names, one)
	}
	sort.Strings(names)

	return names
}

func collectPackageNames(pkgName string, srcDir string, set map[string]string) {
	if pkg, err := build.Default.Import(pkgName, srcDir, 0); err == nil {
		if !pkg.Goroot {
			set[pkgName] = findRepoRoot(pkg.SrcRoot, pkg.ImportPath)
			collectPackageNamesFromSlice(pkg.Imports, srcDir, set)
		}
	} else {
		// Package can't be found, so no way to check its dependencies, but we can add it to our set.
		set[pkgName] = pkgName
	}
}

func collectPackageNamesFromSlice(pkgNames []string, srcDir string, set map[string]string) {
	for _, pkgName := range pkgNames {
		if _, exists := set[pkgName]; !exists {
			collectPackageNames(pkgName, srcDir, set)
		}
	}
}

func findRepoRoot(root, path string) string {
	dir := path
	for {
		if fi, err := os.Stat(filepath.Join(root, dir, ".git")); err == nil && fi.IsDir() {
			return dir
		}
		dir = filepath.Dir(dir)
		if dir == "." || dir == "/" {
			return path
		}
	}
}
