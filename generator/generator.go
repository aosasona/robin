package generator

import (
	_ "go.trulyao.dev/mirror/v2"
	"go.trulyao.dev/robin/types"
)

type Schema struct {
	Name           string `mirror:"name"`
	RequestSchema  any    `mirror:"name:request"`
	ResponseSchema any    `mirror:"name:response"`
}

type Exported struct {
	Queries   []Schema `mirror:"name:queries"`
	Mutations []Schema `mirror:"name:mutations"`
}

type generator struct{}

func New() generator {
	return generator{}
}

func Generate(procedures []types.Procedure) Exported {
	exported := Exported{
		Queries:   []Schema{},
		Mutations: []Schema{},
	}

	for _, procedure := range procedures {
		var schema Schema

		schema.Name = procedure.Name()
		schema.RequestSchema = procedure.PayloadInterface()
		schema.ResponseSchema = procedure.ReturnInterface()

		switch procedure.Type() {
		case types.ProcedureTypeQuery:
			exported.Queries = append(exported.Queries, schema)

		case types.ProcedureTypeMutation:
			exported.Mutations = append(exported.Mutations, schema)
		}
	}

	return exported
}
