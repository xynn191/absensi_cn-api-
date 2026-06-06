package attendance

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
	return &Handler{service: service}
}

func (h *Handler) Me(c *gin.Context) {
	userID := c.GetString("auth_user_id")

	result, err := h.service.GetStudentContextByUserID(userID)
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "student context fetched", result)
}

func (h *Handler) Today(c *gin.Context) {
	userID := c.GetString("auth_user_id")

	result, err := h.service.GetTodayAttendance(userID)
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "today attendance fetched", result)
}

func (h *Handler) History(c *gin.Context) {
	userID := c.GetString("auth_user_id")

	result, err := h.service.ListStudentAttendanceHistory(userID)
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "attendance history fetched", result)
}

func (h *Handler) CheckIn(c *gin.Context) {
	userID := c.GetString("auth_user_id")

	photo, err := c.FormFile("photo")
	if err != nil {
		handleError(c, ErrAttendancePhotoRequired)
		return
	}

	result, err := h.service.CheckIn(userID, photo, c.PostForm("notes"))
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, http.StatusCreated, "check-in recorded", result)
}

func (h *Handler) CheckOut(c *gin.Context) {
	response.Error(c, http.StatusMethodNotAllowed, "check-out is not used in this attendance flow")
}

func handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrAttendanceDataUnavailable):
		response.Error(c, http.StatusServiceUnavailable, err.Error())
	case errors.Is(err, ErrStudentContextNotFound),
		errors.Is(err, ErrAttendanceRecordNotFound):
		response.Error(c, http.StatusNotFound, err.Error())
	case errors.Is(err, ErrAttendanceNotOpen),
		errors.Is(err, ErrAttendanceAlreadyRecorded),
		errors.Is(err, ErrAttendancePhotoRequired),
		errors.Is(err, ErrAttendancePhotoTooLarge),
		errors.Is(err, ErrAttendancePhotoInvalid),
		errors.Is(err, ErrAttendanceReviewInvalid):
		response.Error(c, http.StatusBadRequest, err.Error())
	default:
		response.Error(c, http.StatusBadRequest, err.Error())
	}
}
