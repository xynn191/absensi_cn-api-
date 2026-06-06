package staff

import (
	"errors"
	"net/http"

	"absensi-cn-api/internal/modules/attendance"
	"absensi-cn-api/pkg/response"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) TeacherMe(c *gin.Context) {
	result, err := h.service.GetTeacherMe(c.GetString("auth_user_id"))
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "teacher profile fetched", result)
}

func (h *Handler) TeacherSubjectAssignments(c *gin.Context) {
	result, err := h.service.ListTeacherSubjectAssignments(c.GetString("auth_user_id"))
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "teacher subject assignments fetched", result)
}

func (h *Handler) TeacherHomeroom(c *gin.Context) {
	result, err := h.service.GetTeacherHomeroom(c.GetString("auth_user_id"))
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "teacher homeroom fetched", result)
}

func (h *Handler) TeacherHomeroomDashboard(c *gin.Context) {
	result, err := h.service.GetHomeroomDashboard(c.GetString("auth_user_id"))
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "homeroom dashboard fetched", result)
}

func (h *Handler) TeacherHomeroomStudents(c *gin.Context) {
	result, err := h.service.ListHomeroomStudents(c.GetString("auth_user_id"))
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "homeroom students fetched", result)
}

func (h *Handler) TeacherHomeroomStudentDetail(c *gin.Context) {
	result, err := h.service.GetHomeroomStudentDetail(c.GetString("auth_user_id"), c.Param("student_id"))
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "homeroom student detail fetched", result)
}

func (h *Handler) TeacherHomeroomAttendance(c *gin.Context) {
	result, err := h.service.ListHomeroomAttendance(c.GetString("auth_user_id"), c.Query("date"), c.Query("status"))
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "homeroom attendance fetched", result)
}

func (h *Handler) TeacherHomeroomAttendanceOverview(c *gin.Context) {
	result, err := h.service.GetHomeroomAttendanceOverview(
		c.GetString("auth_user_id"),
		c.Query("date"),
		c.Query("status"),
		c.Query("query"),
	)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "homeroom attendance overview fetched", result)
}

func (h *Handler) TeacherHomeroomAttendanceSummary(c *gin.Context) {
	result, err := h.service.GetHomeroomAttendanceSummary(c.GetString("auth_user_id"), c.Query("date"))
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "homeroom attendance summary fetched", result)
}

func (h *Handler) TeacherHomeroomAttendanceReview(c *gin.Context) {
	var request attendance.ReviewAttendanceRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.service.ReviewHomeroomAttendance(c.GetString("auth_user_id"), c.Param("id"), c.GetString("auth_user_id"), request)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "homeroom attendance reviewed", result)
}

func (h *Handler) TeacherStudentAttendanceHistory(c *gin.Context) {
	result, err := h.service.ListHomeroomStudentAttendanceHistory(c.GetString("auth_user_id"), c.Param("student_id"))
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "student attendance history fetched", result)
}

func (h *Handler) TeacherHomeroomSubmissions(c *gin.Context) {
	result, err := h.service.ListHomeroomSubmissions(c.GetString("auth_user_id"), c.Query("status"))
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "homeroom submissions fetched", result)
}

func (h *Handler) TeacherHomeroomSubmissionsOverview(c *gin.Context) {
	result, err := h.service.GetHomeroomSubmissionsOverview(
		c.GetString("auth_user_id"),
		c.Query("status"),
		c.Query("type"),
		c.Query("query"),
	)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "homeroom submissions overview fetched", result)
}

func (h *Handler) TeacherHomeroomSubmissionReview(c *gin.Context) {
	var request ReviewSubmissionRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.service.ReviewHomeroomSubmission(c.GetString("auth_user_id"), c.Param("id"), c.GetString("auth_user_id"), request)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "homeroom submission reviewed", result)
}

func (h *Handler) StudentSubmissions(c *gin.Context) {
	result, err := h.service.ListStudentSubmissions(c.GetString("auth_user_id"))
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "student submissions fetched", result)
}

func (h *Handler) StudentCreateSubmission(c *gin.Context) {
	var request CreateSubmissionRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.service.CreateStudentSubmission(c.GetString("auth_user_id"), request)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, http.StatusCreated, "student submission created", result)
}

func (h *Handler) BKStudents(c *gin.Context) {
	result, err := h.service.ListBKStudents(c.Query("class_id"))
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "bk students fetched", result)
}

