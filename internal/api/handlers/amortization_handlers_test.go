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
	"github.com/tahopetis/porta/internal/amortization"
	"github.com/tahopetis/porta/internal/api/middleware"
	"github.com/tahopetis/porta/internal/auth"
	"github.com/tahopetis/porta/pkg/logger"
)

// fakeAmortizationService embeds the interface and overrides every method with
// a minimal value, so handlers can be exercised without a database. Tests that
// care about a specific response set the corresponding field.
type fakeAmortizationService struct {
	amortization.Service

	list       *amortization.AmortizationCIList
	adjustment *amortization.LedgerEntry
}

func (f *fakeAmortizationService) GetAmortizationDetails(context.Context, uuid.UUID) (*amortization.AmortizationDetails, error) {
	return &amortization.AmortizationDetails{}, nil
}

func (f *fakeAmortizationService) UpdateAmortizationConfig(context.Context, uuid.UUID, *amortization.UpdateAmortizationConfig, uuid.UUID) (*amortization.AmortizationDetails, error) {
	return &amortization.AmortizationDetails{}, nil
}

func (f *fakeAmortizationService) ListAmortizableCIs(context.Context, *amortization.AmortizableCIFilters) (*amortization.AmortizationCIList, error) {
	if f.list != nil {
		return f.list, nil
	}
	return &amortization.AmortizationCIList{}, nil
}

func (f *fakeAmortizationService) GetAmortizationSettings(context.Context) (*amortization.AmortizationSettings, error) {
	return &amortization.AmortizationSettings{}, nil
}

func (f *fakeAmortizationService) UpdateAmortizationSettings(context.Context, *amortization.UpdateAmortizationSettings, uuid.UUID) (*amortization.AmortizationSettings, error) {
	return &amortization.AmortizationSettings{}, nil
}

func (f *fakeAmortizationService) GetLedgerEntries(context.Context, *amortization.LedgerFilters) (*amortization.LedgerEntryList, error) {
	return &amortization.LedgerEntryList{}, nil
}

func (f *fakeAmortizationService) GetLedgerEntry(context.Context, uuid.UUID) (*amortization.LedgerEntry, error) {
	return &amortization.LedgerEntry{}, nil
}

func (f *fakeAmortizationService) CreateAdjustment(_ context.Context, req *amortization.CreateAdjustmentRequest, _ uuid.UUID) (*amortization.LedgerEntry, error) {
	if f.adjustment != nil {
		return f.adjustment, nil
	}
	return &amortization.LedgerEntry{ID: uuid.New(), CIID: req.CIID, Amount: req.Amount}, nil
}

func (f *fakeAmortizationService) ProcessDailyAmortization(context.Context, time.Time) (*amortization.AmortizationRun, error) {
	return &amortization.AmortizationRun{}, nil
}

func (f *fakeAmortizationService) TriggerManualRun(context.Context, *amortization.ManualRunRequest, uuid.UUID) (*amortization.AmortizationRun, error) {
	return &amortization.AmortizationRun{}, nil
}

func (f *fakeAmortizationService) GetAmortizationRun(context.Context, uuid.UUID) (*amortization.AmortizationRun, error) {
	return &amortization.AmortizationRun{}, nil
}

func (f *fakeAmortizationService) ListAmortizationRuns(context.Context, *amortization.AmortizationRunFilters) (*amortization.AmortizationRunList, error) {
	return &amortization.AmortizationRunList{}, nil
}

func (f *fakeAmortizationService) GetAmortizationSummaries(context.Context, *amortization.SummaryRequest) (*amortization.AmortizationSummary, error) {
	return &amortization.AmortizationSummary{}, nil
}

func (f *fakeAmortizationService) GenerateDepreciationSchedule(context.Context, *amortization.DepreciationScheduleRequest) (*amortization.DepreciationScheduleResponse, error) {
	return &amortization.DepreciationScheduleResponse{}, nil
}

func (f *fakeAmortizationService) HandleTerminalStatusChange(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, uuid.UUID) error {
	return nil
}

func (f *fakeAmortizationService) RecalculateAmortization(context.Context, uuid.UUID, uuid.UUID) error {
	return nil
}

