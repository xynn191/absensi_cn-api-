package auth

import (
	"errors"
	"net/http"

	"absensi-cn-api/pkg/response"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) Login(c *gin.Context) {
	var request LoginRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}

	if validationErrors := ValidateLoginRequest(request); len(validationErrors) > 0 {
		response.ValidationError(c, http.StatusBadRequest, validationErrors)
		return
	}

	result, err := h.service.Login(request)
	if err != nil {
		switch {
		case errors.Is(err, ErrAuthDataUnavailable):
			response.Error(c, http.StatusServiceUnavailable, err.Error())
		case errors.Is(err, ErrInvalidCredentials):
			response.Error(c, http.StatusUnauthorized, err.Error())
		case errors.Is(err, ErrStudentPortalOnly), errors.Is(err, ErrStaffPortalOnly):
			response.Error(c, http.StatusForbidden, err.Error())
		default:
			response.Error(c, http.StatusUnauthorized, err.Error())
		}

		return
	}

	response.Success(c, http.StatusOK, "login success", result)
}
