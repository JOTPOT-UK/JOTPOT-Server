package header

//Interface provides a standard interface for use with header implementations
type Interface interface {
	Get(key string) string
	Set(key, value string)
	Add(key, value string)
	Del(key string)
}

//Adder provides an interface for something which implements an add method to add a value to a key.
type Adder interface {
	Add(key, value string)
}
