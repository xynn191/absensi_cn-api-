package middleware

import (
	"net/http"
	"strings"

	"absensi-cn-api/pkg/response"
	"absensi-cn-api/pkg/token"

	"github.com/gin-gonic/gin"
)

func RequireRoles(roles ...string) gin.HandlerFunc {
	allowedRoles := make(map[string]struct{}, len(roles))
	for _, role := range roles {
		allowedRoles[role] = struct{}{}
	}

	return func(c *gin.Context) {
		rawRole, exists := c.Get("auth_user_role")
		if !exists {
			response.Error(c, http.StatusForbidden, "missing authenticated role")
			c.Abort()
			return
		}

		role, ok := rawRole.(string)
		if !ok {
			response.Error(c, http.StatusForbidden, "invalid authenticated role")
			c.Abort()
			return
		}

		if _, allowed := allowedRoles[role]; !allowed {
			response.Error(c, http.StatusForbidden, "forbidden for this role")
			c.Abort()
			return
		}

		c.Next()
	}
}

func RequireAuth(jwtManager *token.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Error(c, http.StatusUnauthorized, "missing authorization header")
			c.Abort()
			return
		}

		tokenString := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer"))
		if tokenString == "" {
			response.Error(c, http.StatusUnauthorized, "invalid bearer token")
			c.Abort()
			return
		}

		claims, err := jwtManager.Verify(tokenString)
		if err != nil {
			response.Error(c, http.StatusUnauthorized, "invalid or expired token")
			c.Abort()
			return
		}

		c.Set("auth_user_id", claims.UserID)
		c.Set("auth_user_role", claims.Role)
		c.Next()
	}
}
