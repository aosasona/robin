package generator

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"text/template"

	"go.trulyao.dev/mirror/v2"
	"go.trulyao.dev/mirror/v2/config"
	"go.trulyao.dev/mirror/v2/extractor/meta"
	"go.trulyao.dev/mirror/v2/generator/typescript"
	"go.trulyao.dev/mirror/v2/parser"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"go.trulyao.dev/robin/generator/templates"
	"go.trulyao.dev/robin/types"
)

var unexpectedErr = func(message string) error {
	return fmt.Errorf("error: %s. This should not happen, please file a bug report.", message)
}

// This is just a phanton type that is used to dynamically build the real schema
type (
	_RobinSchema struct{}

	_RobinExport struct {
		Queries   _RobinSchema `mirror:"name:queries"`
		Mutations _RobinSchema `mirror:"name:mutations"`
	}
)

type (
	generator struct {
		procedures     []types.Procedure
		mirrorInstance *mirror.Mirror
	}

	TemplateOpts struct {
		// Whether to include the schema in the generated bindings file or not
		IncludeSchema bool

		// The generated schema type
		Schema string

		// The generated query methods
		QueryMethods string

		// The generated mutation methods
		MutationMethods string

		// Whether to use the union result type or not - when enabled, the result type will be a uniion of the Ok and Error types which would disallow access to any of the fields without checking the `ok` field first
		UseUnionResult bool

		// Whether to throw a ProcedureCallError when a procedure call fails for any reason (e.g. invalid payload, user-defined error, etc.) instead of returning an error result
		ThrowOnError bool
	}

	MethodTemplateOpts struct {
		OriginalName string
		Name         string
		Type         string
		HasPayload   bool
		ThrowOnError bool
	}

	GenerateMethodsOpts struct {
		ThrowOnError bool
	}

	GenerateBindingsOpts struct {
		// Whether to include the schema in the generated bindings file or not
		IncludeSchema bool

		// The generated schema type
		Schema string

		// Whether to use the union result type or not - when enabled, the result type will be a uniion of the Ok and Error types which would disallow access to any of the fields without checking the `ok` field first
		UseUnionResult bool

		// Whether to throw a ProcedureCallError when a procedure call fails for any reason (e.g. invalid payload, user-defined error, etc.) instead of returning an error result
		ThrowOnError bool
	}

	GeneratedMethods struct {
		Mutations []string
		Queries   []string
	}
)

var invalidCharsRegex = regexp.MustCompile(`[^a-zA-Z0-9]`)

func New(procedures []types.Procedure) *generator {
	m := mirror.New(config.Config{
		Enabled:              true,
		EnableParserCache:    true,
		FlattenEmbeddedTypes: true,
	})

	return &generator{procedures: procedures, mirrorInstance: m}
}

func (g *generator) GenerateBindings(opts GenerateBindingsOpts) (string, error) {
	bindingsTemplate, err := template.ParseFS(templates.ClientTemplateFS, "client.template")
	if err != nil {
		return "", fmt.Errorf("failed to parse bindings template: %w", err)
	}

	methods, err := g.GenerateMethods(GenerateMethodsOpts{ThrowOnError: opts.ThrowOnError})
	if err != nil {
		return "", fmt.Errorf("failed to generate methods: %w", err)
	}

	var builder strings.Builder
	if err := bindingsTemplate.Execute(&builder, TemplateOpts{
		IncludeSchema:   opts.IncludeSchema,
		Schema:          strings.TrimSpace(opts.Schema),
		MutationMethods: strings.Join(methods.Mutations, "\n"),
		QueryMethods:    strings.Join(methods.Queries, "\n"),
		UseUnionResult:  opts.UseUnionResult,
		ThrowOnError:    opts.ThrowOnError,
	}); err != nil {
		return "", fmt.Errorf("failed to execute bindings template: %w", err)
	}

	return builder.String(), nil
}

