package utils

import (
	"errors"
	"slices"
)

type ContextKey string

func Authorize(userRole string, allowedRoles ...string) (bool, error) {
	if slices.Contains(allowedRoles, userRole) {
		return true, nil
	}
	return false, errors.New("user not authorized")
}
