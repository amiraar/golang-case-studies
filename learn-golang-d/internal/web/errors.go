package web

import (
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

// httpErrorResponse: bentuk error JSON yang konsisten untuk SEMUA jenis
// error (echo.HTTPError bawaan, error validasi, atau error biasa) - klien
// API tidak perlu membedakan sumber error, cukup baca field ini.
type httpErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// ErrorHandler (C.6 HTTP Error Handling): dipasang lewat e.HTTPErrorHandler
// di main.go, menggantikan default handler Echo. Dua kasus dibedakan
// eksplisit persis pola tutorial: (1) error validasi (validator.
// ValidationErrors) diterjemahkan jadi pesan per-field yang bisa dibaca
// manusia, (2) error lain dibungkus jadi *echo.HTTPError generik.
func ErrorHandler(err error, c echo.Context) {
	if c.Response().Committed {
		return
	}

	report, ok := err.(*echo.HTTPError)
	if !ok {
		report = echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if validationErrs, ok := err.(validator.ValidationErrors); ok {
		fieldErr := validationErrs[0]
		report = echo.NewHTTPError(http.StatusBadRequest,
			fmt.Sprintf("%s gagal validasi: %s", fieldErr.Field(), fieldErr.Tag()))
	}

	message := fmt.Sprint(report.Message)
	if err := c.JSON(report.Code, httpErrorResponse{Code: report.Code, Message: message}); err != nil {
		c.Logger().Error(err)
	}
}
