package repo

// State holds the state of a repo.
type State struct {
	Import   string
	Branches []string
	Tags     []string
	Commit   string
	Dirty    bool
	Exists   bool
}

// HasBranch returns true if the repo has the specified branch.
func (state *State) HasBranch(branch string) bool {
	for _, one := range state.Branches {
		if one == branch {
			return true
		}
	}
	return false
}

// HasTag returns true if the repo has the specified tag.
func (state *State) HasTag(tag string) bool {
	for _, one := range state.Tags {
		if one == tag {
			return true
		}
	}
	return false
}
