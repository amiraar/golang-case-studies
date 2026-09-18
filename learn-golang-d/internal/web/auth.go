package web

import (
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

// AppClaims (C.32 JSON Web Token): tutorial mendefinisikan struct claim
// custom yang meng-embed jwt.RegisteredClaims - dipakai di sini juga,
// menyimpan Username supaya endpoint terproteksi tahu siapa yang login
// tanpa query ulang ke store (mirip UsernameFromContext di learn-golang-c,
// tapi sumbernya token, bukan hasil Basic Auth per-request).
type AppClaims struct {
	jwt.RegisteredClaims
	Username string `json:"username"`
}

type AuthHandler struct {
	username      string
	password      string
	signingKey    []byte
	expiryMinutes int
}

func NewAuthHandler(username, password string, signingKey []byte, expiryMinutes int) *AuthHandler {
	return &AuthHandler{username: username, password: password, signingKey: signingKey, expiryMinutes: expiryMinutes}
}

type loginPayload struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type loginResponse struct {
	Token string `json:"token"`
}

// Login (C.4 Bind + C.5 Validate + C.32 JWT): POST /login. Kredensial
// dibandingkan ke satu akun demo dari config (sama seperti admin/admin123
// di learn-golang-c B.18) - project ini fokus belajar mekanisme JWT-nya,
// bukan membangun sistem user management penuh.
func (h *AuthHandler) Login(c echo.Context) error {
	var payload loginPayload
	if err := c.Bind(&payload); err != nil {
		return err
	}
	if err := c.Validate(&payload); err != nil {
		return err
	}

	if payload.Username != h.username || payload.Password != h.password {
		return echo.NewHTTPError(http.StatusUnauthorized, "username atau password salah")
	}

	claims := AppClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(h.expiryMinutes) * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		Username: payload.Username,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(h.signingKey)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, loginResponse{Token: signed})
}
