package guarded

import (
	"reflect"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"go.trulyao.dev/robin/types"
)

// CastType attempts to cast the given value to the target type
// If the value is nil, it will return the target value
// If the value is already of the correct type, it will return it
// Numbers are a bit tricky, they are automatically converted to float64 when unmarshalled from JSON, so we need to check for that and convert to what we expect
//
// See implementation for more details
func CastType[Target any](from any, to Target) (Target, error) {
	if from == nil {
		return to, nil
	}

	// If the value is already of the correct type, we can just return it
	if target, ok := from.(Target); ok {
		return target, nil
	}

	// If our target is the void type (types.Void) and the value is nil, we can just return the target
	void := types.Void{}
	if reflect.TypeOf(to).Kind() == reflect.TypeOf(void).Kind() && from == nil {
		return to, nil
	}

	// Attempt to cast the value to the correct type

	var (
		targetType = reflect.TypeOf(to)
		params     Target
		ok         bool
	)

	switch targetType.Kind() {
	// Numbers are a bit tricky, they are automatically converted to float64 when unmarshalled from JSON, so we need to check for that and convert to what we expect
	// If our expected param type is a number (int8, int16, int32, int64, int, uint8, uint16, uint32, uint64, uint, float32, float64), we can convert the raw param to that type and use it
	case reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64,
		reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64,
		reflect.Float32,
		reflect.Float64:
		if reflect.TypeOf(from).Kind() == reflect.Float64 {
			// this is cursed
			params, ok = reflect.ValueOf(from).Convert(reflect.TypeOf(to)).Interface().(Target)
			if !ok {
				return to, MakeCastError(to, from)
			}
		} else {
			return to, MakeCastError(to, from)
		}

		// Structs, arrays etc are decoded into map[key]|[] interface{} by the JSON decoder, so we can use mapstructure to decode them into the expected type
	case reflect.Struct, reflect.Slice, reflect.Array:
		decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
			ErrorUnused: false,
			ErrorUnset:  false,
			Result:      &params,
			TagName:     "json",
		})
		if err != nil {
			return to, types.RobinError{Reason: "failed to create decoder", OriginalError: err}
		}

		err = decoder.Decode(from)
		if err != nil {
			return to, types.RobinError{Reason: "failed to decode value", OriginalError: err}
		}

	default:
		return to, MakeCastError(to, from)
	}

	return params, nil
}

// Attempts to construct a CastError from the expected and gotten types
func MakeCastError(expected, got any) error {
	expectedTypeOf := reflect.TypeOf(expected)
	gotTypeOf := reflect.TypeOf(got)

	if expectedTypeOf == nil || gotTypeOf == nil {
		return nil
	}

	expectedType := expectedTypeOf.Name()
	gotType := gotTypeOf.Name()

	if expectedType == "" {
		expectedType = expectedTypeOf.String()
	}

	if gotType == "" {
		gotType = gotTypeOf.String()
	}

	expectedType = strings.ReplaceAll(expectedType, "\"", "")
	gotType = strings.ReplaceAll(gotType, "\"", "")

	// If they are the same, we have overridden the gotten type in the guardedCast function, we will assume the gotten type is nil
	if expectedType == gotType {
		gotType = "nil"
	}

	return types.CastError{Expected: expectedType, Actual: gotType}
}
