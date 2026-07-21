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

	"github.com/clinicmanager/services/tenant/graph"
	"github.com/clinicmanager/services/tenant/graph/generated"
	"github.com/clinicmanager/services/tenant/repository"
	"github.com/clinicmanager/services/tenant/service"
	sharedCtx "github.com/clinicmanager/shared/context"
	sharedDB "github.com/clinicmanager/shared/db"
)

func getTenantTestDBURL() string {
	url := os.Getenv("TENANTDB_URL")
	if url == "" {
		url = "postgres://postgres:1@localhost:5432/tenant?sslmode=disable"
	}
	return url
}

func setupTenantTestServer(t *testing.T) *httptest.Server {
	t.Helper()

	dbURL := getTenantTestDBURL()
	db, err := sharedDB.NewDB(dbURL)
	if err != nil {
		t.Skipf("Skipping integration test: cannot connect to real PostgreSQL database at %s: %v", dbURL, err)
	}
	t.Cleanup(func() { db.Close() })

	dbSet := sharedDB.NewDBSet(db)
	tenantRepo := repository.NewTenantRepo(dbSet)
	tenantService := service.NewTenantService(tenantRepo, nil)

	resolver := graph.NewResolver(tenantService)
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

func TestTenantService_Integration_TenantByID(t *testing.T) {
	server := setupTenantTestServer(t)
	defer server.Close()

	query := `
		query {
			tenant(id: "a1b2c3d4-0001-4000-8000-000000000001") {
				id
				name
				slug
				domain
				status
			}
		}
	`

	gqlResp := executeGraphQL(t, server, query, map[string]string{
		"x-tenant-id": "a1b2c3d4-0001-4000-8000-000000000001",
	})

	if len(gqlResp.Errors) > 0 {
		t.Fatalf("GraphQL query returned unexpected errors: %v", gqlResp.Errors)
	}

	var data map[string]struct {
		ID     string  `json:"id"`
		Name   string  `json:"name"`
		Slug   string  `json:"slug"`
		Domain *string `json:"domain"`
		Status string  `json:"status"`
	}

	if err := json.Unmarshal(gqlResp.Data, &data); err != nil {
		t.Fatalf("failed to unmarshal GraphQL data: %v", err)
	}

	tenant, ok := data["tenant"]
	if !ok {
		t.Fatal("expected 'tenant' key in data payload")
	}

	t.Logf("Integration Test Success: got tenant %s (%s)", tenant.Name, tenant.Slug)

	if tenant.Name != "Clinic A" {
		t.Errorf("expected Clinic A, got %s", tenant.Name)
	}
	if tenant.Slug != "clinic-a" {
		t.Errorf("expected clinic-a, got %s", tenant.Slug)
	}
}

func TestTenantService_Integration_BranchByID(t *testing.T) {
	server := setupTenantTestServer(t)
	defer server.Close()

	query := `
		query {
			branch(id: "b1b2c3d4-0001-4000-8000-000000000001") {
				id
				tenantId
				name
				timezone
				isActive
			}
		}
	`

	gqlResp := executeGraphQL(t, server, query, map[string]string{
		"x-tenant-id": "a1b2c3d4-0001-4000-8000-000000000001",
		"x-branch-id": "b1b2c3d4-0001-4000-8000-000000000001",
	})

	if len(gqlResp.Errors) > 0 {
		t.Fatalf("GraphQL query returned unexpected errors: %v", gqlResp.Errors)
	}

	var data map[string]struct {
		ID       string `json:"id"`
		TenantID string `json:"tenantId"`
		Name     string `json:"name"`
		Timezone string `json:"timezone"`
		IsActive bool   `json:"isActive"`
	}

	if err := json.Unmarshal(gqlResp.Data, &data); err != nil {
		t.Fatalf("failed to unmarshal GraphQL data: %v", err)
	}

	branch, ok := data["branch"]
	if !ok {
		t.Fatal("expected 'branch' key in data payload")
	}

	t.Logf("Integration Test Success: got branch %s (tenant: %s)", branch.Name, branch.TenantID)

	if branch.Name != "Branch A1" {
		t.Errorf("expected Branch A1, got %s", branch.Name)
	}
}

func TestTenantService_Integration_TenantRoleByID(t *testing.T) {
	server := setupTenantTestServer(t)
	defer server.Close()

	query := `
		query {
			tenantRole(id: "placeholder") {
				id
				name
				description
				isSystemRole
				tenantId
				permissions {
					id
					resource
					action
					scope
				}
			}
		}
	`

	gqlResp := executeGraphQL(t, server, query, map[string]string{
		"x-tenant-id": "a1b2c3d4-0001-4000-8000-000000000001",
		"x-branch-id": "b1b2c3d4-0001-4000-8000-000000000001",
	})

	if len(gqlResp.Errors) == 0 {
		t.Fatal("expected error for placeholder role ID, got success")
	}

	t.Logf("Integration Test Success: got expected error: %s", gqlResp.Errors[0].Message)
}

func TestTenantService_Integration_AssignmentsByUser(t *testing.T) {
	server := setupTenantTestServer(t)
	defer server.Close()

	query := `
		query {
			assignmentsByUser(userId: "00000000-0000-0000-0000-000000000002") {
				id
				userId
				role {
					id
					name
				}
				branch {
					id
					name
				}
				tenant {
					id
					name
				}
				isActive
			}
		}
	`

	gqlResp := executeGraphQL(t, server, query, map[string]string{
		"x-tenant-id": "a1b2c3d4-0001-4000-8000-000000000001",
		"x-branch-id": "b1b2c3d4-0001-4000-8000-000000000001",
	})

	if len(gqlResp.Errors) > 0 {
		t.Fatalf("GraphQL query returned unexpected errors: %v", gqlResp.Errors)
	}

	var data map[string][]struct {
		ID       string `json:"id"`
		UserID   string `json:"userId"`
		IsActive bool   `json:"isActive"`
		Role     struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"role"`
		Branch struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"branch"`
		Tenant struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"tenant"`
	}

	if err := json.Unmarshal(gqlResp.Data, &data); err != nil {
		t.Fatalf("failed to unmarshal GraphQL data: %v", err)
	}

	assignments, ok := data["assignmentsByUser"]
	if !ok {
		t.Fatal("expected 'assignmentsByUser' key in data payload")
	}

	t.Logf("Integration Test Success: got %d assignments for user", len(assignments))
	for _, a := range assignments {
		t.Logf("  Assignment: role=%s branch=%s tenant=%s", a.Role.Name, a.Branch.Name, a.Tenant.Name)
	}
}

func TestTenantService_Integration_MyAssignments(t *testing.T) {
	server := setupTenantTestServer(t)
	defer server.Close()

	query := `
		query {
			myAssignments {
				id
				userId
				role {
					id
					name
				}
				branch {
					id
					name
				}
				tenant {
					id
					name
				}
				isActive
			}
		}
	`

	gqlResp := executeGraphQL(t, server, query, map[string]string{
		"x-user-id":   "00000000-0000-0000-0000-000000000002",
		"x-tenant-id": "a1b2c3d4-0001-4000-8000-000000000001",
		"x-branch-id": "b1b2c3d4-0001-4000-8000-000000000001",
	})

	if len(gqlResp.Errors) > 0 {
		t.Fatalf("GraphQL query returned unexpected errors: %v", gqlResp.Errors)
	}

	var data map[string][]struct {
		ID       string `json:"id"`
		UserID   string `json:"userId"`
		IsActive bool   `json:"isActive"`
		Role     struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"role"`
		Branch struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"branch"`
		Tenant struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"tenant"`
	}

	if err := json.Unmarshal(gqlResp.Data, &data); err != nil {
		t.Fatalf("failed to unmarshal GraphQL data: %v", err)
	}

	assignments, ok := data["myAssignments"]
	if !ok {
		t.Fatal("expected 'myAssignments' key in data payload")
	}

	t.Logf("Integration Test Success: got %d myAssignments", len(assignments))
}

func TestTenantService_Integration_MeTenant(t *testing.T) {
	server := setupTenantTestServer(t)
	defer server.Close()

	query := `
		query {
			meTenant {
				id
				assignments {
					id
					userId
					role {
						id
						name
					}
					branch {
						id
						name
					}
					tenant {
						id
						name
					}
					isActive
				}
			}
		}
	`

	gqlResp := executeGraphQL(t, server, query, map[string]string{
		"x-user-id":   "00000000-0000-0000-0000-000000000003",
		"x-tenant-id": "a1b2c3d4-0001-4000-8000-000000000001",
		"x-branch-id": "b1b2c3d4-0001-4000-8000-000000000001",
	})

	if len(gqlResp.Errors) > 0 {
		t.Fatalf("GraphQL query returned unexpected errors: %v", gqlResp.Errors)
	}

	var data map[string]struct {
		ID          string `json:"id"`
		Assignments []struct {
			ID       string `json:"id"`
			UserID   string `json:"userId"`
			IsActive bool   `json:"isActive"`
			Role     struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"role"`
			Branch struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"branch"`
			Tenant struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"tenant"`
		} `json:"assignments"`
	}

	if err := json.Unmarshal(gqlResp.Data, &data); err != nil {
		t.Fatalf("failed to unmarshal GraphQL data: %v", err)
	}

	me, ok := data["meTenant"]
	if !ok {
		t.Fatal("expected 'meTenant' key in data payload")
	}

	t.Logf("Integration Test Success: got meTenant user %s with %d assignments", me.ID, len(me.Assignments))
}

func TestTenantService_Integration_MeTenantUnauthenticated(t *testing.T) {
	server := setupTenantTestServer(t)
	defer server.Close()

	query := `
		query {
			meTenant {
				id
			}
		}
	`

	gqlResp := executeGraphQL(t, server, query, map[string]string{})

	if len(gqlResp.Errors) == 0 {
		t.Fatal("expected error for unauthenticated meTenant query, got none")
	}

	t.Logf("Integration Test Success: got expected error: %s", gqlResp.Errors[0].Message)
}

func TestTenantService_Integration_FindTenantByIDFederation(t *testing.T) {
	server := setupTenantTestServer(t)
	defer server.Close()

	query := `
		query {
			_entities(representations: [
				{ __typename: "Tenant", id: "a1b2c3d4-0001-4000-8000-000000000001" }
			]) {
				... on Tenant {
					id
					name
					slug
				}
			}
		}
	`

	gqlResp := executeGraphQL(t, server, query, map[string]string{
		"x-tenant-id": "a1b2c3d4-0001-4000-8000-000000000001",
	})

	if len(gqlResp.Errors) > 0 {
		t.Fatalf("GraphQL query returned unexpected errors: %v", gqlResp.Errors)
	}

	var data map[string][]struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Slug string `json:"slug"`
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

	t.Logf("Integration Test Success: resolved Tenant entity %s (%s)", entities[0].Name, entities[0].Slug)
}

func TestTenantService_Integration_FindBranchByIDFederation(t *testing.T) {
	server := setupTenantTestServer(t)
	defer server.Close()

	query := `
		query {
			_entities(representations: [
				{ __typename: "Branch", id: "b1b2c3d4-0001-4000-8000-000000000001" }
			]) {
				... on Branch {
					id
					tenantId
					name
				}
			}
		}
	`

	gqlResp := executeGraphQL(t, server, query, map[string]string{
		"x-tenant-id": "a1b2c3d4-0001-4000-8000-000000000001",
		"x-branch-id": "b1b2c3d4-0001-4000-8000-000000000001",
	})

	if len(gqlResp.Errors) > 0 {
		t.Fatalf("GraphQL query returned unexpected errors: %v", gqlResp.Errors)
	}

	var data map[string][]struct {
		ID       string `json:"id"`
		TenantID string `json:"tenantId"`
		Name     string `json:"name"`
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

	t.Logf("Integration Test Success: resolved Branch entity %s", entities[0].Name)
}

func TestTenantService_Integration_FindUserAssignmentsFederation(t *testing.T) {
	server := setupTenantTestServer(t)
	defer server.Close()

	query := `
		query {
			_entities(representations: [
				{ __typename: "User", id: "00000000-0000-0000-0000-000000000003" }
			]) {
				... on User {
					id
					assignments {
						id
						role {
							name
						}
						branch {
							name
						}
						tenant {
							name
						}
					}
				}
			}
		}
	`

	gqlResp := executeGraphQL(t, server, query, map[string]string{
		"x-tenant-id": "a1b2c3d4-0001-4000-8000-000000000001",
		"x-branch-id": "b1b2c3d4-0001-4000-8000-000000000001",
	})

	if len(gqlResp.Errors) > 0 {
		t.Fatalf("GraphQL query returned unexpected errors: %v", gqlResp.Errors)
	}

	var data map[string][]struct {
		ID          string `json:"id"`
		Assignments []struct {
			ID     string                `json:"id"`
			Role   struct{ Name string } `json:"role"`
			Branch struct{ Name string } `json:"branch"`
			Tenant struct{ Name string } `json:"tenant"`
		} `json:"assignments"`
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

	t.Logf("Integration Test Success: resolved User entity %s with %d assignments", entities[0].ID, len(entities[0].Assignments))
	for _, a := range entities[0].Assignments {
		t.Logf("  Assignment: role=%s branch=%s tenant=%s", a.Role.Name, a.Branch.Name, a.Tenant.Name)
	}
}

func TestTenantService_Integration_FindTenantRoleByIDFederation(t *testing.T) {
	server := setupTenantTestServer(t)
	defer server.Close()

	query := `
		query {
			_entities(representations: [
				{ __typename: "TenantRole", id: "placeholder" }
			]) {
				... on TenantRole {
					id
					name
				}
			}
		}
	`

	gqlResp := executeGraphQL(t, server, query, map[string]string{
		"x-tenant-id": "a1b2c3d4-0001-4000-8000-000000000001",
		"x-branch-id": "b1b2c3d4-0001-4000-8000-000000000001",
	})

	if len(gqlResp.Errors) == 0 {
		t.Fatal("expected error for placeholder role ID, got success")
	}

	t.Logf("Integration Test Success: got expected error: %s", gqlResp.Errors[0].Message)
}