func (f *fakeAmortizationService) PreviewRestructuring(context.Context, uuid.UUID, int) (*amortization.RestructuringCalculation, error) {
	return &amortization.RestructuringCalculation{}, nil
}

func (f *fakeAmortizationService) RestructureAmortization(context.Context, *amortization.RestructureRequest, uuid.UUID) (*amortization.RestructureResult, error) {
	return &amortization.RestructureResult{}, nil
}

func newAmortizationTestRouter(t *testing.T, svc amortization.Service) *chi.Mux {
	t.Helper()
	router := chi.NewRouter()
	NewAmortizationHandler(svc, logger.New(logger.Config{Level: "disabled"})).RegisterRoutes(router)
	return router
}

func withAuthenticatedUser(t *testing.T, req *http.Request) *http.Request {
	t.Helper()
	claims := &auth.JWTClaims{UserID: uuid.New(), Username: "test-user"}
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, claims)
	return req.WithContext(ctx)
}

// perform issues one request against the router.
func perform(t *testing.T, router http.Handler, method, path string, body []byte, authenticated bool) *httptest.ResponseRecorder {
	t.Helper()
	var reader *bytes.Reader
	if body == nil {
		reader = bytes.NewReader(nil)
	} else {
		reader = bytes.NewReader(body)
	}
	req := httptest.NewRequest(method, path, reader)
	if authenticated {
		req = withAuthenticatedUser(t, req)
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func TestAmortizationReadRoutesReturnOK(t *testing.T) {
	router := newAmortizationTestRouter(t, &fakeAmortizationService{})
	id := uuid.NewString()

	routes := []string{
		"/amortization/configuration-items",
		"/amortization/configuration-items/" + id,
		"/amortization/ledger",
		"/amortization/ledger/" + id,
		"/amortization/runs",
		"/amortization/runs/" + id,
		"/amortization/summaries?group_by=ci_type",
		"/amortization/reports/depreciation-schedule?date_from=2026-01-01&date_to=2026-03-01",
		"/amortization/settings",
	}

	for _, path := range routes {
		t.Run(path, func(t *testing.T) {
			rec := perform(t, router, http.MethodGet, path, nil, false)
			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200 (body %q)", rec.Code, rec.Body.String())
			}
			var body map[string]any
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatalf("response is not JSON: %v (%q)", err, rec.Body.String())
			}
		})
	}
}

func TestAmortizationReadRoutesRejectInvalidIDs(t *testing.T) {
	router := newAmortizationTestRouter(t, &fakeAmortizationService{})

	for _, path := range []string{
		"/amortization/configuration-items/not-a-uuid",
		"/amortization/ledger/not-a-uuid",
		"/amortization/runs/not-a-uuid",
	} {
		t.Run(path, func(t *testing.T) {
			rec := perform(t, router, http.MethodGet, path, nil, false)
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400 (body %q)", rec.Code, rec.Body.String())
			}
		})
	}
}

func TestAmortizationWriteRoutesRequireAuthenticatedUser(t *testing.T) {
	router := newAmortizationTestRouter(t, &fakeAmortizationService{})
	id := uuid.NewString()

	writes := []struct {
		method string
		path   string
		body   string
	}{
		{http.MethodPut, "/amortization/configuration-items/" + id, `{"useful_life_months":24}`},
		{http.MethodPost, "/amortization/adjustments", `{"ci_id":"` + id + `","amount":25,"description":"fix"}`},
		{http.MethodPost, "/amortization/runs", `{}`},
		{http.MethodPut, "/amortization/settings", `{}`},
		{http.MethodPost, "/amortization/restructuring", `{"ci_id":"` + id + `","new_useful_life_months":12,"reason":"audit"}`},
	}

	for _, w := range writes {
		t.Run(w.method+" "+w.path, func(t *testing.T) {
			rec := perform(t, router, w.method, w.path, []byte(w.body), false)
			if rec.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want 401 (body %q)", rec.Code, rec.Body.String())
			}
		})
	}
}

