package auth

import (
	"regexp"
	"strings"
)

var nisRegex = regexp.MustCompile(`^\d{10}$`)

func ValidateLoginRequest(input LoginRequest) map[string]string {
	errors := map[string]string{}

	switch input.Portal {
	case "student":
		if strings.TrimSpace(input.NIS) == "" {
			errors["nis"] = "NIS is required"
		} else if !nisRegex.MatchString(strings.TrimSpace(input.NIS)) {
			errors["nis"] = "NIS must contain exactly 10 numeric digits"
		}

		if strings.TrimSpace(input.Username) != "" {
			errors["username"] = "Username is not allowed for student portal login"
		}
	case "staff":
		if strings.TrimSpace(input.Username) == "" {
			errors["username"] = "Username is required"
		}

		if strings.TrimSpace(input.NIS) != "" {
			errors["nis"] = "NIS is not allowed for staff portal login"
		}
	default:
		errors["portal"] = "Portal must be either 'student' or 'staff'"
	}

	if strings.TrimSpace(input.Password) == "" {
		errors["password"] = "Password is required"
	} else if len(strings.TrimSpace(input.Password)) < 6 {
		errors["password"] = "Password must be at least 6 characters"
	}

	if len(errors) == 0 {
		return nil
	}

	return errors
}
