package types

// TODO: move other error types here
type (
	CastError struct {
		Expected string
		Actual   string
	}
)

func (ce CastError) Error() string {
	return "Failed to cast value, expected " + ce.Expected + ", got " + ce.Actual
}
