package imports

import (
	"log"
	"sort"

	"github.com/richardwilkes/gopathdep/repo"
)

// DepState holds the state of the dependency
type DepState int

// The possible DepStates
const (
	MissingConfig DepState = iota
	MissingOnDisk
	MissingOnDiskAndConfig
	NotNeeded
	IncorrectVersion
	Dirty
	Good
)

// DepInfo holds the dependency info for a one import.
type DepInfo struct {
	Import     string
	Dependency *repo.Dependency
	State      DepState
}

// DepInfos holds multiple DepInfo records and provides convenient sorting.
type DepInfos []*DepInfo

// MarkerAndDescription returns the single-character marker and description for a DepState.
func (ds DepState) MarkerAndDescription() (rune, string) {
	switch ds {
	case MissingConfig:
		return 'C', "missing from configuration"
	case MissingOnDisk:
		return 'D', "missing from $GOPATH"
	case MissingOnDiskAndConfig:
		return 'B', "missing from configuration and $GOPATH"
	case NotNeeded:
		return 'X', "can be removed from the configuration"
	case IncorrectVersion:
		return 'S', "needs to be synced with this configuration"
	case Dirty:
		return 'M', "is modified"
	case Good:
		return '✓', ""
	default:
		log.Fatalf("Unknown dependency state: %v\n", ds)
		return 0, ""
	}
}

func (di DepInfos) Len() int {
	return len(di)
}

func (di DepInfos) Less(i, j int) bool {
	return di[i].Import < di[j].Import
}

func (di DepInfos) Swap(i, j int) {
	di[i], di[j] = di[j], di[i]
}

// GetDepInfo returns the dependency information.
func GetDepInfo(cfg *repo.Config) DepInfos {
	states := GetRepoStates(cfg.Dir)
	pkgToStateMap := make(map[string]*repo.State)
	for _, state := range states {
		pkgToStateMap[state.Import] = state
	}
	cfgCnt := len(cfg.Dependencies)
	pkgCnt := len(states)
	if pkgCnt < cfgCnt {
		pkgCnt = cfgCnt
	}
	deps := make(DepInfos, 0, pkgCnt)
	for _, dep := range cfg.Dependencies {
		di := &DepInfo{Import: dep.Import, Dependency: dep}
		if state, exists := pkgToStateMap[dep.Import]; exists {
			delete(pkgToStateMap, dep.Import)
			if state.Exists {
				di.State = GetDepState(dep, state)
			} else {
				di.State = MissingOnDisk
			}
		} else {
			di.State = NotNeeded
		}
		deps = append(deps, di)
	}
	for pkgName, state := range pkgToStateMap {
		if state.Exists {
			deps = append(deps, &DepInfo{Import: pkgName, State: MissingConfig})
		} else {
			deps = append(deps, &DepInfo{Import: pkgName, State: MissingOnDiskAndConfig})
		}
	}
	sort.Sort(deps)
	return deps
}

// GetDepState returns the dependency status for a dependency.
func GetDepState(dep *repo.Dependency, state *repo.State) DepState {
	if (dep.Commit != "" && dep.Commit != state.Commit) || (dep.Tag != "" && !state.HasTag(dep.Tag) || (dep.Branch != "" && !state.HasBranch(dep.Branch))) {
		return IncorrectVersion
	} else if state.Dirty {
		return Dirty
	} else {
		return Good
	}
}
