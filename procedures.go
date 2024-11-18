package robin

import "go.trulyao.dev/robin/types"

type (
	Procedures []Procedure
)

func (p *Procedures) Keys() []string {
	keys := make([]string, len(*p))
	for i, procedure := range *p {
		keys[i] = procedure.Name()
	}

	return keys
}

// Get returns a procedure by name
func (p *Procedures) Get(name string, procedureType ProcedureType) (Procedure, bool) {
	for _, procedure := range *p {
		if procedure.Name() == name && procedure.Type() == procedureType {
			return procedure, true
		}
	}

	return nil, false
}

func (p *Procedures) Exists(name string, procedureType types.ProcedureType) bool {
	for _, procedure := range *p {
		if procedure.Name() == name && procedure.Type() == procedureType {
			return true
		}
	}

	return false
}

// Add adds a procedure to the procedures map
func (p *Procedures) Add(procedure Procedure) {
	if p.Exists(procedure.Name(), procedure.Type()) {
		return
	}

	*p = append(*p, procedure)
}

// Remove removes a procedure from the procedures map
func (p *Procedures) Remove(name string, procedureType types.ProcedureType) {
	for i, procedure := range *p {
		if procedure.Name() == name && procedure.Type() == procedureType {
			*p = append((*p)[:i], (*p)[i+1:]...)
			break
		}
	}
}

// List returns the procedures as a slice
// NOTE: this is retained in case the underlying structure needs to ever change
func (p *Procedures) List() []Procedure {
	return *p
}

// Map returns the procedures as a map
func (p Procedures) Map() map[string]Procedure {
	procedures := make(map[string]Procedure)
	for _, procedure := range p {
		procedures[procedure.Name()] = procedure
	}

	return procedures
}
