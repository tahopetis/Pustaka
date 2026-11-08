package ci

import (
	"context"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/tahopetis/porta/internal/testutils"
	"github.com/tahopetis/porta/pkg/logger"
)

// newDBService builds a Service backed by the real test database. Neo4j,
// Redis and audit are not exercised by the DB-only paths under test.
func newDBService(t *testing.T) (*Service, uuid.UUID) {
	t.Helper()
	pool, cleanup := testutils.SetupTestDB(t)
	t.Cleanup(cleanup)
	testutils.CleanupDB(t, pool)
	testutils.InsertTestData(t, pool)

	// Rows created by the tests need a real creator (FK to users).
	userID := uuid.New()
	if _, err := pool.Exec(context.Background(),
		`INSERT INTO users (id, username, email, password_hash) VALUES ($1, $2, $3, $4)`,
		userID, "ci-db-tests", "ci-db-tests@example.com", "not-a-real-hash",
	); err != nil {
		t.Fatalf("could not seed user: %v", err)
	}

	log := logger.New(logger.Config{Level: "disabled"})
	repo := NewRepository(pool, log)

	redisAddr := os.Getenv("TEST_REDIS_URL")
	if redisAddr == "" {
		redisAddr = "redis://localhost:6379"
	}
	redisOptions, err := redis.ParseURL(redisAddr)
	if err != nil {
		t.Fatalf("invalid TEST_REDIS_URL: %v", err)
	}
	client := redis.NewClient(redisOptions)
	t.Cleanup(func() { client.Close() })
	if err := client.FlushDB(context.Background()).Err(); err != nil {
		t.Fatalf("could not flush test redis: %v", err)
	}

	return NewService(repo, nil, client, nil, log), userID
}

func seedCIType(t *testing.T, svc *Service, userID uuid.UUID, name string) *CITypeDefinition {
	t.Helper()
	description := "integration test type"
	created, err := svc.repo.CreateCIType(context.Background(), &CITypeDefinition{
		Name:               name,
		Description:        &description,
		CreatedBy:          userID,
		RequiredAttributes: []AttributeDefinition{},
		OptionalAttributes: []AttributeDefinition{},
	})
	if err != nil {
		t.Fatalf("could not create CI type: %v", err)
	}
	return created
}

func TestServiceGetCITypeFromDatabase(t *testing.T) {
	svc, userID := newDBService(t)
	created := seedCIType(t, svc, userID, "Hardware")

	got, err := svc.GetCIType(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("GetCIType: %v", err)
	}
	if got.Name != "Hardware" || got.Description == nil || *got.Description != "integration test type" {
		t.Fatalf("unexpected CI type: %+v", got)
	}
}

func TestServiceGetCIFromDatabase(t *testing.T) {
	svc, userID := newDBService(t)
	seedCIType(t, svc, userID, "Hardware")

	created, err := svc.repo.CreateCI(context.Background(), &ConfigurationItem{
		Name:       "laptop-db-01",
		CIType:     "Hardware",
		Attributes: map[string]interface{}{"cpu": "M3"},
		Tags:       []string{"lab"},
		CreatedBy:  userID,
	})
	if err != nil {
		t.Fatalf("seed CI: %v", err)
	}

	got, err := svc.GetCI(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("GetCI: %v", err)
	}
	if got.Name != "laptop-db-01" || got.CIType != "Hardware" {
		t.Fatalf("unexpected CI: %+v", got)
	}

	if _, err := svc.GetCI(context.Background(), uuid.New()); err == nil {
		t.Fatal("expected an error for an unknown CI id")
	}
}

func TestServiceListCIsPaginationFromDatabase(t *testing.T) {
	svc, userID := newDBService(t)
	seedCIType(t, svc, userID, "Hardware")

	for _, name := range []string{"laptop-a", "laptop-b", "laptop-c"} {
		if _, err := svc.repo.CreateCI(context.Background(), &ConfigurationItem{
			Name:       name,
			CIType:     "Hardware",
			Attributes: map[string]interface{}{},
			Tags:       []string{},
			CreatedBy:  userID,
		}); err != nil {
			t.Fatalf("seed %s: %v", name, err)
		}
	}

	first, err := svc.ListCIs(context.Background(), ListCIFilters{}, 1, 2)
	if err != nil {
		t.Fatalf("ListCIs page 1: %v", err)
	}
	if len(first.CIs) != 2 || first.Total != 3 || first.Page != 1 || first.Limit != 2 || first.TotalPages != 2 {
		t.Fatalf("page 1 = %d items total=%d page=%d limit=%d pages=%d, want 2/3/1/2/2",
			len(first.CIs), first.Total, first.Page, first.Limit, first.TotalPages)
	}

	second, err := svc.ListCIs(context.Background(), ListCIFilters{}, 2, 2)
	if err != nil {
		t.Fatalf("ListCIs page 2: %v", err)
	}
	if len(second.CIs) != 1 {
		t.Fatalf("page 2 = %d items, want 1", len(second.CIs))
	}
}
