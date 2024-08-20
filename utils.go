package robin

import (
	"reflect"
	"strings"

	"github.com/go-viper/mapstructure/v2"
)

func guardedCast[Target any](value any, wantValue Target) (Target, error) {
	if value == nil {
		return wantValue, nil
	}

	params, ok := value.(Target)

	if !ok {
		switch reflect.TypeOf(wantValue).Kind() {
		// Numbers are a bit tricky, they are automatically converted to float64 when unmarshalled from JSON, so we need to check for that and convert to what we expect
		// If our expected param type is a number (int8, int16, int32, int64, int, uint8, uint16, uint32, uint64, uint, float32, float64), we can convert the raw param to that type and use it
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64:
			if reflect.TypeOf(value).Kind() == reflect.Float64 {
				params = reflect.ValueOf(value).Convert(reflect.TypeOf(wantValue)).Interface().(Target)
			} else {
				return wantValue, invalidTypesError(wantValue, value)
			}

			// Structs, arrays etc are decoded into map[key]|[] interface{} by the JSON decoder, so we can use mapstructure to decode them into the expected type
		case reflect.Struct, reflect.Slice, reflect.Array:
			mapstructure.Decode(value, &params)

		default:
			return wantValue, invalidTypesError(wantValue, value)
		}
	}

	return params, nil
}

func invalidTypesError(expected, got any) error {
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

	return RobinError{Reason: "Invalid types, expected " + expectedType + ", got " + gotType}
}
