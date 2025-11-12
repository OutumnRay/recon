package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"Recontext.online/internal/models"
	"github.com/google/uuid"
)

// CreateDepartment godoc
// @Summary Create a new department
// @Description Create a new department in the organizational hierarchy
// @Tags Departments
// @Accept json
// @Produce json
// @Param request body models.CreateDepartmentRequest true "Department creation data"
// @Success 201 {object} models.Department
// @Failure 400 {object} models.ErrorResponse
// @Failure 409 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/departments [post]
func (mp *ManagingPortal) createDepartmentHandler(w http.ResponseWriter, r *http.Request) {
	var req models.CreateDepartmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		mp.respondWithError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Validate required fields
	if req.Name == "" {
		mp.respondWithError(w, http.StatusBadRequest, "Department name is required", "")
		return
	}

	// Check if department name already exists at the same level
	var parentIDStr *string
	if req.ParentID != nil {
		str := req.ParentID.String()
		parentIDStr = &str
	}
	exists, err := mp.departmentRepo.NameExists(req.Name, parentIDStr, "")
	if err != nil {
		mp.respondWithError(w, http.StatusInternalServerError, "Failed to check department name", err.Error())
		return
	}
	if exists {
		mp.respondWithError(w, http.StatusConflict, "Department with this name already exists at this level", "")
		return
	}

	// Create new department
	dept := &models.Department{
		ID:          uuid.New(),
		Name:        req.Name,
		Description: req.Description,
		ParentID:    req.ParentID,
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := mp.departmentRepo.Create(dept); err != nil {
		mp.respondWithError(w, http.StatusInternalServerError, "Failed to create department", err.Error())
		return
	}

	mp.logger.Infof("Department created: %s (%s)", dept.Name, dept.ID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(dept)
}

// ListDepartments godoc
// @Summary List departments
// @Description Get a list of departments with optional filters. Can return flat list or hierarchical tree.
// @Tags Departments
// @Produce json
// @Param parent_id query string false "Filter by parent department ID"
// @Param include_all query bool false "Include inactive departments" default(false)
// @Param tree query bool false "Return hierarchical tree structure" default(false)
// @Success 200 {array} models.Department
// @Success 200 {object} models.DepartmentTreeNode "When tree=true"
// @Security BearerAuth
// @Router /api/v1/departments [get]
func (mp *ManagingPortal) listDepartmentsHandler(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	parentIDStr := r.URL.Query().Get("parent_id")
	includeAll := r.URL.Query().Get("include_all") == "true"
	returnTree := r.URL.Query().Get("tree") == "true"

	var parentID *string
	if parentIDStr != "" {
		parentID = &parentIDStr
	}

	// Return tree structure if requested
	if returnTree {
		tree, err := mp.departmentRepo.GetTree(parentID)
		if err != nil {
			mp.respondWithError(w, http.StatusInternalServerError, "Failed to get department tree", err.Error())
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(tree)
		return
	}

	// Return flat list
	departments, err := mp.departmentRepo.List(parentID, includeAll)
	if err != nil {
		mp.respondWithError(w, http.StatusInternalServerError, "Failed to list departments", err.Error())
		return
	}

	// Convert to slice of values for standardized response
	var departmentsList []models.Department
	for _, dept := range departments {
		departmentsList = append(departmentsList, *dept)
	}

	// Return standardized response
	response := models.ListDepartmentsResponse{
		Items:    departmentsList,
		Total:    len(departmentsList),
		Offset:   0,
		PageSize: len(departmentsList),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetDepartment godoc
// @Summary Get department by ID
// @Description Get detailed information about a specific department including statistics
// @Tags Departments
// @Produce json
// @Param id path string true "Department ID"
// @Param stats query bool false "Include statistics (user count, child count)" default(false)
// @Success 200 {object} models.Department
// @Success 200 {object} models.DepartmentWithStats "When stats=true"
// @Failure 404 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/departments/{id} [get]
func (mp *ManagingPortal) getDepartmentHandler(w http.ResponseWriter, r *http.Request) {
	departmentID := strings.TrimPrefix(r.URL.Path, "/api/v1/departments/")
	includeStats := r.URL.Query().Get("stats") == "true"

	if includeStats {
		// Get department with statistics
		deptStats, err := mp.departmentRepo.GetWithStats(uuid.Must(uuid.Parse(departmentID)))
		if err != nil {
			mp.respondWithError(w, http.StatusNotFound, "Department not found", err.Error())
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(deptStats)
		return
	}

	// Get basic department info
	dept, err := mp.departmentRepo.GetByID(uuid.Must(uuid.Parse(departmentID)))
	if err != nil {
		mp.respondWithError(w, http.StatusNotFound, "Department not found", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dept)
}

// UpdateDepartment godoc
// @Summary Update department
// @Description Update department information (admin only). Automatically recalculates paths for all child departments.
// @Tags Departments
// @Accept json
// @Produce json
// @Param id path string true "Department ID"
// @Param request body models.UpdateDepartmentRequest true "Update data"
// @Success 200 {object} models.Department
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 409 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/departments/{id} [put]
func (mp *ManagingPortal) updateDepartmentHandler(w http.ResponseWriter, r *http.Request) {
	departmentID := strings.TrimPrefix(r.URL.Path, "/api/v1/departments/")

	var req models.UpdateDepartmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		mp.respondWithError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Find department
	dept, err := mp.departmentRepo.GetByID(uuid.Must(uuid.Parse(departmentID)))
	if err != nil {
		mp.respondWithError(w, http.StatusNotFound, "Department not found", err.Error())
		return
	}

	// Update fields
	if req.Name != "" {
		// Check if name already exists at the same level (excluding current department)
		var parentIDStr *string
		if dept.ParentID != nil {
			str := dept.ParentID.String()
			parentIDStr = &str
		}
		exists, err := mp.departmentRepo.NameExists(req.Name, parentIDStr, departmentID)
		if err != nil {
			mp.respondWithError(w, http.StatusInternalServerError, "Failed to check department name", err.Error())
			return
		}
		if exists {
			mp.respondWithError(w, http.StatusConflict, "Department with this name already exists at this level", "")
			return
		}
		dept.Name = req.Name
	}
	if req.Description != "" {
		dept.Description = req.Description
	}
	if req.ParentID != nil {
		dept.ParentID = req.ParentID
	}
	if req.IsActive != nil {
		dept.IsActive = *req.IsActive
	}

	dept.UpdatedAt = time.Now()

	// Update department (this will also update all child paths if parent changed)
	if err := mp.departmentRepo.Update(dept); err != nil {
		// Check if it's a circular reference error
		if strings.Contains(err.Error(), "circular reference") {
			mp.respondWithError(w, http.StatusBadRequest, "Cannot create circular reference", err.Error())
			return
		}
		mp.respondWithError(w, http.StatusInternalServerError, "Failed to update department", err.Error())
		return
	}

	mp.logger.Infof("Department updated: %s (%s)", dept.Name, dept.ID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dept)
}

// DeleteDepartment godoc
// @Summary Delete department
// @Description Soft delete a department (marks as inactive). Cannot delete department with active users.
// @Tags Departments
// @Produce json
// @Param id path string true "Department ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/departments/{id} [delete]
func (mp *ManagingPortal) deleteDepartmentHandler(w http.ResponseWriter, r *http.Request) {
	departmentID := strings.TrimPrefix(r.URL.Path, "/api/v1/departments/")

	// Check if department exists and get stats
	deptStats, err := mp.departmentRepo.GetWithStats(uuid.Must(uuid.Parse(departmentID)))
	if err != nil {
		mp.respondWithError(w, http.StatusNotFound, "Department not found", err.Error())
		return
	}

	// Prevent deletion if department has active users
	if deptStats.UserCount > 0 {
		mp.respondWithError(w, http.StatusBadRequest,
			fmt.Sprintf("Cannot delete department with %d active users. Please reassign users first.", deptStats.UserCount),
			"")
		return
	}

	// Soft delete (mark as inactive)
	if err := mp.departmentRepo.Delete(uuid.Must(uuid.Parse(departmentID))); err != nil {
		mp.respondWithError(w, http.StatusInternalServerError, "Failed to delete department", err.Error())
		return
	}

	mp.logger.Infof("Department deleted: %s", departmentID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message":       "Department deleted successfully",
		"department_id": departmentID,
	})
}

// GetDepartmentChildren godoc
// @Summary Get child departments
// @Description Get all direct child departments of a given department
// @Tags Departments
// @Produce json
// @Param id path string true "Parent Department ID"
// @Success 200 {array} models.Department
// @Failure 404 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/departments/{id}/children [get]
func (mp *ManagingPortal) getDepartmentChildrenHandler(w http.ResponseWriter, r *http.Request) {
	// Extract department ID from path like /api/v1/departments/{id}/children
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/departments/")
	departmentID := strings.TrimSuffix(path, "/children")

	// Verify parent department exists
	_, err := mp.departmentRepo.GetByID(uuid.Must(uuid.Parse(departmentID)))
	if err != nil {
		mp.respondWithError(w, http.StatusNotFound, "Department not found", err.Error())
		return
	}

	// Get children
	children, err := mp.departmentRepo.GetChildren(uuid.Must(uuid.Parse(departmentID)))
	if err != nil {
		mp.respondWithError(w, http.StatusInternalServerError, "Failed to get child departments", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(children)
}
