package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"time"

	"Recontext.online/internal/models"
	"Recontext.online/pkg/email"
	"golang.org/x/crypto/bcrypt"
)

const (
	resetCodeLength     = 6
	resetCodeExpiration = 15 * time.Minute
)

// RequestPasswordReset godoc
// @Summary Request password reset
// @Description Request a password reset code to be sent via email
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body models.RequestPasswordResetRequest true "Email address"
// @Success 200 {object} models.RequestPasswordResetResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/auth/password-reset/request [post]
func (up *UserPortal) requestPasswordResetHandler(w http.ResponseWriter, r *http.Request) {
	var req models.RequestPasswordResetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		up.respondWithError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Find user by email
	user, err := up.userRepo.GetByEmail(req.Email)
	if err != nil {
		// Don't reveal if email exists or not for security
		up.logger.Infof("Password reset requested for non-existent email: %s", req.Email)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(models.RequestPasswordResetResponse{
			Message: "If the email exists, a reset code has been sent",
			TokenID: "dummy-token-id",
		})
		return
	}

	if !user.IsActive {
		up.respondWithError(w, http.StatusForbidden, "Account is inactive", "Cannot reset password for inactive account")
		return
	}

	// Generate 6-digit code
	code, err := generateResetCode()
	if err != nil {
		up.respondWithError(w, http.StatusInternalServerError, "Failed to generate reset code", err.Error())
		return
	}

	// Create reset token
	tokenID := generateUUID()
	expiresAt := time.Now().Add(resetCodeExpiration)

	query := `
		INSERT INTO password_reset_tokens (id, user_id, email, code, expires_at, used, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err = up.db.Exec(query, tokenID, user.ID, user.Email, code, expiresAt, false, time.Now())
	if err != nil {
		up.respondWithError(w, http.StatusInternalServerError, "Failed to create reset token", err.Error())
		return
	}

	// Send email with code
	resetEmailData := email.PasswordResetEmailData{
		Email:    user.Email,
		Code:     code,
		Language: user.Language,
	}
	err = up.emailService.SendPasswordResetEmail(user.Email, resetEmailData)
	if err != nil {
		up.logger.Errorf("Failed to send reset email to %s: %v", user.Email, err)
		// Don't fail the request, token is created
	}

	up.logger.Infof("Password reset code sent to %s (Token ID: %s)", user.Email, tokenID)

	response := models.RequestPasswordResetResponse{
		Message: "Password reset code sent to your email",
		TokenID: tokenID,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// VerifyResetCode godoc
// @Summary Verify password reset code
// @Description Verify that the provided reset code is valid
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body models.VerifyResetCodeRequest true "Token ID and code"
// @Success 200 {object} models.VerifyResetCodeResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/v1/auth/password-reset/verify [post]
func (up *UserPortal) verifyResetCodeHandler(w http.ResponseWriter, r *http.Request) {
	var req models.VerifyResetCodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		up.respondWithError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Find token
	var token models.PasswordResetToken
	query := `
		SELECT id, user_id, email, code, expires_at, used, created_at
		FROM password_reset_tokens
		WHERE id = $1
	`

	err := up.db.QueryRow(query, req.TokenID).Scan(
		&token.ID,
		&token.UserID,
		&token.Email,
		&token.Code,
		&token.ExpiresAt,
		&token.Used,
		&token.CreatedAt,
	)

	if err == sql.ErrNoRows {
		up.respondWithError(w, http.StatusNotFound, "Invalid token", "Token not found")
		return
	}

	if err != nil {
		up.respondWithError(w, http.StatusInternalServerError, "Database error", err.Error())
		return
	}

	// Validate token
	if token.Used {
		up.respondWithError(w, http.StatusBadRequest, "Token already used", "This reset code has already been used")
		return
	}

	if time.Now().After(token.ExpiresAt) {
		up.respondWithError(w, http.StatusBadRequest, "Token expired", "This reset code has expired")
		return
	}

	if token.Code != req.Code {
		up.respondWithError(w, http.StatusBadRequest, "Invalid code", "The provided code is incorrect")
		return
	}

	response := models.VerifyResetCodeResponse{
		Valid:   true,
		Message: "Code verified successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ResetPassword godoc
// @Summary Reset password with code
// @Description Reset user password using verified code
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body models.ResetPasswordRequest true "Reset password data"
// @Success 200 {object} models.ResetPasswordResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/auth/password-reset/reset [post]
func (up *UserPortal) resetPasswordHandler(w http.ResponseWriter, r *http.Request) {
	var req models.ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		up.respondWithError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Find and validate token
	var token models.PasswordResetToken
	query := `
		SELECT id, user_id, email, code, expires_at, used, created_at
		FROM password_reset_tokens
		WHERE id = $1
	`

	err := up.db.QueryRow(query, req.TokenID).Scan(
		&token.ID,
		&token.UserID,
		&token.Email,
		&token.Code,
		&token.ExpiresAt,
		&token.Used,
		&token.CreatedAt,
	)

	if err == sql.ErrNoRows {
		up.respondWithError(w, http.StatusNotFound, "Invalid token", "Token not found")
		return
	}

	if err != nil {
		up.respondWithError(w, http.StatusInternalServerError, "Database error", err.Error())
		return
	}

	// Validate token
	if token.Used {
		up.respondWithError(w, http.StatusBadRequest, "Token already used", "This reset code has already been used")
		return
	}

	if time.Now().After(token.ExpiresAt) {
		up.respondWithError(w, http.StatusBadRequest, "Token expired", "This reset code has expired")
		return
	}

	if token.Code != req.Code {
		up.respondWithError(w, http.StatusBadRequest, "Invalid code", "The provided code is incorrect")
		return
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		up.respondWithError(w, http.StatusInternalServerError, "Failed to hash password", err.Error())
		return
	}

	// Update user password
	updateQuery := `
		UPDATE users
		SET password = $1, updated_at = $2
		WHERE id = $3
	`

	_, err = up.db.Exec(updateQuery, string(hashedPassword), time.Now(), token.UserID)
	if err != nil {
		up.respondWithError(w, http.StatusInternalServerError, "Failed to update password", err.Error())
		return
	}

	// Mark token as used
	markUsedQuery := `
		UPDATE password_reset_tokens
		SET used = TRUE
		WHERE id = $1
	`

	_, err = up.db.Exec(markUsedQuery, token.ID)
	if err != nil {
		up.logger.Errorf("Failed to mark token as used: %v", err)
		// Don't fail the request, password is already updated
	}

	up.logger.Infof("Password reset successfully for user %s", token.UserID)

	response := models.ResetPasswordResponse{
		Message: "Password reset successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// generateResetCode generates a 6-digit random code
func generateResetCode() (string, error) {
	code := ""
	for i := 0; i < resetCodeLength; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		code += fmt.Sprintf("%d", n.Int64())
	}
	return code, nil
}

// generateUUID generates a simple UUID (replace with proper UUID library if needed)
func generateUUID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}
