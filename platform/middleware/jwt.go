package middleware

import (
	"context"
	"errors"
	"net/http"
	"os"
	"slices"
	"strings"

	firebase "firebase.google.com/go"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

// var ExpectedAudience = os.Getenv("AUTH0_AUDIENCE")

func interfaceSliceToStringSlice(s []interface{}) ([]string, error) {
	var stringSlice []string
	for _, v := range s {
		str, ok := v.(string)
		if !ok {
			return nil, errors.New("value is not a string")
		}
		stringSlice = append(stringSlice, str)
	}
	return stringSlice, nil
}

func JWTAuthMiddleware(firebase *firebase.App) gin.HandlerFunc {
	return func(c *gin.Context) {

		if c.Request.Method != "POST" {
			c.Next()
			return
		}

		if TokenIsLocal(c) {
			if err := ValidateLocalToken(c); err != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Local token error: " + err.Error()})
				return
			}
		} else {

			// if !OldAuth(c) {
			// 	return
			// }
			if !verifyToken(firebase, c) {
				return
			}

		}
		c.Next()
	}
}

func OldAuth(c *gin.Context) bool {
	ExpectedAudience := os.Getenv("AUTH0_AUDIENCE")

	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
		return false
	}

	splitToken := strings.Split(authHeader, "Bearer ")
	if len(splitToken) != 2 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid Authorization header format"})
		return false
	}

	tokenString := splitToken[1]

	token, err := jwt.Parse(tokenString, getKeyFunc())
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token: " + err.Error()})
		return false
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		audience := claims["aud"]
		typed := false

		if aud, ok := audience.(string); ok {
			if aud != ExpectedAudience {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token audience"})
				return false
			}
			typed = true
		} else if aud, ok := audience.([]any); ok {
			sliced, err := interfaceSliceToStringSlice(aud)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token audience type"})
				return false
			}
			if !slices.Contains(sliced, ExpectedAudience) {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token audience"})
				return false
			}
			typed = true
		}

		if !typed {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token audience type"})
			return false
		}

	} else {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
		return false
	}
	return true
}

func verifyToken(app *firebase.App, c *gin.Context) bool {

	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
		return false
	}

	splitToken := strings.Split(authHeader, "Bearer ")
	if len(splitToken) != 2 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid Authorization header format"})
		return false
	}

	idToken := splitToken[1]

	client, err := app.Auth(context.TODO())
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Error getting Auth client: " + err.Error()})
		return false
	}

	_, err = client.VerifyIDToken(context.TODO(), idToken)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Error verifying ID token: " + err.Error()})
		return false
	}

	return true
}
