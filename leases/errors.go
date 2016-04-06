package leases

import (
	"fmt"
)

// ErrMalformedUUID is an error implementation that's returned whenever a func failed to parse a
// UUID from plain text
type ErrMalformedUUID struct {
	uuidStr string
}

// Error is the error interface implementation
func (m ErrMalformedUUID) Error() string {
	return "malformed UUID: " + m.uuidStr
}

// ErrParseLease is an error implementation that's returned whenever a func failed to parse a
// lease from json
type ErrParseLease struct {
	leaseStr string
	parseErr error
}

// Error is the error interface implementation
func (e ErrParseLease) Error() string {
	return fmt.Sprintf("parsing lease %s (%s)", e.leaseStr, e.parseErr)
}
