package ugrpc

import (
	"fmt"
	"strings"
)

// SM2FM .
func SM2FM(service, method string) string {
	return fmt.Sprintf("/%s/%s", service, method)
}

// FM2SM .
func FM2SM(fm string) (s, m string) {
	vals := strings.Split(fm, "/")
	if len(vals) < 3 {
		return
	}
	s = vals[1]
	m = vals[2]
	return
}
