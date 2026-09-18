package web

import "github.com/go-playground/validator/v10"

// CustomValidator (C.5 HTTP Request Payload Validation): tutorial memakai
// package "gopkg.in/go-playground/validator.v9", di sini dipakai v10
// (penerus resmi package yang sama, v9 sudah tidak dikembangkan) - pola
// integrasinya ke Echo identik: implement interface echo.Validator lewat
// satu method Validate(interface{}) error, didaftarkan sekali ke e.Validator
// di main.go.
type CustomValidator struct {
	validator *validator.Validate
}

func NewCustomValidator() *CustomValidator {
	return &CustomValidator{validator: validator.New()}
}

func (cv *CustomValidator) Validate(i any) error {
	return cv.validator.Struct(i)
}
