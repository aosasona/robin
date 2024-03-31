package robin

type (
	ErrorHandler func(error) ([]byte, int)

	Error struct {
		Message       string
		Code          int
		OriginalError error
	}

	InternalError struct {
		Reason        string
		OriginalError error
	}
)

func (e Error) Error() string {
	return e.Message
}

func (ie InternalError) Error() string {
	return ie.Reason
}
