package jpserror

type UserSafeError interface {
	UserSafeStr() string
	Unwrap() error
}

func GetUserSafeError(err error, defaultSafeStr string) UserSafeError {
	if u, ok := err.(UserSafeError); ok {
		return u
	}
	return Error{
		SafeStr: defaultSafeStr,
		err:     err,
	}
}

/*
func PrintError(w io.Writer, err error) error {
	_, e := io.WriteString(w, err.Error())
	if e != nil {
		return e
	}
	err = errors.Unwrap(err)
	if err != nil {
		_, e = io.WriteString(w, "Unwrapped:")
		if e != nil {
			return e
		}
		return PrintError(w, err)
	}
	return nil
}*/

//Error represents an error, wrapped with a string which is safe to show to a user.
//This can be used to return an error with a useful message for the user but also with the origional error which could be logged further up but may contain information that the user should not see.
type Error struct {
	SafeStr string
	err     error
}

func (e Error) Error() string {
	return e.SafeStr + ": " + e.err.Error()
}

func (e Error) UserSafeStr() string {
	return e.SafeStr
}

func (e Error) Unwrap() error {
	return e.err
}
