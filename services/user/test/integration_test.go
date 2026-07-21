package test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/99designs/gqlgen/graphql/handler"

	"github.com/clinicmanager/services/user/graph"
	"github.com/clinicmanager/services/user/graph/generated"
	"github.com/clinicmanager/services/user/repository"
	"github.com/clinicmanager/services/user/service"
	sharedCtx "github.com/clinicmanager/shared/context"
	sharedDB "github.com/clinicmanager/shared/db"
)

func getUserTestDBURL() string {
	url := os.Getenv("USERDB_URL")
	if url == "" {
		url = "postgres://postgres:1@localhost:5432/user?sslmode=disable"
	}
	return url
}

func setupUserTestServer(t *testing.T) *httptest.Server {
	t.Helper()

	dbURL := getUserTestDBURL()
	db, err := sharedDB.NewDB(dbURL)
	if err != nil {
		t.Skipf("Skipping integration test: cannot connect to real PostgreSQL database at %s: %v", dbURL, err)
	}
	t.Cleanup(func() { db.Close() })

	dbSet := sharedDB.NewDBSet(db)
	userRepo := repository.NewUserRepo(dbSet)
	userService := service.NewUserService(userRepo, sharedCtx.UserIDFromCtx, nil)

	resolver := &graph.Resolver{UserService: userService}
	executableSchema := generated.NewExecutableSchema(generated.Config{Resolvers: resolver})
	gqlServer := handler.NewDefaultServer(executableSchema)
	contextedHandler := sharedCtx.Tenant(gqlServer)

	return httptest.NewServer(contextedHandler)
}

type GraphQLResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

