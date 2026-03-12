package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/xex-exchange/xexplay-api/internal/pkg/response"
)

// exchangeAdminValidateResponse is the response from Exchange /api/v1/admin/auth/validate.
type exchangeAdminValidateResponse struct {
	Data struct {
		Admin struct {
			ID    string `json:"id"`
			Email string `json:"email"`
			Role  string `json:"role"`
			Name  string `json:"name"`
		} `json:"admin"`
	} `json:"data"`
}

// ExchangeAdminAuth validates the admin session token against the Exchange API.
// This replaces the JWT-based admin auth — the admin panel authenticates via Exchange
// and sends the Exchange session token as Bearer token to the Play backend.
func ExchangeAdminAuth(exchangeAPIURL string) gin.HandlerFunc {
	client := &http.Client{Timeout: 5 * time.Second}

	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "missing authorization header")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			response.Unauthorized(c, "invalid authorization header format")
			c.Abort()
			return
		}

		sessionToken := parts[1]

		// Call Exchange API to validate admin session
		req, err := http.NewRequestWithContext(
			c.Request.Context(),
			http.MethodGet,
			fmt.Sprintf("%s/api/v1/admin/auth/validate", exchangeAPIURL),
			nil,
		)
		if err != nil {
			log.Error().Err(err).Msg("failed to create Exchange validate request")
			response.InternalError(c, "failed to validate admin session")
			c.Abort()
			return
		}
		req.Header.Set("Authorization", "Bearer "+sessionToken)

		resp, err := client.Do(req)
		if err != nil {
			log.Error().Err(err).Msg("failed to call Exchange admin validate")
			response.InternalError(c, "failed to validate admin session")
			c.Abort()
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			response.Unauthorized(c, "invalid or expired admin session")
			c.Abort()
			return
		}

		var validateResp exchangeAdminValidateResponse
		if err := json.NewDecoder(resp.Body).Decode(&validateResp); err != nil {
			log.Error().Err(err).Msg("failed to decode Exchange validate response")
			response.InternalError(c, "failed to parse admin session")
			c.Abort()
			return
		}

		admin := validateResp.Data.Admin

		// Set context values for downstream handlers
		adminID, _ := uuid.Parse(admin.ID)
		c.Set(ContextKeyUserID, adminID)
		c.Set(ContextKeyXexUserID, adminID)
		c.Set(ContextKeyEmail, admin.Email)
		c.Set(ContextKeyRole, "admin")
		c.Next()
	}
}
