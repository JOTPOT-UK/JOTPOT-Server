package jps

import "time"

//ConditionType is used by a Condition to specify how it should behave (when it should fail).
type ConditionType uint8

const (
	//ConditionTypeModSince is a condition which fails iff the requested resource was not modified after the Time parameter.
	// PASS: The resource has a modification time that is after the Time parameter.
	// FAIL: The resource has not been modified since the time in the Time parameter.
	ConditionTypeModSince ConditionType = iota
	//ConditionTypeNotModSince is a condition which fails iff the requested resource was modified after the Time parameter.
	ConditionTypeNotModSince
	//ConditionTypeETag is a condition which fails iff the ETag of the requested resource is not equal to the String parameter.
	ConditionTypeETag
	//ConditionTypeETagNot is a condition which fails iff the ETag of the requested resource is equal to the String parameter.
	ConditionTypeETagNot
	ConditionTypeExists
	ConditionTypeNotExists
)

type TimeParam time.Time

func (tp TimeParam) Time() time.Time {
	return time.Time(tp)
}
func (tp TimeParam) Interface() interface{} {
	return tp.Time()
}
func (tp TimeParam) Str() string {
	return ""
}
func (tp TimeParam) Strs() []string {
	return nil
}

type StrParam string

func (sp StrParam) Str() string {
	return string(sp)
}
func (sp StrParam) Strs() []string {
	return []string{sp.Str()}
}
func (sp StrParam) Time() time.Time {
	return time.Time{}
}
func (sp StrParam) Interface() interface{} {
	return sp.Str()
}

type StrsParam []string

func (sp StrsParam) Str() string {
	if len(sp) > 0 {
		return sp[0]
	}
	return ""
}
func (sp StrsParam) Strs() []string {
	return sp
}
func (sp StrsParam) Time() time.Time {
	return time.Time{}
}
func (sp StrsParam) Interface() interface{} {
	return sp.Strs()
}

//ConditionParams is a struct of parameters that a condition can use.
//The Type of a Condition specifies which parameters should be used for what. Note that not all parameters may be used.
type ConditionParams interface {
	Interface() interface{}
	Str() string
	Strs() []string
	Time() time.Time
}

//Condition represents a condition that should be met for the requested resource to be sent.
//The Type field is used to specify when the Condition should fail - it may refer to parameters which are part of the ConditionParams struct.
//If a condition fails, then the response status of the response should be set to that Fail field, and no body should be sent.
type Condition struct {
	Type ConditionType
	Fail ResponseStatus
	ConditionParams
}
