package amortization

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/google/uuid"
	"github.com/go-chi/chi/v5"
	"github.com/pustaka/pustaka/internal/api/middleware"
)

// Fix for missing imports and function references in contract tests

// Fixed handler function routing for performance tests
func (suite *AmortizationContractTestSuite) routeRequest(handler http.Handler, req *http.Request) (*httptest.ResponseRecorder, error) {
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	return rr, nil
}

// Additional helper methods for testing

func (suite *AmortizationContractTestSuite) createAuthenticatedRequest(method, path string, userID uuid.UUID, permissions []string) *http.Request {
	req, _ := http.NewRequest(method, path, nil)

	// Create authenticated user context
	user := &middleware.AuthenticatedUser{
		UserID:      userID,
		Username:    "test-user",
		Email:       "test@example.com",
		Role:        "user",
		Permissions: permissions,
	}

	ctx := context.WithValue(req.Context(), middleware.UserContextKey, user)
	return req.WithContext(ctx)
}

// Add missing chi router for the handler tests
func (suite *AmortizationContractTestSuite) createTestRouter() chi.Router {
	r := chi.NewRouter()

	// Wrap handler with authentication middleware
	authMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := &middleware.AuthenticatedUser{
				UserID:      suite.adminUserID,
				Username:    "admin",
				Email:       "admin@example.com",
				Role:        "admin",
				Permissions: []string{"amortization:read", "amortization:write", "amortization:adjust", "amortization:admin"},
			}
			ctx := context.WithValue(r.Context(), middleware.UserContextKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}

	// Register routes
	r.Route("/amortization", func(r chi.Router) {
		r.Get("/configuration-items", suite.handler.ListAmortizableCIs)
		r.Get("/configuration-items/{ciId}", suite.handler.GetAmortizationDetails)
		r.Put("/configuration-items/{ciId}", suite.handler.UpdateAmortizationConfig)
		r.Get("/ledger", suite.handler.GetLedgerEntries)
		r.Get("/ledger/{entryId}", suite.handler.GetLedgerEntry)
		r.Post("/adjustments", suite.handler.CreateAdjustment)
		r.Get("/runs", suite.handler.ListAmortizationRuns)
		r.Get("/runs/{runId}", suite.handler.GetAmortizationRun)
		r.Post("/runs", suite.handler.TriggerManualRun)
		r.Get("/summaries", suite.handler.GetAmortizationSummaries)
		r.Get("/reports/depreciation-schedule", suite.handler.GenerateDepreciationSchedule)
	})

	return authMiddleware(r)
}