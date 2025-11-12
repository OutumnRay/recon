package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"time"

	"github.com/google/uuid"
	"Recontext.online/internal/models"
	"Recontext.online/pkg/database"
	"Recontext.online/pkg/email"
	"golang.org/x/crypto/bcrypt"
)

const (
	resetCodeLength     = 6
	resetCodeExpiration = 15 * time.Minute
)

// RequestPasswordReset godoc
// @Summary Запросить сброс пароля
// @Description Запросить код сброса пароля для отправки по электронной почте
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
			TokenID: uuid.Nil,
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
	token := &database.PasswordResetToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		Email:     user.Email,
		Code:      code,
		ExpiresAt: time.Now().Add(resetCodeExpiration),
		Used:      false,
		CreatedAt: time.Now(),
	}

	if err := up.db.DB.Create(token).Error; err != nil {
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

	up.logger.Infof("Password reset code sent to %s (Token ID: %s)", user.Email, token.ID)

	response := models.RequestPasswordResetResponse{
		Message: "Password reset code sent to your email",
		TokenID: token.ID,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// VerifyResetCode godoc
// @Summary Проверить код сброса пароля
// @Description Проверить, что предоставленный код сброса действителен
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
	var token database.PasswordResetToken
	if err := up.db.DB.Where("id = ?", req.TokenID.String()).First(&token).Error; err != nil {
		up.respondWithError(w, http.StatusNotFound, "Invalid token", "Token not found")
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
// @Summary Сбросить пароль с кодом
// @Description Сбросить пароль пользователя с использованием проверенного кода
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
	var token database.PasswordResetToken
	if err := up.db.DB.Where("id = ?", req.TokenID.String()).First(&token).Error; err != nil {
		up.respondWithError(w, http.StatusNotFound, "Invalid token", "Token not found")
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
	up.logger.Infof("=== PASSWORD RESET DEBUG ===")
	up.logger.Infof("User ID: %s", token.UserID)
	up.logger.Infof("Email: %s", token.Email)
	up.logger.Infof("New password (plain): %s", req.NewPassword)

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		up.respondWithError(w, http.StatusInternalServerError, "Failed to hash password", err.Error())
		return
	}

	up.logger.Infof("New password hash (first 20 chars): %s...", string(hashedPassword[:20]))

	// Update user password
	up.logger.Infof("Executing UPDATE query for user: %s", token.UserID)
	if err := up.db.DB.Model(&database.User{}).Where("id = ?", token.UserID).Updates(map[string]interface{}{
		"password":   string(hashedPassword),
		"updated_at": time.Now(),
	}).Error; err != nil {
		up.logger.Errorf("Failed to update password in database: %v", err)
		up.respondWithError(w, http.StatusInternalServerError, "Failed to update password", err.Error())
		return
	}

	up.logger.Infof("✓ Password updated in database")
	up.logger.Infof("============================")

	// Mark token as used
	if err := up.db.DB.Model(&token).Update("used", true).Error; err != nil {
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
