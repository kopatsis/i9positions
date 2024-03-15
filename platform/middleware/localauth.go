package middleware

import (
	"errors"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func GenerateJWT(id string) (string, error) {

	localIssuer := os.Getenv("LOCAL_ISSUER")

	claims := &jwt.RegisteredClaims{
		Issuer:  localIssuer,
		Subject: id,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	mySigningKey := []byte(os.Getenv("LOCAL_KEY"))

	tokenString, err := token.SignedString(mySigningKey)

	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func ValidateLocalToken(c *gin.Context) error {

	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return jwt.ErrInvalidType
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return jwt.ErrInvalidType
	}

	tokenString := parts[1]

	localIssuer := os.Getenv("LOCAL_ISSUER")
	mySigningKey := []byte(os.Getenv("LOCAL_KEY"))

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {

		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("Unexpected signing method: " + token.Header["alg"].(string))
		}

		return mySigningKey, nil
	})

	if err != nil {
		return err
	}

	if !token.Valid {
		return jwt.ErrInvalidKey
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return jwt.ErrTokenInvalidClaims
	}

	iss, ok := claims["iss"].(string)
	if !ok {
		return jwt.ErrTokenInvalidClaims
	}

	if iss != localIssuer {
		return jwt.ErrTokenInvalidIssuer
	}

	return nil
}

func TokenIsLocal(c *gin.Context) bool {

	localIssuer := os.Getenv("LOCAL_ISSUER")

	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return false
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return false
	}

	tokenString := parts[1]

	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return false
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return false
	}

	iss, ok := claims["iss"].(string)
	if !ok {
		return false
	}

	return iss == localIssuer
}
