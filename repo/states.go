package repo

// States provides a sortable slice of states.
type States []*State

func (s States) Len() int {
	return len(s)
}

func (s States) Less(i, j int) bool {
	return s[i].Import < s[j].Import
}

func (s States) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
