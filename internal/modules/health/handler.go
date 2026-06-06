package health

import (
	"net/http"

	"absensi-cn-api/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct {
	db *gorm.DB
}

func NewHandler(db *gorm.DB) *Handler {
	return &Handler{db: db}
}

func (h *Handler) Check(c *gin.Context) {
	response.Success(c, http.StatusOK, "api is healthy", gin.H{
		"database_enabled": h.db != nil,
	})
}
