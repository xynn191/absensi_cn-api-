package public

import (
	"net/http"

	"absensi-cn-api/pkg/response"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) AttendanceWindow(c *gin.Context) {
	result, err := h.service.AttendanceWindow()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "attendance window fetched", result)
}
