package robin

import (
	"fmt"
	"io"
	"reflect"

	"go.trulyao.dev/robin/types"
)

type ProcedureFn[Out any, In any] func(ctx *Context, body In) (Out, error)

type baseProcedure[Out any, In any] struct {
	// The name of the procedure
	name string

	// The function that will be called when the procedure is executed
	fn ProcedureFn[Out, In]

	// A placeholder for the type of the body that the procedure expects
	// WARNING: This never really has a value, it's just used for "type inference/reflection" during runtime
	in types.In[In]

	// A placeholder for the return type of the procedure
	// WARNING: This never really has a value, it's just used for "type inference/reflection" during runtime
	out Out

	// Middleware functions to be executed before the mutation is called
	middlewareFns []types.Middleware

	// Indicates whether the procedure expects a payload, and if so, what type of payload it expects
	expectedPayloadType types.ExpectedPayloadType

	// Excluded middleware functions
	excludedMiddleware *types.ExclusionList

	// The procedure alias
	alias string
}

func (b *baseProcedure[_, _]) Name() string {
	return b.name
}

func (b *baseProcedure[_, _]) Alias() string {
	return b.alias
}

// PayloadInterface returns a placeholder variable with the type of the payload that the procedure expects, this value is empty and only used for type inference/reflection during runtime
func (b *baseProcedure[_, _]) PayloadInterface() any {
	if b.expectedPayloadType == types.ExpectedPayloadRaw {
		return b.in.OverrideType()
	}

	return b.in.InferredType()
}

// ReturnInterface returns a placeholder variable with the type of the return value of the procedure, this value is empty and only used for type inference/reflection during runtime
func (b *baseProcedure[_, _]) ReturnInterface() any {
	return b.out
}

// ExpectsPayload returns whether the procedure expects a payload or not
func (b *baseProcedure[_, _]) ExpectedPayloadType() types.ExpectedPayloadType {
	return b.expectedPayloadType
}

// INFO: If you are wondering "why are you not just passing in in.InferredType()?",
// it is because interfaces are nil by default and lose all type information when passed around,
// so, the only way to keep the type information is to pass in the actual type from the function signature
func implementsReadCloser[Out, In any](fn ProcedureFn[Out, In]) bool {
	fnType := reflect.TypeOf(fn)
	if fnType.NumIn() != 2 {
		return false
	}

	secondArg := fnType.In(1)
	return secondArg == reflect.TypeOf((*io.ReadCloser)(nil)).Elem()
}

func mustImplementReadCloser[Out, In any](
	fn ProcedureFn[Out, In],
	procedureType types.ProcedureType,
) {
	if !implementsReadCloser(fn) {
		secondArg := reflect.TypeOf(fn).In(1)

		panic(
			fmt.Sprintf(
				"you called `WithRawPayload` on a %s that doesn't expect a raw payload, expected `io.ReadCloser` but got %s",
				procedureType,
				secondArg,
			),
		)
	}
}
