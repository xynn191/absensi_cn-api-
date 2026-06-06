package admin

import (
	"net/http"

	attendanceModule "absensi-cn-api/internal/modules/attendance"
	"absensi-cn-api/pkg/response"

	"github.com/gin-gonic/gin"
)

func (h *Handler) ListStudents(c *gin.Context) {
	result, err := h.service.ListStudents()
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "students fetched", result)
}

func (h *Handler) GetStudent(c *gin.Context) {
	result, err := h.service.GetStudent(c.Param("id"))
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "student fetched", result)
}

func (h *Handler) CreateStudent(c *gin.Context) {
	var request UpsertStudentRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}
	if validationErrors := ValidateUpsertStudentRequest(request, true); len(validationErrors) > 0 {
		response.ValidationError(c, http.StatusBadRequest, validationErrors)
		return
	}

	result, err := h.service.CreateStudent(request)
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, http.StatusCreated, "student created", result)
}

func (h *Handler) UpdateStudent(c *gin.Context) {
	var request UpsertStudentRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}
	if validationErrors := ValidateUpsertStudentRequest(request, false); len(validationErrors) > 0 {
		response.ValidationError(c, http.StatusBadRequest, validationErrors)
		return
	}

	result, err := h.service.UpdateStudent(c.Param("id"), request)
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "student updated", result)
}

func (h *Handler) DeleteStudent(c *gin.Context) {
	if err := h.service.DeleteStudent(c.Param("id")); err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "student deleted", nil)
}

func (h *Handler) ListStudentClassMemberships(c *gin.Context) {
	result, err := h.service.ListStudentClassMemberships()
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "student class memberships fetched", result)
}

func (h *Handler) GetStudentClassMembership(c *gin.Context) {
	result, err := h.service.GetStudentClassMembership(c.Param("id"))
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "student class membership fetched", result)
}

func (h *Handler) CreateStudentClassMembership(c *gin.Context) {
	var request UpsertStudentClassMembershipRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}
	if validationErrors := ValidateUpsertStudentClassMembershipRequest(request); len(validationErrors) > 0 {
		response.ValidationError(c, http.StatusBadRequest, validationErrors)
		return
	}

	result, err := h.service.CreateStudentClassMembership(request)
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, http.StatusCreated, "student class membership created", result)
}

func (h *Handler) UpdateStudentClassMembership(c *gin.Context) {
	var request UpsertStudentClassMembershipRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}
	if validationErrors := ValidateUpsertStudentClassMembershipRequest(request); len(validationErrors) > 0 {
		response.ValidationError(c, http.StatusBadRequest, validationErrors)
		return
	}

	result, err := h.service.UpdateStudentClassMembership(c.Param("id"), request)
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "student class membership updated", result)
}

func (h *Handler) DeleteStudentClassMembership(c *gin.Context) {
	if err := h.service.DeleteStudentClassMembership(c.Param("id")); err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "student class membership deleted", nil)
}

func (h *Handler) ListAttendanceRules(c *gin.Context) {
	result, err := h.service.ListAttendanceRules()
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "attendance rules fetched", result)
}

func (h *Handler) GetAttendanceRule(c *gin.Context) {
	result, err := h.service.GetAttendanceRule(c.Param("id"))
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "attendance rule fetched", result)
}

func (h *Handler) CreateAttendanceRule(c *gin.Context) {
	var request UpsertAttendanceRuleRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}
	if validationErrors := ValidateUpsertAttendanceRuleRequest(request); len(validationErrors) > 0 {
		response.ValidationError(c, http.StatusBadRequest, validationErrors)
		return
	}

	result, err := h.service.CreateAttendanceRule(request)
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, http.StatusCreated, "attendance rule created", result)
}

func (h *Handler) UpdateAttendanceRule(c *gin.Context) {
	var request UpsertAttendanceRuleRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}
	if validationErrors := ValidateUpsertAttendanceRuleRequest(request); len(validationErrors) > 0 {
		response.ValidationError(c, http.StatusBadRequest, validationErrors)
		return
	}

	result, err := h.service.UpdateAttendanceRule(c.Param("id"), request)
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "attendance rule updated", result)
}

func (h *Handler) DeleteAttendanceRule(c *gin.Context) {
	if err := h.service.DeleteAttendanceRule(c.Param("id")); err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "attendance rule deleted", nil)
}

func (h *Handler) ListAttendanceRecords(c *gin.Context) {
	result, err := h.service.ListAdminAttendance(c.Query("date"), c.Query("status"), c.Query("class_id"))
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "attendance records fetched", result)
}

func (h *Handler) ReviewAttendance(c *gin.Context) {
	actorID, _ := c.Get("auth_user_id")

	var request attendanceModule.ReviewAttendanceRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.service.ReviewAttendance(c.Param("id"), actorID.(string), request)
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "attendance reviewed", result)
}