func (g *generator) GenerateMethods(opts GenerateMethodsOpts) (*GeneratedMethods, error) {
	var mutations, queries []string

	for _, procedure := range g.procedures {
		methodTemplate := `
  /**
   * @procedure {{ .OriginalName }}
   *
   * @returns Promise<ProcedureResult<CSchema, "query", {{ printf "%q" .OriginalName }}>>
   {{if .ThrowOnError}}* @throws {ProcedureCallError} if the procedure call fails
   {{end}}**/
  async {{.Name}}({{ if .HasPayload }}payload: PayloadOf<CSchema, {{ printf "%q" .Type }}, {{ printf "%q" .OriginalName }}>, {{end}}opts?: CallOpts<CSchema, {{ printf "%q" .Type }}, {{ printf "%q" .OriginalName }}>): Promise<ProcedureResult<CSchema, {{ printf "%q" .Type }}, {{ printf "%q" .OriginalName }}>> {
    return await this.client.call({{ printf "%q" .Type }}, { ...opts, name: {{ printf "%q" .OriginalName }}, payload: {{ if .HasPayload }}payload{{else}}undefined{{end}} });
  }`

		var procedureType string
		switch procedure.Type() {
		case types.ProcedureTypeQuery:
			procedureType = "query"
		case types.ProcedureTypeMutation:
			procedureType = "mutation"
		default: // This should never happen
			return &GeneratedMethods{}, fmt.Errorf("unknown procedure type: %s", procedure.Type())
		}

		opts := MethodTemplateOpts{
			OriginalName: procedure.Name(),
			Name:         NormalizeProcedureName(procedure.Name()),
			Type:         procedureType,
			HasPayload:   reflect.TypeOf(procedure.PayloadInterface()).Name() != "_RobinVoid",
			ThrowOnError: opts.ThrowOnError,
		}

		method, err := template.New("method").Parse(methodTemplate)
		if err != nil {
			return &GeneratedMethods{}, fmt.Errorf("failed to parse method template: %w", err)
		}

		var methodBuilder strings.Builder
		if err := method.Execute(&methodBuilder, opts); err != nil {
			return &GeneratedMethods{}, fmt.Errorf("failed to execute method template: %w", err)
		}

		if procedure.Type() == types.ProcedureTypeQuery {
			queries = append(queries, methodBuilder.String())
		} else {
			mutations = append(mutations, methodBuilder.String())
		}
	}

	return &GeneratedMethods{Queries: queries, Mutations: mutations}, nil
}

// Generates the typescript schema for the given procedures
func (g *generator) GenerateSchema() (string, error) {
	g.mirrorInstance.Parser().OnParseItem(g.onParseItem)
	_ = g.mirrorInstance.Parser().
		AddCustomType("_RobinVoid", &parser.Scalar{ItemType: parser.TypeVoid})
	g.mirrorInstance.AddSources(_RobinExport{})

	target := typescript.DefaultConfig().
		// Inliing here is required for the actual schema bits to be in the final export rather than separately
		SetInlineObjects(true).
		SetIncludeSemiColon(true).
		SetPreferNullForNullable(true).
		SetPreferUnknown(true).
		SetIndentationType(config.IndentSpace).
		SetIndentationCount(2)

	target.Generator().SetHeaderText("")

	return g.mirrorInstance.GenerateforTarget(target)
}

func (g *generator) onParseItem(sourceName string, target parser.Item) error {
	switch item := target.(type) {
	case *parser.Struct:
		switch item.Name() {
		case "_RobinExport":
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
			Meta: meta.Meta{
				Name: fmt.Sprintf(`"%s"`, procedure.Name()),
			},
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

// NormalizeProcedureName normalizes the procedure name to a valid typescript function name
// For example, todo.create -> todoCreate, sign-in -> signIn etc
// This is done to ensure that the generated method names are valid
func NormalizeProcedureName(name string) string {
	// Split the name by spaces and capitalize each word
	words := strings.Split(invalidCharsRegex.ReplaceAllString(name, " "), " ")

	for i, word := range words {
		word = strings.ToLower(word)
		if i == 0 {
			continue
		}

		words[i] = cases.Title(language.English).String(word)
	}

	return strings.Join(words, "")
}
