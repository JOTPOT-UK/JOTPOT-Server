package jpsheader

import "strings"

func Prefix(s string) string {
	if strings.HasPrefix(s, "Jps-") {
		return "Jps-Had-" + s[4:]
	}
	return s
}
