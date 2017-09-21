package repo

// Dependencies provides a sortable slice of dependencies.
type Dependencies []*Dependency

func (d Dependencies) Len() int {
	return len(d)
}

func (d Dependencies) Less(i, j int) bool {
	return d[i].Import < d[j].Import
}

func (d Dependencies) Swap(i, j int) {
	d[i], d[j] = d[j], d[i]
}
