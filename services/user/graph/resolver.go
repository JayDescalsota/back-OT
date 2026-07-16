package graph

import `github.com/clinicmanager/services/user/service`

// Resolver is the root resolver. Kept in this file so gqlgen never overwrites it.
type Resolver struct {
	UserService *service.UserService
}
