package web

import (
	"time"

	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
)

// LoggingMiddleware (C.8 Advanced Middleware & Logging): tutorial
// mengontraskan dua opsi - middleware.LoggerWithConfig bawaan Echo (cukup
// untuk format teks sederhana) vs middleware custom berbasis Logrus (kalau
// butuh log terstruktur/fields, dikirim ke sistem log agregator, dst).
// Dipilih Logrus di sini karena project 3 (learn-golang-c) sudah memakai
// pola logging manual ala stdlib (log.Printf) - projek ini sengaja
// menunjukkan alternatif "logger pihak ketiga sebagai middleware Echo".
func LoggingMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		start := time.Now()
		err := next(c)

		entry := log.WithFields(log.Fields{
			"method":   c.Request().Method,
			"uri":      c.Request().URL.String(),
			"status":   c.Response().Status,
			"duration": time.Since(start).String(),
		})
		if err != nil {
			entry.WithError(err).Warn("request selesai dengan error")
		} else {
			entry.Info("request selesai")
		}
		return err
	}
}