func (h *Handler) BKDashboard(c *gin.Context) {
	result, err := h.service.GetBKDashboard()
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "bk dashboard fetched", result)
}

func (h *Handler) BKStudentsOverview(c *gin.Context) {
	result, err := h.service.GetBKStudentsOverview(
		c.Query("class_id"),
		c.Query("risk"),
		c.Query("query"),
	)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "bk students overview fetched", result)
}

func (h *Handler) BKStudent(c *gin.Context) {
	result, err := h.service.GetBKStudent(c.Param("student_id"))
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "bk student fetched", result)
}

func (h *Handler) BKAttendance(c *gin.Context) {
	result, err := h.service.ListBKAttendance(c.Query("date"), c.Query("status"), c.Query("class_id"))
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "bk attendance fetched", result)
}

func (h *Handler) BKAttendanceOverview(c *gin.Context) {
	result, err := h.service.GetBKAttendanceOverview(
		c.Query("date"),
		c.Query("status"),
		c.Query("class_id"),
		c.Query("query"),
	)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "bk attendance overview fetched", result)
}

func (h *Handler) BKAttendanceSummary(c *gin.Context) {
	result, err := h.service.GetBKAttendanceSummary(c.Query("date"))
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "bk attendance summary fetched", result)
}

func (h *Handler) BKAttendanceReview(c *gin.Context) {
	var request attendance.ReviewAttendanceRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.service.ReviewBKAttendance(c.Param("id"), c.GetString("auth_user_id"), request)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "bk attendance reviewed", result)
}

func (h *Handler) BKStudentAttendanceHistory(c *gin.Context) {
	result, err := h.service.ListBKStudentAttendanceHistory(c.Param("student_id"))
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "bk student attendance history fetched", result)
}

func (h *Handler) BKStudentCounselingNotes(c *gin.Context) {
	result, err := h.service.ListBKStudentCounselingNotes(c.Param("student_id"))
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "bk counseling notes fetched", result)
}

func (h *Handler) BKCounselingOverview(c *gin.Context) {
	result, err := h.service.GetBKCounselingOverview(
		c.Query("class_id"),
		c.Query("student_id"),
		c.Query("query"),
	)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "bk counseling overview fetched", result)
}

func (h *Handler) BKCreateCounselingNote(c *gin.Context) {
	var request UpsertCounselingNoteRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.service.CreateBKStudentCounselingNote(c.Param("student_id"), c.GetString("auth_user_id"), request)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, http.StatusCreated, "bk counseling note created", result)
}

func (h *Handler) BKUpdateCounselingNote(c *gin.Context) {
	var request UpsertCounselingNoteRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.service.UpdateBKCounselingNote(c.Param("id"), request)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "bk counseling note updated", result)
}

func (h *Handler) BKDeleteCounselingNote(c *gin.Context) {
	if err := h.service.DeleteBKCounselingNote(c.Param("id")); err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "bk counseling note deleted", nil)
}

func (h *Handler) BKSubmissions(c *gin.Context) {
	result, err := h.service.ListBKSubmissions(c.Query("status"), c.Query("class_id"))
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "bk submissions fetched", result)
}

func (h *Handler) BKSubmissionsOverview(c *gin.Context) {
	result, err := h.service.GetBKSubmissionsOverview(
		c.Query("status"),
		c.Query("type"),
		c.Query("class_id"),
		c.Query("query"),
	)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "bk submissions overview fetched", result)
}

func (h *Handler) BKSubmissionReview(c *gin.Context) {
	var request ReviewSubmissionRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.service.ReviewBKSubmission(c.Param("id"), c.GetString("auth_user_id"), request)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "bk submission reviewed", result)
}

func handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrStaffDataUnavailable):
		response.Error(c, http.StatusServiceUnavailable, err.Error())
	case errors.Is(err, ErrTeacherProfileNotFound),
		errors.Is(err, ErrHomeroomNotFound),
		errors.Is(err, ErrStudentNotFound),
		errors.Is(err, ErrSubmissionNotFound),
		errors.Is(err, ErrCounselingNoteNotFound),
		errors.Is(err, attendance.ErrAttendanceRecordNotFound):
		response.Error(c, http.StatusNotFound, err.Error())
	default:
		response.Error(c, http.StatusBadRequest, err.Error())
	}
}
