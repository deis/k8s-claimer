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

// ErrEncodeLease is an error implementation that's returned whenever a func failed to encode a
// lease into json
type ErrEncodeLease struct {
	l         *Lease
	encodeErr error
}

// Error is the error interface implementation
func (e ErrEncodeLease) Error() string {
	return fmt.Sprintf("error encoding lease for cluster %s (%s)", e.l.ClusterName, e.encodeErr)
}
