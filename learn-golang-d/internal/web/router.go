package web

import "github.com/labstack/echo/v4"

// RegisterRoutes (C.1 project layout + C.3 Echo routing): wiring rute
// dikumpulkan di satu tempat terpisah dari main.go, supaya main.go cuma
// urusan bootstrap (baca config, bikin dependency, jalankan server) -
// sama pembagian tanggung jawab seperti cmd/notetracker/main.go di
// learn-golang-c.
func RegisterRoutes(e *echo.Echo, notesH *NotesHandler, habitsH *HabitsHandler, authH *AuthHandler, signingKey []byte) {
	jwtAuth := JWTMiddleware(signingKey)

	e.POST("/login", authH.Login)

	notes := e.Group("/notes")
	notes.GET("", notesH.List)
	notes.GET("/:id", notesH.Detail)
	notes.POST("", notesH.Create, jwtAuth)
	notes.DELETE("/:id", notesH.Delete, jwtAuth)

	habits := e.Group("/habits")
	habits.GET("", habitsH.List)
	habits.POST("", habitsH.Create, jwtAuth)
	habits.DELETE("/:id", habitsH.Delete, jwtAuth)
	habits.POST("/:id/checkin", habitsH.Checkin, jwtAuth)
}
