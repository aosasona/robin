package robin

import "go.trulyao.dev/robin/types"

type Procedures map[string]Procedure

// Get returns a procedure by name
func (p Procedures) Get(name string) Procedure {
	return p[name]
}

func (p Procedures) Exists(name string, procedureType types.ProcedureType) bool {
	proc, exists := p[name]
	return exists && proc.Type() == procedureType
}

// Add adds a procedure to the procedures map
func (p Procedures) Add(procedure Procedure) {
	p[procedure.Name()] = procedure
}

// Remove removes a procedure from the procedures map
func (p Procedures) Remove(name string) {
	delete(p, name)
}

// List returns the procedures as a slice
func (p Procedures) List() []Procedure {
	procedures := make([]Procedure, 0, len(p))
	for _, procedure := range p {
		procedures = append(procedures, procedure)
	}
	return procedures
}
