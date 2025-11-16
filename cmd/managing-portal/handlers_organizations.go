package main

import (
	"encoding/json"
	"net/http"
	"strings"

	"Recontext.online/pkg/database"
	"github.com/google/uuid"
)

// GetOrganizationsHandler retrieves all organizations
func (mp *ManagingPortal) GetOrganizationsHandler(w http.ResponseWriter, r *http.Request) {
	orgRepo := database.NewOrganizationRepository(mp.db)
	organizations, err := orgRepo.GetAllOrganizations()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(organizations)
}

// GetOrganizationHandler retrieves a single organization by ID
func (mp *ManagingPortal) GetOrganizationHandler(w http.ResponseWriter, r *http.Request) {
	// Extract ID from path: /api/v1/organizations/{id}
	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/v1/organizations/"), "/")
	if len(pathParts) == 0 || pathParts[0] == "" {
		http.Error(w, "Organization ID is required", http.StatusBadRequest)
		return
	}

	id, err := uuid.Parse(pathParts[0])
	if err != nil {
		http.Error(w, "Invalid organization ID", http.StatusBadRequest)
		return
	}

	orgRepo := database.NewOrganizationRepository(mp.db)
	organization, err := orgRepo.GetOrganizationByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(organization)
}

// CreateOrganizationRequest represents the request body for creating an organization
type CreateOrganizationRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Domain      *string `json:"domain"`
	LogoURL     *string `json:"logo_url"`
	IsActive    bool    `json:"is_active"`
	Settings    string  `json:"settings"`
}

// CreateOrganizationHandler creates a new organization
func (mp *ManagingPortal) CreateOrganizationHandler(w http.ResponseWriter, r *http.Request) {
	var req CreateOrganizationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "Organization name is required", http.StatusBadRequest)
		return
	}

	org := &database.Organization{
		Name:        req.Name,
		Description: req.Description,
		Domain:      req.Domain,
		LogoURL:     req.LogoURL,
		IsActive:    req.IsActive,
		Settings:    req.Settings,
	}

	orgRepo := database.NewOrganizationRepository(mp.db)
	if err := orgRepo.CreateOrganization(org); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(org)
}

// UpdateOrganizationRequest represents the request body for updating an organization
type UpdateOrganizationRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Domain      *string `json:"domain"`
	LogoURL     *string `json:"logo_url"`
	IsActive    bool    `json:"is_active"`
	Settings    string  `json:"settings"`
}

// UpdateOrganizationHandler updates an existing organization
func (mp *ManagingPortal) UpdateOrganizationHandler(w http.ResponseWriter, r *http.Request) {
	// Extract ID from path: /api/v1/organizations/{id}
	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/v1/organizations/"), "/")
	if len(pathParts) == 0 || pathParts[0] == "" {
		http.Error(w, "Organization ID is required", http.StatusBadRequest)
		return
	}

	id, err := uuid.Parse(pathParts[0])
	if err != nil {
		http.Error(w, "Invalid organization ID", http.StatusBadRequest)
		return
	}

	var req UpdateOrganizationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "Organization name is required", http.StatusBadRequest)
		return
	}

	org := &database.Organization{
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
		Domain:      req.Domain,
		LogoURL:     req.LogoURL,
		IsActive:    req.IsActive,
		Settings:    req.Settings,
	}

	orgRepo := database.NewOrganizationRepository(mp.db)
	if err := orgRepo.UpdateOrganization(org); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(org)
}

// DeleteOrganizationHandler deletes an organization
func (mp *ManagingPortal) DeleteOrganizationHandler(w http.ResponseWriter, r *http.Request) {
	// Extract ID from path: /api/v1/organizations/{id}
	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/v1/organizations/"), "/")
	if len(pathParts) == 0 || pathParts[0] == "" {
		http.Error(w, "Organization ID is required", http.StatusBadRequest)
		return
	}

	id, err := uuid.Parse(pathParts[0])
	if err != nil {
		http.Error(w, "Invalid organization ID", http.StatusBadRequest)
		return
	}

	orgRepo := database.NewOrganizationRepository(mp.db)
	if err := orgRepo.DeleteOrganization(id); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetOrganizationStatsHandler retrieves statistics for an organization
func (mp *ManagingPortal) GetOrganizationStatsHandler(w http.ResponseWriter, r *http.Request) {
	// Extract ID from path: /api/v1/organizations/{id}/stats
	pathAfterOrgs := strings.TrimPrefix(r.URL.Path, "/api/v1/organizations/")
	pathParts := strings.Split(pathAfterOrgs, "/")
	if len(pathParts) < 2 || pathParts[0] == "" {
		http.Error(w, "Organization ID is required", http.StatusBadRequest)
		return
	}

	id, err := uuid.Parse(pathParts[0])
	if err != nil {
		http.Error(w, "Invalid organization ID", http.StatusBadRequest)
		return
	}

	orgRepo := database.NewOrganizationRepository(mp.db)
	stats, err := orgRepo.GetOrganizationStats(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
