package generator

import (
	"fmt"
	"reflect"

	"go.trulyao.dev/mirror/v2"
	"go.trulyao.dev/mirror/v2/config"
	"go.trulyao.dev/mirror/v2/extractor/meta"
	"go.trulyao.dev/mirror/v2/generator/typescript"
	"go.trulyao.dev/mirror/v2/parser"

	"go.trulyao.dev/robin/types"
)

var unexpectedErr = func(message string) error {
	return fmt.Errorf("error: %s. This should not happen, please file a bug report.", message)
}

// This is just a phanton type that is used to dynamically build the real schema
type (
	__robin_schema struct{}

	__robin_export struct {
		Queries   __robin_schema `mirror:"name:queries"`
		Mutations __robin_schema `mirror:"name:mutations"`
	}
)

type generator struct {
	procedures     []types.Procedure
	mirrorInstance *mirror.Mirror
}

func New(procedures []types.Procedure) *generator {
	m := mirror.New(config.Config{
		Enabled:              true,
		EnableParserCache:    true,
		FlattenEmbeddedTypes: true,
	})

	return &generator{procedures: procedures, mirrorInstance: m}
}

// Generates the typescript schema for the given procedures
func (g *generator) GenerateSchema() (string, error) {
	g.mirrorInstance.Parser().OnParseItem(g.onParseItem)
	g.mirrorInstance.Parser().
		AddCustomType("__robin_void", &parser.Scalar{ItemType: parser.TypeVoid})
	g.mirrorInstance.AddSources(__robin_export{})

	target := typescript.DefaultConfig().
		// Inliing here is required for the actual schema bits to be in the final export rather than separately
		SetInlineObjects(true).
		SetIncludeSemiColon(true).
		SetPreferNullForNullable(true).
		SetPreferUnknown(true)

	target.Generator().SetHeaderText("")

	return g.mirrorInstance.GenerateforTarget(target)
}

func (g *generator) onParseItem(sourceName string, target parser.Item) error {
	switch item := target.(type) {
	case *parser.Struct:
		switch item.Name() {
		case "__robin_export":
			return g.handleClientType(item)

		default:
			return nil
		}
	}

	return nil
}

func (g *generator) handleClientType(item *parser.Struct) error {
	item.ItemName = "Schema"

	queries, mutations, err := g.getProcedureFields(item)
	if err != nil {
		return err
	}

	for _, procedure := range g.procedures {
		var returnItem, payloadItem parser.Item

		// Parse the output and input items
		if returnItem, err = g.mirrorInstance.Parser().Parse(reflect.TypeOf(procedure.ReturnInterface())); err != nil {
			return fmt.Errorf(
				"failed to parse output item for procedure %s: %w",
				procedure.Name(),
				err,
			)
		}

		if payloadItem, err = g.mirrorInstance.Parser().Parse(reflect.TypeOf(procedure.PayloadInterface())); err != nil {
			return fmt.Errorf(
				"failed to parse input item for procedure %s: %w",
				procedure.Name(),
				err,
			)
		}

		procedureField := parser.Field{
			ItemName: procedure.Name(),
			BaseItem: &parser.Struct{
				ItemName: procedure.Name(),
				Fields: []parser.Field{
					// Result
					{ItemName: "result", BaseItem: returnItem},

					// Payload
					{ItemName: "payload", BaseItem: payloadItem},
				},
			},
			Meta: meta.Meta{},
		}

		switch procedure.Type() {
		case types.ProcedureTypeQuery:
			queries.Fields = append(queries.Fields, procedureField)

		case types.ProcedureTypeMutation:
			mutations.Fields = append(mutations.Fields, procedureField)
		}
	}

	return nil
}

// WARNING: please do not refactor this to use implicit returns even though the function signature allows it (naked return), the signature here is used as some form of documentation since the first two	return values are the same type
func (g *generator) getProcedureFields(
	schema *parser.Struct,
) (queries *parser.Struct, mutations *parser.Struct, err error) {
	// Attempt to get the fields
	queriesField, exists := schema.GetField("queries")
	if !exists {
		return nil, nil, unexpectedErr("missing queries field in Export struct")
	}

	mutationsField, exists := schema.GetField("mutations")
	if !exists {
		return nil, nil, unexpectedErr("missing mutations field in Export struct")
	}

	var ok bool

	// Attempt to get a reference to base item's fields
	if queries, ok = queriesField.BaseItem.(*parser.Struct); !ok {
		return nil, nil, unexpectedErr("`queries` field in `Export` struct is not a struct")
	}

	if mutations, ok = mutationsField.BaseItem.(*parser.Struct); !ok {
		return nil, nil, unexpectedErr("`mutations` field in `Export` struct is not a struct")
	}

	return queries, mutations, nil
}
