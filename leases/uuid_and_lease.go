package leases

import (
	"github.com/pborman/uuid"
)

type UUIDAndLease struct {
	UUID  uuid.UUID
	Lease *Lease
}

func NewUUIDAndLease(u uuid.UUID, l *Lease) *UUIDAndLease {
	return &UUIDAndLease{UUID: u, Lease: l}
}
