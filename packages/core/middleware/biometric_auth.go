package middleware

import (
	"errors"
	"net/http"
	"os"
	"strconv"

	"github.com/NasTecSol/nembus-core/repository"
	"github.com/NasTecSol/nembus-core/utils"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const PermissionBiometricsEnroll = "biometrics.enroll"

// ActorCanEnrollBiometrics returns the acting user ID when the caller holds biometrics.enroll
// via JWT session or a valid X-Authorization-Token scoped to biometrics.enroll.
func ActorCanEnrollBiometrics(c *gin.Context, repo *repository.Queries) (int32, bool) {
	if repo == nil {
		return 0, false
	}
	if supID, ok := parseAuthorizationTokenForPermission(c, PermissionBiometricsEnroll); ok {
		return supID, true
	}
	userIDStr := c.GetString("user_id")
	if userIDStr == "" {
		return 0, false
	}
	uid, err := strconv.ParseInt(userIDStr, 10, 32)
	if err != nil || uid <= 0 {
		return 0, false
	}
	has, err := repo.CheckUserHasPermission(c.Request.Context(), repository.CheckUserHasPermissionParams{
		UserID: int32(uid),
		Code:   PermissionBiometricsEnroll,
	})
	if err != nil || !has {
		return 0, false
	}
	return int32(uid), true
}

func parseAuthorizationTokenForPermission(c *gin.Context, permissionCode string) (int32, bool) {
	tokenString := c.GetHeader("X-Authorization-Token")
	if tokenString == "" {
		return 0, false
	}
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return 0, false
	}
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(jwtSecret), nil
	})
	if err != nil || !token.Valid {
		return 0, false
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, false
	}
	isAuthToken, _ := claims["is_authorization_token"].(bool)
	if !isAuthToken {
		return 0, false
	}
	tokenPerm, _ := claims["permission_code"].(string)
	if tokenPerm != permissionCode {
		return 0, false
	}
	if userID, ok := claims["authorized_by_user_id"].(float64); ok && int32(userID) > 0 {
		return int32(userID), true
	}
	return 0, false
}

// AbortUnlessBiometricEnrollPermission writes 403 and aborts when enrollment is not allowed.
func AbortUnlessBiometricEnrollPermission(c *gin.Context, repo *repository.Queries) (actorID int32, ok bool) {
	actorID, ok = ActorCanEnrollBiometrics(c, repo)
	if !ok {
		c.JSON(http.StatusForbidden, utils.NewResponse(utils.CodeForbidden,
			"biometric enrollment requires permission "+PermissionBiometricsEnroll+" or matching X-Authorization-Token", nil))
		c.Abort()
		return 0, false
	}
	return actorID, true
}