func executeGraphQL(t *testing.T, server *httptest.Server, query string, headers map[string]string) GraphQLResponse {
	t.Helper()

	body := map[string]string{"query": query}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal GraphQL request: %v", err)
	}

	req, err := http.NewRequestWithContext(context.Background(), "POST", server.URL, bytes.NewBuffer(bodyBytes))
	if err != nil {
		t.Fatalf("failed to create HTTP request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	client := server.Client()
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("HTTP request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var respBuf bytes.Buffer
		respBuf.ReadFrom(resp.Body)
		t.Fatalf("expected HTTP 200 OK, got %d: %s", resp.StatusCode, respBuf.String())
	}

	var gqlResp GraphQLResponse
	if err := json.NewDecoder(resp.Body).Decode(&gqlResp); err != nil {
		t.Fatalf("failed to decode response JSON: %v", err)
	}

	return gqlResp
}

func TestUserService_Integration_MeUser(t *testing.T) {
	server := setupUserTestServer(t)
	defer server.Close()

	query := `
		query {
			meUser {
				id
				email
				name
				isActive
				isValidated
				appRoles {
					id
					name
				}
			}
		}
	`

	gqlResp := executeGraphQL(t, server, query, map[string]string{
		"x-user-id": "00000000-0000-0000-0000-000000000001",
	})

	if len(gqlResp.Errors) > 0 {
		t.Fatalf("GraphQL query returned unexpected errors: %v", gqlResp.Errors)
	}

	var data map[string]struct {
		ID          string `json:"id"`
		Email       string `json:"email"`
		Name        string `json:"name"`
		IsActive    bool   `json:"isActive"`
		IsValidated bool   `json:"isValidated"`
		AppRoles    []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"appRoles"`
	}

	if err := json.Unmarshal(gqlResp.Data, &data); err != nil {
		t.Fatalf("failed to unmarshal GraphQL data: %v", err)
	}

	me, ok := data["meUser"]
	if !ok {
		t.Fatal("expected 'meUser' key in data payload")
	}

	t.Logf("Integration Test Success: got user %s (%s)", me.Name, me.Email)
	t.Logf("App Roles: %+v", me.AppRoles)

	if me.Email != "superadmin@clinic.com" {
		t.Errorf("expected email superadmin@clinic.com, got %s", me.Email)
	}
	if !me.IsActive {
		t.Error("expected user to be active")
	}
	if !me.IsValidated {
		t.Error("expected user to be validated")
	}
}

func TestUserService_Integration_MeUserUnauthenticated(t *testing.T) {
	server := setupUserTestServer(t)
	defer server.Close()

	query := `
		query {
			meUser {
				id
				email
			}
		}
	`

	gqlResp := executeGraphQL(t, server, query, map[string]string{})

	if len(gqlResp.Errors) == 0 {
		t.Fatal("expected error for unauthenticated meUser query, got none")
	}

	t.Logf("Integration Test Success: got expected error: %s", gqlResp.Errors[0].Message)
}

func TestUserService_Integration_UserByIDAsSuperAdmin(t *testing.T) {
	server := setupUserTestServer(t)
	defer server.Close()

	query := `
		query {
			user(id: "00000000-0000-0000-0000-000000000003") {
				id
				email
				name
				isActive
				isValidated
				appRoles {
					id
					name
				}
			}
		}
	`

	gqlResp := executeGraphQL(t, server, query, map[string]string{
		"x-user-id": "00000000-0000-0000-0000-000000000001",
	})

	if len(gqlResp.Errors) > 0 {
		t.Fatalf("GraphQL query returned unexpected errors: %v", gqlResp.Errors)
	}

	var data map[string]struct {
		ID          string `json:"id"`
		Email       string `json:"email"`
		Name        string `json:"name"`
		IsActive    bool   `json:"isActive"`
		IsValidated bool   `json:"isValidated"`
		AppRoles    []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"appRoles"`
	}

	if err := json.Unmarshal(gqlResp.Data, &data); err != nil {
		t.Fatalf("failed to unmarshal GraphQL data: %v", err)
	}

	user, ok := data["user"]
	if !ok {
		t.Fatal("expected 'user' key in data payload")
	}

	t.Logf("Integration Test Success: got user %s (%s)", user.Name, user.Email)

	if user.Email != "user01@clinic.com" {
		t.Errorf("expected email user01@clinic.com, got %s", user.Email)
	}
}

func TestUserService_Integration_UserByIDForbidden(t *testing.T) {
	server := setupUserTestServer(t)
	defer server.Close()

	query := `
		query {
			user(id: "00000000-0000-0000-0000-000000000004") {
				id
				email
			}
		}
	`

	gqlResp := executeGraphQL(t, server, query, map[string]string{
		"x-user-id": "00000000-0000-0000-0000-000000000003",
	})

	if len(gqlResp.Errors) == 0 {
		t.Fatal("expected forbidden error for non-super-admin user query, got none")
	}

	t.Logf("Integration Test Success: got expected error: %s", gqlResp.Errors[0].Message)
}

func TestUserService_Integration_FindUserByIDFederation(t *testing.T) {
	server := setupUserTestServer(t)
	defer server.Close()

	query := `
		query {
			_entities(representations: [
				{ __typename: "User", id: "00000000-0000-0000-0000-000000000002" }
			]) {
				... on User {
					id
					email
					name
					appRoles {
						id
						name
					}
				}
			}
		}
	`

	gqlResp := executeGraphQL(t, server, query, map[string]string{
		"x-user-id": "00000000-0000-0000-0000-000000000001",
	})

	if len(gqlResp.Errors) > 0 {
		t.Fatalf("GraphQL query returned unexpected errors: %v", gqlResp.Errors)
	}

	var data map[string][]struct {
		ID       string `json:"id"`
		Email    string `json:"email"`
		Name     string `json:"name"`
		AppRoles []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"appRoles"`
	}

	if err := json.Unmarshal(gqlResp.Data, &data); err != nil {
		t.Fatalf("failed to unmarshal GraphQL data: %v", err)
	}

	entities, ok := data["_entities"]
	if !ok {
		t.Fatal("expected '_entities' key in data payload")
	}

	if len(entities) != 1 {
		t.Fatalf("expected 1 entity, got %d", len(entities))
	}

	t.Logf("Integration Test Success: resolved User entity %s (%s)", entities[0].Name, entities[0].Email)
}

func TestUserService_Integration_FindAppRoleByIDFederation(t *testing.T) {
	server := setupUserTestServer(t)
	defer server.Close()

	query := `
		query {
			_entities(representations: [
				{ __typename: "AppRole", id: "1" }
			]) {
				... on AppRole {
					id
					name
					description
				}
			}
		}
	`

	gqlResp := executeGraphQL(t, server, query, map[string]string{
		"x-user-id": "00000000-0000-0000-0000-000000000001",
	})

	if len(gqlResp.Errors) > 0 {
		t.Fatalf("GraphQL query returned unexpected errors: %v", gqlResp.Errors)
	}

	var data map[string][]struct {
		ID          string  `json:"id"`
		Name        string  `json:"name"`
		Description *string `json:"description"`
	}

	if err := json.Unmarshal(gqlResp.Data, &data); err != nil {
		t.Fatalf("failed to unmarshal GraphQL data: %v", err)
	}

	entities, ok := data["_entities"]
	if !ok {
		t.Fatal("expected '_entities' key in data payload")
	}

	if len(entities) != 1 {
		t.Fatalf("expected 1 entity, got %d", len(entities))
	}

	t.Logf("Integration Test Success: resolved AppRole entity %s (%s)", entities[0].Name, *entities[0].Description)
}
