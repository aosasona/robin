package generator

import (
	"errors"

	"go.trulyao.dev/mirror/v2"
	"go.trulyao.dev/mirror/v2/config"
	"go.trulyao.dev/mirror/v2/parser"

	"go.trulyao.dev/robin/types"
)

type Schema struct{}

type Export struct {
	Queries   Schema `mirror:"name:queries"`
	Mutations Schema `mirror:"name:mutations"`
}

type generator struct {
	procedures []types.Procedure
}

func New(procedures []types.Procedure) generator {
	return generator{
		procedures: procedures,
	}
}

func (g *generator) onParseItem(sourceName string, target parser.Item) error {
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

	switch schema := target.(type) {

	case *parser.Struct:
		if schema.Name() != "Export" {
			return nil
		}

		queries, exists := schema.GetField("queries")
		if !exists {
			return errors.New("missing queries field in Export struct. This should not happen, please file a bug report")
		}

		mutations, exists := schema.GetField("mutations")
		if !exists {
			return errors.New("missing mutations field in Export struct. This should not happen, please file a bug report")
		}

		// TODO: REMOVE THESE AFTER USING THEM
		var (
			_ = queries
			_ = mutations
		)

		for _, procedure := range g.procedures {
			switch procedure.Type() {
			case types.ProcedureTypeQuery:
			case types.ProcedureTypeMutation:
			}
		}

	default:
		// we only care about the struct
	}

	return nil
}

func (g *generator) Generate() Exported {
	exported := Exported{
		Queries:   Schema{},
		Mutations: Schema{},
	}

	m := mirror.New(config.Config{
		Enabled:              true,
		EnableParserCache:    true,
		FlattenEmbeddedTypes: true,
	})

	m.Parser().OnParseItem(g.onParseItem)

	return exported
}
