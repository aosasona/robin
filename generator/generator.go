package generator

import (
	_ "go.trulyao.dev/mirror/v2"
	"go.trulyao.dev/robin/types"
)

type Schema struct{}

type Exported struct {
	Queries   Schema `mirror:"name:queries"`
	Mutations Schema `mirror:"name:mutations"`
}

type generator struct{}

func New() generator {
	return generator{}
}

func Generate(procedures []types.Procedure) Exported {
	exported := Exported{
		Queries:   Schema{},
		Mutations: Schema{},
	}

	// TODO: separate queries and mutations

	// TODO: dynamically create the schema via mirror hooks
	// Shape =>
	// type Robin = {
	//     queries: {
	//         queryName: {
	//             output: ...,  // type of the output/result here
	//             input: ...,  // type of the input here
	//         }
	//     }
	// }

	return exported
}
