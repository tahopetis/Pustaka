package api

import (
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
	// API versioning
	v1 := r.router.PathPrefix("/api/v1").Subrouter()

	// Health check
	v1.HandleFunc("/health", r.healthCheck).Methods("GET")

	// Configuration Items
	ci := v1.PathPrefix("/ci").Subrouter()
	ci.HandleFunc("", r.ciHandlers.CreateCI).Methods("POST")
	ci.HandleFunc("", r.ciHandlers.ListCIs).Methods("GET")
	ci.HandleFunc("/{id}", r.ciHandlers.GetCI).Methods("GET")
	ci.HandleFunc("/{id}", r.ciHandlers.UpdateCI).Methods("PUT")
	ci.HandleFunc("/{id}", r.ciHandlers.DeleteCI).Methods("DELETE")
	ci.HandleFunc("/{id}/relationships", r.ciHandlers.GetCIRelationships).Methods("GET")
	ci.HandleFunc("/{id}/network", r.ciHandlers.GetCINetwork).Methods("GET")
	ci.HandleFunc("/{id}/impact", r.ciHandlers.GetImpactAnalysis).Methods("GET")

	// CI Types
	types := v1.PathPrefix("/ci-types").Subrouter()
	types.HandleFunc("", r.typeHandlers.CreateCIType).Methods("POST")
	types.HandleFunc("", r.typeHandlers.ListCITypes).Methods("GET")
	types.HandleFunc("/{id}", r.typeHandlers.GetCIType).Methods("GET")
	types.HandleFunc("/{id}", r.typeHandlers.UpdateCIType).Methods("PUT")
	types.HandleFunc("/{id}", r.typeHandlers.DeleteCIType).Methods("DELETE")

	// Relationships
	relationships := v1.PathPrefix("/relationships").Subrouter()
	relationships.HandleFunc("", r.relHandlers.CreateRelationship).Methods("POST")
	relationships.HandleFunc("", r.relHandlers.ListRelationships).Methods("GET")
	relationships.HandleFunc("/{id}", r.relHandlers.GetRelationship).Methods("GET")
	relationships.HandleFunc("/{id}", r.relHandlers.UpdateRelationship).Methods("PUT")
	relationships.HandleFunc("/{id}", r.relHandlers.DeleteRelationship).Methods("DELETE")

	// Graph operations
	graph := v1.PathPrefix("/graph").Subrouter()
	graph.HandleFunc("", r.ciHandlers.GetGraphData).Methods("GET")

	// Analytics
	analytics := v1.PathPrefix("/analytics").Subrouter()
	analytics.HandleFunc("/cycles", r.relHandlers.FindCycles).Methods("GET")
	analytics.HandleFunc("/most-connected", r.relHandlers.GetMostConnectedCIs).Methods("GET")
	analytics.HandleFunc("/ci-types/usage", r.typeHandlers.GetCITypesByUsage).Methods("GET")
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