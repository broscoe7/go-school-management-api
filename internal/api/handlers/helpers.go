package handlers

import (
	"errors"
	"reflect"

	"github.com/broscoe7/go-school-api/pkg/utils"
)

func CheckBlankFields(value any) error {
	val := reflect.ValueOf(value)
	for _, field := range val.Fields() {
		if field.Kind() == reflect.String && field.String() == "" {
			return utils.ErrorHandler(errors.New("all fields are required"), "All fields are required")
		}
	}
	return nil
}
