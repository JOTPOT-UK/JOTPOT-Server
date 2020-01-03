package jpvariable

type Variables interface {
	Get(string) (interface{}, bool)
	Set(string, interface{})
}

type MapVariables map[string]interface{}

func (c MapVariables) Get(key string) (val interface{}, ok bool) {
	val, ok = c[key]
	return
}

func (c MapVariables) Set(key string, val interface{}) {
	c[key] = key
}

type VariablesScope struct {
	This, Parent Variables
}

func (s VariablesScope) Get(key string) (val interface{}, ok bool) {
	val, ok = s.This.Get(key)
	if !ok {
		val, ok = s.Parent.Get(key)
	}
	return
}

func (s VariablesScope) Set(key string, val interface{}) {
	s.This.Set(key, val)
}
