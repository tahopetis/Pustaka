package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/tahopetis/porta/internal/auth"
	"github.com/tahopetis/porta/internal/testutils"
	"github.com/tahopetis/porta/pkg/logger"
)

const (
	dbTestPassword = "S3cure-Passw0rd-xyz!"
	// Seeded by testutils.InsertTestData.
	adminRoleID = "550e8400-e29b-41d4-a716-446655440001"
)

// TestAuthHandlerLoginAgainstDatabase exercises the real login handler against
// the migrated PostgreSQL schema: credential lookup, password verification,
// role loading and token generation.
func TestAuthHandlerLoginAgainstDatabase(t *testing.T) {
	pool, cleanup := testutils.SetupTestDB(t)
	t.Cleanup(cleanup)
	testutils.CleanupDB(t, pool)
	testutils.InsertTestData(t, pool)

	ctx := context.Background()

	passwordService := auth.NewPasswordService()
	hash, err := passwordService.HashPassword(dbTestPassword)
	if err != nil {
		t.Fatalf("could not hash password: %v", err)
	}

	userID := uuid.New()
	if _, err := pool.Exec(ctx,
		`INSERT INTO users (id, username, email, password_hash, is_active) VALUES ($1, $2, $3, $4, true)`,
		userID, "dbuser", "dbuser@example.com", hash,
	); err != nil {
		t.Fatalf("could not insert user: %v", err)
	}
	if _, err := pool.Exec(ctx,
		`INSERT INTO user_roles (user_id, role_id, assigned_by) VALUES ($1, $2, $1)`,
		userID, adminRoleID,
	); err != nil {
		t.Fatalf("could not assign role: %v", err)
	}

	rbac := auth.NewRBACService(pool)
	jwtService := auth.NewJWTService("test-secret-with-at-least-32-characters", time.Hour, 24*time.Hour, "porta-test")
	handler := NewAuthHandler(jwtService, passwordService, rbac, logger.New(logger.Config{Level: "disabled"}))

	router := chi.NewRouter()
	router.Post("/login", handler.Login)

	login := func(t *testing.T, body map[string]string) *httptest.ResponseRecorder {
		t.Helper()
		payload, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("could not marshal request: %v", err)
		}
		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(payload))
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		return rec
	}

	t.Run("valid credentials return tokens", func(t *testing.T) {
		rec := login(t, map[string]string{"username": "dbuser", "password": dbTestPassword})
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200 (body %q)", rec.Code, rec.Body.String())
		}
		var resp LoginResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("response is not JSON: %v", err)
		}
		if resp.AccessToken == "" || resp.RefreshToken == "" {
			t.Fatal("expected both tokens in the response")
		}
		if resp.TokenType != "Bearer" || resp.ExpiresIn <= 0 {
			t.Fatalf("unexpected token metadata: type=%q expires_in=%d", resp.TokenType, resp.ExpiresIn)
		}
		if resp.User == nil || resp.User.Username != "dbuser" || resp.User.ID != userID {
			t.Fatalf("unexpected user payload: %+v", resp.User)
		}
	})

	t.Run("wrong password is rejected", func(t *testing.T) {
		rec := login(t, map[string]string{"username": "dbuser", "password": "definitely-wrong"})
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want 401 (body %q)", rec.Code, rec.Body.String())
		}
	})

	t.Run("unknown user is rejected", func(t *testing.T) {
		rec := login(t, map[string]string{"username": "nobody", "password": dbTestPassword})
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want 401 (body %q)", rec.Code, rec.Body.String())
		}
	})
}
