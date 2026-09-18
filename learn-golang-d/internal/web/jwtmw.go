package web

import (
	"errors"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

const contextKeyUsername = "username"

// JWTMiddleware (C.32): tutorial SENGAJA tidak pakai middleware JWT bawaan
// framework, melainkan tulis sendiri - alasannya supaya proses ekstrak
// header, verifikasi signature, dan penyimpanan claim ke context semuanya
// eksplisit terlihat, bukan tersembunyi di balik satu baris middleware.JWT().
// Pola ini dipertahankan di sini untuk alasan belajar yang sama.
func JWTMiddleware(signingKey []byte) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			header := c.Request().Header.Get("Authorization")
			const prefix = "Bearer "
			if !strings.HasPrefix(header, prefix) {
				return echo.NewHTTPError(http.StatusUnauthorized, "header Authorization Bearer wajib diisi")
			}
			tokenString := strings.TrimPrefix(header, prefix)

			claims := &AppClaims{}
			token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, errors.New("signing method tidak dikenal")
				}
				return signingKey, nil
			})
			if err != nil || !token.Valid {
				return echo.NewHTTPError(http.StatusUnauthorized, "token tidak valid atau sudah kedaluwarsa")
			}

			c.Set(contextKeyUsername, claims.Username)
			return next(c)
		}
	}
}

func UsernameFromContext(c echo.Context) string {
	username, _ := c.Get(contextKeyUsername).(string)
	return username
}
