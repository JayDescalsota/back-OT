package graph

import "github.com/ot/identity-svc/service"

type Resolver struct {
	AuthService *service.AuthService
	UserService *service.UserService
}
