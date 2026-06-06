package admin

import (
	"net/http"

	"absensi-cn-api/pkg/response"

	"github.com/gin-gonic/gin"
)

func (h *Handler) ListTeacherProfiles(c *gin.Context) {
	result, err := h.service.ListTeacherProfiles()
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "teacher profiles fetched", result)
}

func (h *Handler) CreateTeacherProfile(c *gin.Context) {
	var request UpsertTeacherRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}
	if validationErrors := ValidateUpsertTeacherRequest(request); len(validationErrors) > 0 {
		response.ValidationError(c, http.StatusBadRequest, validationErrors)
		return
	}
	result, err := h.service.CreateTeacherProfile(request)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, http.StatusCreated, "teacher profile created", result)
}

func (h *Handler) UpdateTeacherProfile(c *gin.Context) {
	var request UpsertTeacherRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}
	if validationErrors := ValidateUpsertTeacherRequest(request); len(validationErrors) > 0 {
		response.ValidationError(c, http.StatusBadRequest, validationErrors)
		return
	}
	result, err := h.service.UpdateTeacherProfile(c.Param("id"), request)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "teacher profile updated", result)
}

func (h *Handler) DeleteTeacherProfile(c *gin.Context) {
	if err := h.service.DeleteTeacherProfile(c.Param("id")); err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "teacher profile deleted", nil)
}

func (h *Handler) ListSubjects(c *gin.Context) {
	result, err := h.service.ListSubjects()
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "subjects fetched", result)
}

func (h *Handler) CreateSubject(c *gin.Context) {
	var request UpsertSubjectRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}
	if validationErrors := ValidateUpsertSubjectRequest(request); len(validationErrors) > 0 {
		response.ValidationError(c, http.StatusBadRequest, validationErrors)
		return
	}
	result, err := h.service.CreateSubject(request)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, http.StatusCreated, "subject created", result)
}

func (h *Handler) UpdateSubject(c *gin.Context) {
	var request UpsertSubjectRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}
	if validationErrors := ValidateUpsertSubjectRequest(request); len(validationErrors) > 0 {
		response.ValidationError(c, http.StatusBadRequest, validationErrors)
		return
	}
	result, err := h.service.UpdateSubject(c.Param("id"), request)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "subject updated", result)
}

func (h *Handler) DeleteSubject(c *gin.Context) {
	if err := h.service.DeleteSubject(c.Param("id")); err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "subject deleted", nil)
}

func (h *Handler) ListMajors(c *gin.Context) {
	result, err := h.service.ListMajors()
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "majors fetched", result)
}

func (h *Handler) CreateMajor(c *gin.Context) {
	var request UpsertMajorRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}
	if validationErrors := ValidateUpsertMajorRequest(request); len(validationErrors) > 0 {
		response.ValidationError(c, http.StatusBadRequest, validationErrors)
		return
	}
	result, err := h.service.CreateMajor(request)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, http.StatusCreated, "major created", result)
}

func (h *Handler) UpdateMajor(c *gin.Context) {
	var request UpsertMajorRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}
	if validationErrors := ValidateUpsertMajorRequest(request); len(validationErrors) > 0 {
		response.ValidationError(c, http.StatusBadRequest, validationErrors)
		return
	}
	result, err := h.service.UpdateMajor(c.Param("id"), request)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "major updated", result)
}

func (h *Handler) DeleteMajor(c *gin.Context) {
	if err := h.service.DeleteMajor(c.Param("id")); err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "major deleted", nil)
}

func (h *Handler) ListSchoolYears(c *gin.Context) {
	result, err := h.service.ListSchoolYears()
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "school years fetched", result)
}

func (h *Handler) CreateSchoolYear(c *gin.Context) {
	var request UpsertSchoolYearRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}
	if validationErrors := ValidateUpsertSchoolYearRequest(request); len(validationErrors) > 0 {
		response.ValidationError(c, http.StatusBadRequest, validationErrors)
		return
	}
	result, err := h.service.CreateSchoolYear(request)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, http.StatusCreated, "school year created", result)
}

