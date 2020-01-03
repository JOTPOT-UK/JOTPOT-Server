package jps

import "time"

type Condition struct {
	ModSince time.Time
	ETag string
	
	Fail ResponseStatus
}
