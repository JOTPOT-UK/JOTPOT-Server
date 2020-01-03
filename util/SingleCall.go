package util

//SingleCallError returns a function, which when called for the first time, will call f and return nil, however subsequent calls will not call f and will return the given err.
func SingleCallError(f func(), err error) func() error {
	called := false
	return func() error {
		if called {
			return err
		}
		called = true
		f()
		return nil
	}
}

//SingleCallError returns a function, which when called for the first time, will call f and return nil, however subsequent calls will not call f and will return the given err.
func SingleCallErrorPassthrough(f func() error, err error) func() error {
	called := false
	return func() error {
		if called {
			return err
		}
		called = true
		return f()
	}
}

//SingleCallPanic returns a function, which when called for the first time, will call f, however subsequent calls will not call f, and will panic with the err give.
func SingleCallPanic(f func(), err interface{}) func() {
	called := false
	return func() {
		if called {
			panic(err)
		} else {
			called = true
			f()
		}
	}
}