func TestAmortizationWriteRoutesRejectMalformedInput(t *testing.T) {
	router := newAmortizationTestRouter(t, &fakeAmortizationService{})
	id := uuid.NewString()

	cases := []struct {
		name   string
		method string
		path   string
		body   string
	}{
		{"malformed adjustment body", http.MethodPost, "/amortization/adjustments", `{`},
		{"malformed config body", http.MethodPut, "/amortization/configuration-items/" + id, `{`},
		{"malformed run body", http.MethodPost, "/amortization/runs", `{`},
		{"malformed settings body", http.MethodPut, "/amortization/settings", `{`},
		{"preview missing ci id", http.MethodPost, "/amortization/restructuring/preview", `{"new_useful_life_months":12}`},
		{"preview missing months", http.MethodPost, "/amortization/restructuring/preview", `{"ci_id":"` + id + `"}`},
		{"restructure missing reason", http.MethodPost, "/amortization/restructuring", `{"ci_id":"` + id + `","new_useful_life_months":12}`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := perform(t, router, tc.method, tc.path, []byte(tc.body), true)
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400 (body %q)", rec.Code, rec.Body.String())
			}
		})
	}
}

func TestAmortizationWriteRoutesSucceedWithAuth(t *testing.T) {
	router := newAmortizationTestRouter(t, &fakeAmortizationService{})
	id := uuid.NewString()

	cases := []struct {
		name   string
		method string
		path   string
		body   string
		want   int
	}{
		{"update config", http.MethodPut, "/amortization/configuration-items/" + id, `{"useful_life_months":36}`, http.StatusOK},
		{"create adjustment", http.MethodPost, "/amortization/adjustments", `{"ci_id":"` + id + `","amount":25,"description":"fix"}`, http.StatusCreated},
		{"trigger manual run", http.MethodPost, "/amortization/runs", `{}`, http.StatusAccepted},
		{"update settings", http.MethodPut, "/amortization/settings", `{"currency":"USD"}`, http.StatusOK},
		{"preview restructuring", http.MethodPost, "/amortization/restructuring/preview", `{"ci_id":"` + id + `","new_useful_life_months":12}`, http.StatusOK},
		{"execute restructuring", http.MethodPost, "/amortization/restructuring", `{"ci_id":"` + id + `","new_useful_life_months":12,"reason":"audit"}`, http.StatusOK},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := perform(t, router, tc.method, tc.path, []byte(tc.body), true)
			if rec.Code != tc.want {
				t.Fatalf("status = %d, want %d (body %q)", rec.Code, tc.want, rec.Body.String())
			}
		})
	}
}

func TestListAmortizableCIsReturnsServiceResult(t *testing.T) {
	ciID := uuid.New()
	router := newAmortizationTestRouter(t, &fakeAmortizationService{
		list: &amortization.AmortizationCIList{
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

	rec := perform(t, router, http.MethodGet, "/amortization/configuration-items", nil, false)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body %q)", rec.Code, rec.Body.String())
	}

	var body struct {
		CIs []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"cis"`
		TotalCount int `json:"total_count"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("response is not JSON: %v", err)
	}
	if body.TotalCount != 1 || len(body.CIs) != 1 {
		t.Fatalf("total_count=%d cis=%d, want 1/1", body.TotalCount, len(body.CIs))
	}
	if body.CIs[0].ID != ciID.String() || body.CIs[0].Name != "laptop-01" {
		t.Fatalf("unexpected CI payload: %+v", body.CIs[0])
	}
}

func TestCreateAdjustmentReturnsCreatedEntry(t *testing.T) {
	entryID := uuid.New()
	router := newAmortizationTestRouter(t, &fakeAmortizationService{
		adjustment: &amortization.LedgerEntry{ID: entryID, EntryType: "adjustment", Amount: 25},
	})

	body := `{"ci_id":"` + uuid.NewString() + `","amount":25,"description":"fix"}`
	rec := perform(t, router, http.MethodPost, "/amortization/adjustments", []byte(body), true)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201 (body %q)", rec.Code, rec.Body.String())
	}

	var payload map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("response is not JSON: %v", err)
	}
	if payload["id"] != entryID.String() || payload["amount"] != float64(25) {
		t.Fatalf("unexpected entry payload: %v", payload)
	}
}
