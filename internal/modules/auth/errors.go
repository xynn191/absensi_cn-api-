package auth

import "errors"

var (
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrAuthDataUnavailable = errors.New("auth database is not available")
	ErrStudentPortalOnly   = errors.New("Akun siswa tidak dapat login melalui portal staff")
	ErrStaffPortalOnly     = errors.New("Akun staff tidak dapat login melalui portal siswa")
)
