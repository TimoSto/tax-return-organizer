package domain

import "github.com/google/uuid"

// ErrForbidden is returned when the actor may not use a kind of resource at all, e.g. a client asking for
// personas. An actor who may not see one specific object gets that object's not-found error instead, so the
// answer does not reveal that it exists.
var ErrForbidden = forbidden("not allowed")

// UserID identifies a user: an employee of a collector or a client. A client logs in as itself, so the user id
// of a client is its client id.
type UserID uuid.UUID

// String implements fmt.Stringer.
func (id UserID) String() string { return uuid.UUID(id).String() }

// Role is what an actor is to its collector.
type Role int

const (
	// RoleStaff is an employee of the collector.
	RoleStaff Role = iota + 1
	// RoleClient is a client of the collector.
	RoleClient
)

// Actor is the caller of a use case: a user, its role and the collector (tenant) it belongs to. Use cases take
// it as an argument and decide what it may do; the tenant they act on is always Actor.CollectorID.
// The fields are unexported, so an Actor is made with NewStaffActor or NewClientActor. The zero Actor has no role
// and no collector and is denied everything.
type Actor struct {
	role        Role
	userID      UserID
	collectorID CollectorID
}

// NewStaffActor returns an employee of the collector.
func NewStaffActor(collectorID CollectorID, userID UserID) Actor {
	return Actor{role: RoleStaff, userID: userID, collectorID: collectorID}
}

// NewClientActor returns a client of the collector.
func NewClientActor(collectorID CollectorID, userID UserID) Actor {
	return Actor{role: RoleClient, userID: userID, collectorID: collectorID}
}

// Role returns the role of the actor, 0 for the zero Actor.
func (a Actor) Role() Role { return a.role }

// UserID returns the user that acts.
func (a Actor) UserID() UserID { return a.userID }

// CollectorID returns the collector the actor belongs to, the tenant of every use case it calls.
func (a Actor) CollectorID() CollectorID { return a.collectorID }

// RequireStaff returns ErrForbidden unless a is an employee of its collector.
func (a Actor) RequireStaff() error {
	if a.role != RoleStaff {
		return ErrForbidden
	}

	return nil
}
