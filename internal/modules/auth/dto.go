package auth

import "absensi-cn-api/internal/modules/user"

type LoginRequest struct {
	Portal   string `json:"portal"`
	NIS      string `json:"nis"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	AccessToken string      `json:"access_token"`
	User        UserPayload `json:"user"`
}

type UserPayload struct {
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	Role     user.Role `json:"role"`
	Portal   string    `json:"portal"`
	NIS      *string   `json:"nis,omitempty"`
	Username *string   `json:"username,omitempty"`
}
