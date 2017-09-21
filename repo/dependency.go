package repo

// Dependency holds dependency information.
type Dependency struct {
	Import string
	Commit string `json:",omitempty" yaml:",omitempty"`
	Tag    string `json:",omitempty" yaml:",omitempty"`
	Branch string `json:",omitempty" yaml:",omitempty"`
}
