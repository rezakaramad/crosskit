package pdns

import (
	"strings"
)

// Result represents the availability of a DNS name.
type Result struct {
	Available bool
	Reason    string
}

// EnsureTrailingDot appends a trailing dot to s if it does not already have one.
// PowerDNS and other DNS APIs require FQDNs to be dot-terminated.
func EnsureTrailingDot(s string) string {
	if strings.HasSuffix(s, ".") {
		return s
	}
	return s + "."
}
