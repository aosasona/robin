package types

type Middleware func(*Context) error

type ExclusionList []string

// Add adds a name to the exclusion list
func (e *ExclusionList) Add(name string) {
	*e = append(*e, name)
}

// AddMany adds multiple names to the exclusion list
func (e *ExclusionList) AddMany(names []string) {
	*e = append(*e, names...)
}

// Clear clears the exclusion list to free up memory
func (e *ExclusionList) Clear() {
	*e = []string{}
}

// Contains checks if a name is in the exclusion list
func (e *ExclusionList) Has(name string) bool {
	for _, n := range *e {
		if n == name {
			return true
		}
	}

	return false
}
