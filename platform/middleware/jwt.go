package middleware

import (
	"errors"
	"net/http"
	"os"
	"slices"
	"strings"

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

func JWTAuthMiddleware() gin.HandlerFunc {
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

			ExpectedAudience := os.Getenv("AUTH0_AUDIENCE")

			authHeader := c.GetHeader("Authorization")
			if authHeader == "" {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
				return
			}

			splitToken := strings.Split(authHeader, "Bearer ")
			if len(splitToken) != 2 {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid Authorization header format"})
				return
			}

			tokenString := splitToken[1]

			token, err := jwt.Parse(tokenString, getKeyFunc())
			if err != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token: " + err.Error()})
				return
			}

			if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
				audience := claims["aud"]
				typed := false

				if aud, ok := audience.(string); ok {
					if aud != ExpectedAudience {
						c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token audience"})
						return
					}
					typed = true
				} else if aud, ok := audience.([]any); ok {
					sliced, err := interfaceSliceToStringSlice(aud)
					if err != nil {
						c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token audience type"})
						return
					}
					if !slices.Contains(sliced, ExpectedAudience) {
						c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token audience"})
						return
					}
					typed = true
				}

				if !typed {
					c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token audience type"})
					return
				}

			} else {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
				return
			}
		}
		c.Next()
	}
}
