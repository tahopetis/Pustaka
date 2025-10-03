package api

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	pustakaLogger "github.com/pustaka/pustaka/pkg/logger"
	"github.com/pustaka/pustaka/internal/ci"
)

type Router struct {
	router      *mux.Router
	ciHandlers  *CIHandlers
	typeHandlers *CITypeHandlers
	relHandlers *RelationshipHandlers
}

func NewRouter(
	ciService *ci.Service,
	logger *pustakaLogger.Logger,
) *Router {
	router := mux.NewRouter()

	// Create handlers
	handler := NewHandler(logger)
	ciHandlers := NewCIHandlers(handler, ciService)
	typeHandlers := NewCITypeHandlers(handler, ciService)
	relHandlers := NewRelationshipHandlers(handler, ciService)

	r := &Router{
		router:       router,
		ciHandlers:   ciHandlers,
		typeHandlers: typeHandlers,
		relHandlers:  relHandlers,
	}

	r.setupRoutes()
	return r
}

func (r *Router) setupRoutes() {
	// Add debugging middleware to log all requests
	r.router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Log request path and variables
			vars := mux.Vars(r)
			fmt.Printf("DEBUG: Request path: %s, vars: %+v\n", r.URL.Path, vars)
			next.ServeHTTP(w, r)
		})
	})

	// Health check endpoint
	r.router.HandleFunc("/health", r.healthCheck).Methods("GET")

	// API versioning
	v1 := r.router.PathPrefix("/api/v1").Subrouter()

	// Configuration Items
	v1.HandleFunc("/ci", r.ciHandlers.CreateCI).Methods("POST")
	v1.HandleFunc("/ci", r.ciHandlers.ListCIs).Methods("GET")
	v1.HandleFunc("/ci/{id}", r.ciHandlers.GetCI).Methods("GET")
	v1.HandleFunc("/ci/{id}", r.ciHandlers.UpdateCI).Methods("PUT")
	v1.HandleFunc("/ci/{id}", r.ciHandlers.DeleteCI).Methods("DELETE")
	v1.HandleFunc("/ci/{id}/relationships", r.ciHandlers.GetCIRelationships).Methods("GET")
	v1.HandleFunc("/ci/{id}/network", r.ciHandlers.GetCINetwork).Methods("GET")
	v1.HandleFunc("/ci/{id}/impact", r.ciHandlers.GetImpactAnalysis).Methods("GET")

	// CI Types
	v1.HandleFunc("/ci-types", r.typeHandlers.CreateCIType).Methods("POST")
	v1.HandleFunc("/ci-types", r.typeHandlers.ListCITypes).Methods("GET")
	v1.HandleFunc("/ci-types/{id}", r.typeHandlers.GetCIType).Methods("GET")
	v1.HandleFunc("/ci-types/{id}", r.typeHandlers.UpdateCIType).Methods("PUT")
	v1.HandleFunc("/ci-types/{id}", r.typeHandlers.DeleteCIType).Methods("DELETE")

	// Relationships
	v1.HandleFunc("/relationships", r.relHandlers.CreateRelationship).Methods("POST")
	v1.HandleFunc("/relationships", r.relHandlers.ListRelationships).Methods("GET")
	v1.HandleFunc("/relationships/{id}", r.relHandlers.GetRelationship).Methods("GET")
	v1.HandleFunc("/relationships/{id}", r.relHandlers.UpdateRelationship).Methods("PUT")
	v1.HandleFunc("/relationships/{id}", r.relHandlers.DeleteRelationship).Methods("DELETE")

	// Graph operations
	graph := v1.PathPrefix("/graph").Subrouter()

	// Graph endpoints
	graph.HandleFunc("", r.ciHandlers.GetGraphData).Methods("GET")
	graph.HandleFunc("/explore", r.ciHandlers.ExploreGraph).Methods("GET")

	// CI network endpoint
	v1.HandleFunc("/ci/{id}/network", r.ciHandlers.GetCINetwork).Methods("GET")

	// Analytics
	analytics := v1.PathPrefix("/analytics").Subrouter()
	analytics.HandleFunc("/cycles", r.relHandlers.FindCycles).Methods("GET")
	analytics.HandleFunc("/most-connected", r.relHandlers.GetMostConnectedCIs).Methods("GET")
	analytics.HandleFunc("/ci-types/usage", r.typeHandlers.GetCITypesByUsage).Methods("GET")

	// Dashboard
	v1.HandleFunc("/dashboard/stats", r.ciHandlers.GetDashboardStats).Methods("GET")
}

func (r *Router) healthCheck(w http.ResponseWriter, req *http.Request) {
	r.ciHandlers.writeJSON(w, http.StatusOK, map[string]string{
		"status": "healthy",
		"service": "pustaka-cmdb-api",
		"version": "1.0.0",
	})
}

func (r *Router) GetRouter() *mux.Router {
	return r.router
}