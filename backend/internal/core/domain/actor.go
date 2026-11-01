package domain

import "github.com/google/uuid"

// ErrForbidden is returned when the actor may not use a kind of resource at all, e.g. a client asking for
// personas. An actor who may not see one specific object gets that object's not-found error instead, so the
// answer does not reveal that it exists.
var ErrForbidden = forbidden("not allowed")

// UserID identifies a user: an employee of a tax practice or a client. A client logs in as itself, so the user id
// of a client is its client id.
type UserID uuid.UUID

// String implements fmt.Stringer.
func (id UserID) String() string { return uuid.UUID(id).String() }

// Role is what an actor is to its tax practice.
type Role int

const (
	// RoleStaff is an employee of the tax practice.
	RoleStaff Role = iota + 1
	// RoleClient is a client of the tax practice.
	RoleClient
)

// Actor is the caller of a use case: a user, its role and the tax practice (tenant) it belongs to. Use cases take
// it as an argument and decide what it may do; the tenant they act on is always Actor.TaxPracticeID.
// The fields are unexported, so an Actor is made with NewStaffActor or NewClientActor. The zero Actor has no role
// and no tax practice and is denied everything.
type Actor struct {
	role          Role
	userID        UserID
	taxPracticeID TaxPracticeID
}

// NewStaffActor returns an employee of the tax practice.
func NewStaffActor(taxPracticeID TaxPracticeID, userID UserID) Actor {
	return Actor{role: RoleStaff, userID: userID, taxPracticeID: taxPracticeID}
}

// NewClientActor returns a client of the tax practice.
func NewClientActor(taxPracticeID TaxPracticeID, userID UserID) Actor {
	return Actor{role: RoleClient, userID: userID, taxPracticeID: taxPracticeID}
}

// Role returns the role of the actor, 0 for the zero Actor.
func (a Actor) Role() Role { return a.role }

// UserID returns the user that acts.
func (a Actor) UserID() UserID { return a.userID }

// TaxPracticeID returns the tax practice the actor belongs to, the tenant of every use case it calls.
func (a Actor) TaxPracticeID() TaxPracticeID { return a.taxPracticeID }

// RequireStaff returns ErrForbidden unless a is an employee of its tax practice.
func (a Actor) RequireStaff() error {
	if a.role != RoleStaff {
		return ErrForbidden
	}

	return nil
}
