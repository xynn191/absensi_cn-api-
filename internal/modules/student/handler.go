package student

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

func (h *Handler) Dashboard(c *gin.Context) {
	result, err := h.service.GetDashboard(c.GetString("auth_user_id"))
	if err != nil {
		handleStudentError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "student dashboard fetched", result)
}

func (h *Handler) Profile(c *gin.Context) {
	result, err := h.service.GetProfile(c.GetString("auth_user_id"))
	if err != nil {
		handleStudentError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "student profile fetched", result)
}

func (h *Handler) Today(c *gin.Context) {
	result, err := h.service.GetToday(c.GetString("auth_user_id"))
	if err != nil {
		handleStudentError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "student today attendance fetched", result)
}

func (h *Handler) History(c *gin.Context) {
	result, err := h.service.GetHistory(c.GetString("auth_user_id"))
	if err != nil {
		handleStudentError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "student history fetched", result)
}

func (h *Handler) SubmitDailyReport(c *gin.Context) {
	photo, err := c.FormFile("photo")
	if err != nil {
		handleStudentError(c, ErrPhotoRequired)
		return
	}

	result, err := h.service.SubmitDailyReport(
		c.GetString("auth_user_id"),
		c.PostForm("type"),
		c.PostForm("reason"),
		photo,
	)
	if err != nil {
		handleStudentError(c, err)
		return
	}

	response.Success(c, http.StatusCreated, "student attendance report submitted", result)
}

func (h *Handler) AttendanceHistory(c *gin.Context) {
	result, err := h.service.ListAttendanceHistory(c.GetString("auth_user_id"))
	if err != nil {
		handleStudentError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "student attendance history fetched", result)
}

func (h *Handler) Submissions(c *gin.Context) {
	result, err := h.service.ListSubmissions(c.GetString("auth_user_id"))
	if err != nil {
		handleStudentError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "student submissions fetched", result)
}

func handleStudentError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrStudentDataUnavailable):
		response.Error(c, http.StatusServiceUnavailable, err.Error())
	case errors.Is(err, ErrStudentNotFound):
		response.Error(c, http.StatusNotFound, err.Error())
	case errors.Is(err, ErrAlreadySubmittedToday),
		errors.Is(err, ErrAttendanceNotOpen),
		errors.Is(err, ErrPhotoRequired),
		errors.Is(err, ErrPhotoTooLarge),
		errors.Is(err, ErrPhotoInvalid),
		errors.Is(err, ErrReportTypeInvalid),
		errors.Is(err, ErrReasonRequired):
		response.Error(c, http.StatusBadRequest, err.Error())
	default:
		response.Error(c, http.StatusBadRequest, err.Error())
	}
}