func (h *Handler) UpdateSchoolYear(c *gin.Context) {
	var request UpsertSchoolYearRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}
	if validationErrors := ValidateUpsertSchoolYearRequest(request); len(validationErrors) > 0 {
		response.ValidationError(c, http.StatusBadRequest, validationErrors)
		return
	}
	result, err := h.service.UpdateSchoolYear(c.Param("id"), request)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "school year updated", result)
}

func (h *Handler) DeleteSchoolYear(c *gin.Context) {
	if err := h.service.DeleteSchoolYear(c.Param("id")); err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "school year deleted", nil)
}

func (h *Handler) ListClasses(c *gin.Context) {
	result, err := h.service.ListClasses()
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "classes fetched", result)
}

func (h *Handler) CreateClass(c *gin.Context) {
	var request UpsertClassRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}
	if validationErrors := ValidateUpsertClassRequest(request); len(validationErrors) > 0 {
		response.ValidationError(c, http.StatusBadRequest, validationErrors)
		return
	}
	result, err := h.service.CreateClass(request)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, http.StatusCreated, "class created", result)
}

func (h *Handler) UpdateClass(c *gin.Context) {
	var request UpsertClassRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}
	if validationErrors := ValidateUpsertClassRequest(request); len(validationErrors) > 0 {
		response.ValidationError(c, http.StatusBadRequest, validationErrors)
		return
	}
	result, err := h.service.UpdateClass(c.Param("id"), request)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "class updated", result)
}

func (h *Handler) DeleteClass(c *gin.Context) {
	if err := h.service.DeleteClass(c.Param("id")); err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "class deleted", nil)
}

func (h *Handler) ListTeacherSubjectAssignments(c *gin.Context) {
	result, err := h.service.ListTeacherSubjectAssignments()
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "teacher subject assignments fetched", result)
}

func (h *Handler) CreateTeacherSubjectAssignment(c *gin.Context) {
	var request UpsertTeacherSubjectAssignmentRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}
	if validationErrors := ValidateUpsertTeacherSubjectAssignmentRequest(request); len(validationErrors) > 0 {
		response.ValidationError(c, http.StatusBadRequest, validationErrors)
		return
	}
	result, err := h.service.CreateTeacherSubjectAssignment(request)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, http.StatusCreated, "teacher subject assignment created", result)
}

func (h *Handler) UpdateTeacherSubjectAssignment(c *gin.Context) {
	var request UpsertTeacherSubjectAssignmentRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}
	if validationErrors := ValidateUpsertTeacherSubjectAssignmentRequest(request); len(validationErrors) > 0 {
		response.ValidationError(c, http.StatusBadRequest, validationErrors)
		return
	}
	result, err := h.service.UpdateTeacherSubjectAssignment(c.Param("id"), request)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "teacher subject assignment updated", result)
}

func (h *Handler) DeleteTeacherSubjectAssignment(c *gin.Context) {
	if err := h.service.DeleteTeacherSubjectAssignment(c.Param("id")); err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "teacher subject assignment deleted", nil)
}

func (h *Handler) ListHomeroomAssignments(c *gin.Context) {
	result, err := h.service.ListHomeroomAssignments()
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "homeroom assignments fetched", result)
}

func (h *Handler) CreateHomeroomAssignment(c *gin.Context) {
	var request UpsertHomeroomAssignmentRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}
	if validationErrors := ValidateUpsertHomeroomAssignmentRequest(request); len(validationErrors) > 0 {
		response.ValidationError(c, http.StatusBadRequest, validationErrors)
		return
	}
	result, err := h.service.CreateHomeroomAssignment(request)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, http.StatusCreated, "homeroom assignment created", result)
}

func (h *Handler) UpdateHomeroomAssignment(c *gin.Context) {
	var request UpsertHomeroomAssignmentRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}
	if validationErrors := ValidateUpsertHomeroomAssignmentRequest(request); len(validationErrors) > 0 {
		response.ValidationError(c, http.StatusBadRequest, validationErrors)
		return
	}
	result, err := h.service.UpdateHomeroomAssignment(c.Param("id"), request)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "homeroom assignment updated", result)
}

func (h *Handler) DeleteHomeroomAssignment(c *gin.Context) {
	if err := h.service.DeleteHomeroomAssignment(c.Param("id")); err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "homeroom assignment deleted", nil)
}
