package auth

import (
	"fmt"
	"net/http"
	"strings"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

func Protect(signature string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		s := ctx.Request.Header.Get("Authorization")
		tokenString := strings.TrimPrefix(s, "Bearer ")

		token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Unexpected signing method: %v", t.Header["alg"])
			}

			return []byte(signature), nil
		})

		if err != nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			aud := claims["aud"]
			ctx.Set("aud", aud)
		}

		ctx.Next()
	}
}
