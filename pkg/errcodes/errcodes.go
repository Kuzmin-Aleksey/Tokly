package errcodes

type Code string

func (e Code) String() string {
	return string(e)
}

const (
	ErrUnknown        Code = "unknown error"
	ErrNotFound       Code = "not found"
	ErrInvalidRequest Code = "invalid request"
	ErrUnauthorized   Code = "unauthorized"
)
