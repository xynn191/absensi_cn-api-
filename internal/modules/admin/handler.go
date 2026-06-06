package admin

import (
	"errors"
	"net/http"

	attendanceModule "absensi-cn-api/internal/modules/attendance"
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
	result, err := h.service.GetDashboard()
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "admin dashboard fetched", result)
}

func (h *Handler) ListUsers(c *gin.Context) {
	result, err := h.service.ListUsers()
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "admin users fetched", result)
}

func (h *Handler) ListTeachers(c *gin.Context) {
	result, err := h.service.ListTeachers()
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "admin teachers fetched", result)
}

func (h *Handler) GetUser(c *gin.Context) {
	result, err := h.service.GetUser(c.Param("id"))
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "admin user fetched", result)
}

func (h *Handler) CreateUser(c *gin.Context) {
	var request UpsertUserRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}

	if validationErrors := ValidateUpsertUserRequest(request, true); len(validationErrors) > 0 {
		response.ValidationError(c, http.StatusBadRequest, validationErrors)
		return
	}

	result, err := h.service.CreateUser(request)
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, http.StatusCreated, "admin user created", result)
}

func (h *Handler) UpdateUser(c *gin.Context) {
	var request UpsertUserRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}

	if validationErrors := ValidateUpsertUserRequest(request, false); len(validationErrors) > 0 {
		response.ValidationError(c, http.StatusBadRequest, validationErrors)
		return
	}

	result, err := h.service.UpdateUser(c.Param("id"), request)
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "admin user updated", result)
}

func (h *Handler) DeleteUser(c *gin.Context) {
	actorID, _ := c.Get("auth_user_id")

	if err := h.service.DeleteUser(c.Param("id"), actorID.(string)); err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "admin user deleted", nil)
}

func handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrAdminDataUnavailable):
		response.Error(c, http.StatusServiceUnavailable, err.Error())
	case errors.Is(err, ErrUserNotFound),
		errors.Is(err, ErrTeacherProfileNotFound),
		errors.Is(err, ErrSubjectNotFound),
		errors.Is(err, ErrMajorNotFound),
		errors.Is(err, ErrSchoolYearNotFound),
		errors.Is(err, ErrClassNotFound),
		errors.Is(err, ErrStudentNotFound),
		errors.Is(err, ErrStudentMembershipNotFound),
		errors.Is(err, ErrAttendanceRuleNotFound),
		errors.Is(err, ErrTeacherAssignmentNotFound),
		errors.Is(err, ErrHomeroomAssignmentNotFound),
		errors.Is(err, attendanceModule.ErrAttendanceRecordNotFound):
		response.Error(c, http.StatusNotFound, err.Error())
	case errors.Is(err, ErrCannotDeleteSelf):
		response.Error(c, http.StatusForbidden, err.Error())
	case errors.Is(err, attendanceModule.ErrAttendanceReviewInvalid):
		response.Error(c, http.StatusBadRequest, err.Error())
	default:
		response.Error(c, http.StatusBadRequest, err.Error())
	}
}
