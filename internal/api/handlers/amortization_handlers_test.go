package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/tahopetis/porta/internal/amortization"
	"github.com/tahopetis/porta/internal/api/middleware"
	"github.com/tahopetis/porta/internal/auth"
	"github.com/tahopetis/porta/pkg/logger"
)

// fakeAmortizationService embeds the interface and overrides only the methods a
// test drives, so the suite exercises real handler behaviour without a database.
type fakeAmortizationService struct {
	amortization.Service

	listResult *amortization.AmortizationCIList
	adjustment *amortization.LedgerEntry
	createErr  error
}

func (f *fakeAmortizationService) ListAmortizableCIs(context.Context, *amortization.AmortizableCIFilters) (*amortization.AmortizationCIList, error) {
	return f.listResult, nil
}

func (f *fakeAmortizationService) CreateAdjustment(_ context.Context, req *amortization.CreateAdjustmentRequest, userID uuid.UUID) (*amortization.LedgerEntry, error) {
	if f.createErr != nil {
		return nil, f.createErr
	}
	entry := f.adjustment
	if entry == nil {
		entry = &amortization.LedgerEntry{ID: uuid.New(), CIID: req.CIID, Amount: req.Amount}
	}
	return entry, nil
}

func newAmortizationTestRouter(t *testing.T, svc amortization.Service) *chi.Mux {
	t.Helper()
	handler := NewAmortizationHandler(svc, logger.Default())
	router := chi.NewRouter()
	handler.RegisterRoutes(router)
	return router
}

func withAuthenticatedUser(t *testing.T, req *http.Request) *http.Request {
	t.Helper()
	claims := &auth.JWTClaims{UserID: uuid.New(), Username: "test-user"}
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, claims)
	return req.WithContext(ctx)
}

func decodeBody(t *testing.T, rec *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("response is not JSON: %v (body %q)", err, rec.Body.String())
	}
	return body
}

func TestListAmortizableCIsReturnsServiceResult(t *testing.T) {
	ciID := uuid.New()
	router := newAmortizationTestRouter(t, &fakeAmortizationService{
		listResult: &amortization.AmortizationCIList{
			CIs: []amortization.AmortizableCI{{
				ID:               ciID,
				Name:             "laptop-01",
				CIType:           "Hardware",
				PurchaseCost:     1200,
				UsefulLifeMonths: 24,
				CurrentBookValue: 900,
			}},
			TotalCount: 1,
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/amortization/configuration-items", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body %q)", rec.Code, rec.Body.String())
	}

	body := decodeBody(t, rec)
	if got := body["total_count"]; got != float64(1) {
		t.Fatalf("total_count = %v, want 1", got)
	}
	cis, ok := body["cis"].([]any)
	if !ok || len(cis) != 1 {
		t.Fatalf("cis = %v, want one entry", body["cis"])
	}
	first, _ := cis[0].(map[string]any)
	if got := first["name"]; got != "laptop-01" {
		t.Fatalf("cis[0].name = %v, want laptop-01", got)
	}
	if got := first["id"]; got != ciID.String() {
		t.Fatalf("cis[0].id = %v, want %s", got, ciID)
	}
}

func TestCreateAdjustmentRequiresAuthenticatedUser(t *testing.T) {
	router := newAmortizationTestRouter(t, &fakeAmortizationService{})
	payload := []byte(`{"ci_id":"` + uuid.NewString() + `","amount":25,"description":"correction"}`)
	req := httptest.NewRequest(http.MethodPost, "/amortization/adjustments", bytes.NewReader(payload))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401 (body %q)", rec.Code, rec.Body.String())
	}
}

func TestCreateAdjustmentRejectsMalformedBody(t *testing.T) {
	router := newAmortizationTestRouter(t, &fakeAmortizationService{})
	req := httptest.NewRequest(http.MethodPost, "/amortization/adjustments", bytes.NewReader([]byte(`{`)))
	req = withAuthenticatedUser(t, req)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400 (body %q)", rec.Code, rec.Body.String())
	}
}

func TestCreateAdjustmentReturnsCreated(t *testing.T) {
	ciID := uuid.New()
	entryID := uuid.New()
	router := newAmortizationTestRouter(t, &fakeAmortizationService{
		adjustment: &amortization.LedgerEntry{
			ID:        entryID,
			CIID:      ciID,
			EntryType: "adjustment",
			Amount:    25,
		},
	})

	payload := []byte(`{"ci_id":"` + ciID.String() + `","amount":25,"description":"correction"}`)
	req := httptest.NewRequest(http.MethodPost, "/amortization/adjustments", bytes.NewReader(payload))
	req = withAuthenticatedUser(t, req)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201 (body %q)", rec.Code, rec.Body.String())
	}

	body := decodeBody(t, rec)
	if got := body["id"]; got != entryID.String() {
		t.Fatalf("id = %v, want %s", got, entryID)
	}
	if got := body["amount"]; got != float64(25) {
		t.Fatalf("amount = %v, want 25", got)
	}
}

func TestGetAmortizationDetailsRejectsInvalidID(t *testing.T) {
	router := newAmortizationTestRouter(t, &fakeAmortizationService{})
	req := httptest.NewRequest(http.MethodGet, "/amortization/configuration-items/not-a-uuid", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400 (body %q)", rec.Code, rec.Body.String())
	}
}
