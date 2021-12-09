package unats

import "strings"

// IPSubject .
func IPSubject(ip string) string {
	return strings.ReplaceAll(ip, ".", "-")
}
