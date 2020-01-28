package http

import (
	"fmt"
	"strings"
	"time"

	"github.com/JOTPOT-UK/JOTPOT-Server/http/header"
	"github.com/JOTPOT-UK/JOTPOT-Server/jps"
)

func ParseDate(date string) (time.Time, bool) {
	//Formats: RFC1123, RFC850, ANSIC
	t, err := time.Parse(time.RFC1123, strings.ToUpper(date))
	fmt.Println(err)
	if err == nil {
		return t, true
	}
	t, err = time.Parse(time.RFC850, date)
	if err == nil {
		return t, true
	}
	t, err = time.Parse(time.ANSIC, date)
	if err == nil {
		return t, true
	}
	return time.Time{}, false
}

func FormatDate(date time.Time) string {
	return date.Format(time.RFC1123)
}

func Conditions(h *header.Header, getOrHead bool) (conditions []jps.Condition) {
	if ifMatch := h.GetValues("If-Match"); len(ifMatch) > 0 {
		if len(ifMatch) == 1 && ifMatch[0] == "*" {
			conditions = append(conditions, jps.Condition{
				Type: jps.ConditionTypeExists,
				Fail: jps.ResponseStatusPreconditionFailed,
			})
		} else {
			conditions = append(conditions, jps.Condition{
				Type:            jps.ConditionTypeETag,
				Fail:            jps.ResponseStatusPreconditionFailed,
				ConditionParams: jps.StrsParam(ifMatch),
			})
		}
	} else {
		ifUnmodSince := h.GetValues("If-Unmodified-Since")
		//Since we MUST evaluate this, so be safe, we'll just check every date!
		//In principle, we could check for the most recent date here, but that is extra processing that we might not be required to do.
		for i := range ifUnmodSince {
			if date, ok := ParseDate(ifUnmodSince[i]); ok {
				conditions = append(conditions, jps.Condition{
					Type:            jps.ConditionTypeNotModSince,
					Fail:            jps.ResponseStatusPreconditionFailed,
					ConditionParams: jps.TimeParam(date),
				})
			}
		}
	}

	if ifNoneMatch := h.GetValues("If-None-Match"); len(ifNoneMatch) > 0 {
		if len(ifNoneMatch) == 1 && ifNoneMatch[0] == "*" {
			conditions = append(conditions, jps.Condition{
				Type: jps.ConditionTypeNotExists,
				Fail: jps.ResponseStatusPreconditionFailed,
			})
		} else {
			rs := jps.ResponseStatusPreconditionFailed
			if getOrHead {
				rs = jps.ResponseStatusNotModified
			}
			conditions = append(conditions, jps.Condition{
				Type:            jps.ConditionTypeETagNot,
				Fail:            rs,
				ConditionParams: jps.StrsParam(ifNoneMatch),
			})
		}
	} else if getOrHead {
		//TODO: More than 1 value?
		if ifModSince := h.GetValues("If-Modified-Since"); len(ifModSince) == 1 {
			fmt.Println("Hi!", ifModSince[0], ".")
			if date, ok := ParseDate(ifModSince[0]); ok {
				fmt.Println("See!")
				conditions = append(conditions, jps.Condition{
					Type:            jps.ConditionTypeModSince,
					Fail:            jps.ResponseStatusNotModified,
					ConditionParams: jps.TimeParam(date),
				})
			}
		}
	}
	return
}

//TODO: If-Range, 206 resp
