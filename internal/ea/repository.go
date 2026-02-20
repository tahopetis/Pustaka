package ea

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository handles EA data access
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository creates a new EA repository
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// ============================================================================
// EA Teams CRUD
// ============================================================================

// CreateTeam creates a new EA team
func (r *Repository) CreateTeam(ctx context.Context, req *CreateEATeamRequest, createdBy uuid.UUID) (*EATeam, error) {
	team := &EATeam{
		ID:        uuid.New(),
		Name:      req.Name,
		Description: req.Description,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		CreatedBy: createdBy,
	}

	query := `
		INSERT INTO ea_teams (id, name, description, created_at, updated_at, created_by)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, name, description, created_at, updated_at, created_by
	`

	err := r.db.QueryRow(ctx, query,
		team.ID, team.Name, team.Description, team.CreatedAt, team.UpdatedAt, team.CreatedBy,
	).Scan(
		&team.ID, &team.Name, &team.Description, &team.CreatedAt, &team.UpdatedAt, &team.CreatedBy,
	)

	if err != nil {
			return nil, fmt.Errorf("failed to create EA team: %w", err)
	}

	return team, nil
}

// GetTeamByName retrieves an EA team by name
func (r *Repository) GetTeamByName(ctx context.Context, name string) (*EATeam, error) {
	team := &EATeam{}

	query := `
		SELECT id, name, description, created_at, updated_at, created_by
		FROM ea_teams
		WHERE name = $1
	`

	err := r.db.QueryRow(ctx, query, name).Scan(
		&team.ID, &team.Name, &team.Description, &team.CreatedAt, &team.UpdatedAt, &team.CreatedBy,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get EA team: %w", err)
	}

	return team, nil
}

// GetTeamByID retrieves an EA team by ID
func (r *Repository) GetTeamByID(ctx context.Context, id uuid.UUID) (*EATeam, error) {
	team := &EATeam{}

	query := `
		SELECT id, name, description, created_at, updated_at, created_by
		FROM ea_teams
		WHERE id = $1
	`

	err := r.db.QueryRow(ctx, query, id).Scan(
		&team.ID, &team.Name, &team.Description, &team.CreatedAt, &team.UpdatedAt, &team.CreatedBy,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get EA team: %w", err)
	}

	return team, nil
}

// ListTeams retrieves all EA teams
func (r *Repository) ListTeams(ctx context.Context) ([]*EATeam, error) {
	query := `
		SELECT id, name, description, created_at, updated_at, created_by
		FROM ea_teams
		ORDER BY name
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list EA teams: %w", err)
	}
	defer rows.Close()

	var teams []*EATeam
	for rows.Next() {
		team := &EATeam{}
		err := rows.Scan(
			&team.ID, &team.Name, &team.Description, &team.CreatedAt, &team.UpdatedAt, &team.CreatedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan EA team: %w", err)
		}
		teams = append(teams, team)
	}

	return teams, nil
}

// UpdateTeam updates an EA team
func (r *Repository) UpdateTeam(ctx context.Context, id uuid.UUID, req *UpdateEATeamRequest) (*EATeam, error) {
	team := &EATeam{}

	query := `
		UPDATE ea_teams
		SET name = COALESCE($2, name),
		    description = COALESCE($3, description),
		    updated_at = NOW()
		WHERE id = $1
		RETURNING id, name, description, created_at, updated_at, created_by
	`

	err := r.db.QueryRow(ctx, query, id, req.Name, req.Description).Scan(
		&team.ID, &team.Name, &team.Description, &team.CreatedAt, &team.UpdatedAt, &team.CreatedBy,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to update EA team: %w", err)
	}

	return team, nil
}

// DeleteTeam deletes an EA team
func (r *Repository) DeleteTeam(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM ea_teams WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete EA team: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrTeamNotFound
	}

	return nil
}

// ============================================================================
// CI Type Queries
// ============================================================================

// GetCITypeByName retrieves a CI type definition by name
func (r *Repository) GetCITypeByName(ctx context.Context, name string) (*CITypeDefinition, error) {
	ciType := &CITypeDefinition{}

	query := `
		SELECT id, name, description, required_attributes, optional_attributes,
		       created_at, updated_at, created_by
		FROM ci_type_definitions
		WHERE name = $1
	`

	var requiredAttrs, optionalAttrs []byte
	err := r.db.QueryRow(ctx, query, name).Scan(
		&ciType.ID, &ciType.Name, &ciType.Description, &requiredAttrs, &optionalAttrs,
		&ciType.CreatedAt, &ciType.UpdatedAt, &ciType.CreatedBy,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get CI type: %w", err)
	}

	// Unmarshal JSONB attributes
	if err := json.Unmarshal(requiredAttrs, &ciType.RequiredAttributes); err != nil {
		return nil, fmt.Errorf("failed to unmarshal required attributes: %w", err)
	}
	if err := json.Unmarshal(optionalAttrs, &ciType.OptionalAttributes); err != nil {
		return nil, fmt.Errorf("failed to unmarshal optional attributes: %w", err)
	}

	return ciType, nil
}

// ListEACITypes retrieves all EA CI type definitions
func (r *Repository) ListEACITypes(ctx context.Context) ([]*CITypeDefinition, error) {
	query := `
		SELECT id, name, description, required_attributes, optional_attributes,
		       created_at, updated_at, created_by
		FROM ci_type_definitions
		WHERE name LIKE 'EA.%'
		ORDER BY name
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list EA CI types: %w", err)
	}
	defer rows.Close()

	var types []*CITypeDefinition
	for rows.Next() {
		ciType := &CITypeDefinition{}
		var requiredAttrs, optionalAttrs []byte

		err := rows.Scan(
			&ciType.ID, &ciType.Name, &ciType.Description, &requiredAttrs, &optionalAttrs,
			&ciType.CreatedAt, &ciType.UpdatedAt, &ciType.CreatedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan CI type: %w", err)
		}

		// Unmarshal JSONB attributes
		if err := json.Unmarshal(requiredAttrs, &ciType.RequiredAttributes); err != nil {
			return nil, fmt.Errorf("failed to unmarshal required attributes: %w", err)
		}
		if err := json.Unmarshal(optionalAttrs, &ciType.OptionalAttributes); err != nil {
			return nil, fmt.Errorf("failed to unmarshal optional attributes: %w", err)
		}

		types = append(types, ciType)
	}

	return types, nil
}

// CITypeDefinition represents a CI type definition (reused from CI package)
type CITypeDefinition struct {
	ID                 uuid.UUID              `json:"id"`
	Name               string                 `json:"name"`
	Description        *string                `json:"description"`
	RequiredAttributes []AttributeDefinition  `json:"required_attributes"`
	OptionalAttributes []AttributeDefinition  `json:"optional_attributes"`
	CreatedAt          time.Time              `json:"created_at"`
	UpdatedAt          time.Time              `json:"updated_at"`
	CreatedBy          uuid.UUID              `json:"created_by"`
}

// AttributeDefinition represents an attribute definition in JSONB
type AttributeDefinition struct {
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	Description string                 `json:"description,omitempty"`
	Validation  map[string]interface{} `json:"validation,omitempty"`
}
