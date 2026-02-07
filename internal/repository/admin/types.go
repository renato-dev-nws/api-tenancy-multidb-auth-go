package admin

import "github.com/google/uuid"

// Tenant is a helper struct for middleware
type Tenant struct {
	ID     uuid.UUID
	DBCode uuid.UUID
	Status string
}
