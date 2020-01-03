package jps

type Method uint8

const (
	MethodUnknown = Method(iota)
	MethodOptions
	MethodStat
	MethodRead
	MethodReadDir
	MethodReadFile
	MethodReadLink
	MethodPost
	MethodMkdir
	MethodMkfile
	MethodWrite
	MethodMove
	MethodRemove
	MethodRemoveAll
	MethodLink
	MethodChown
	MethodChmod
	MethodChtimes
)
