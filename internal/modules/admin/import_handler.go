package admin

import (
	"fmt"
	"io"
	"net/http"

	"absensi-cn-api/pkg/response"

	"github.com/gin-gonic/gin"
)

func (h *Handler) DownloadTeacherTemplate(c *gin.Context) {
	data, err := h.service.GenerateTeacherTemplate()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "gagal membuat template: "+err.Error())
		return
	}

	c.Header("Content-Disposition", `attachment; filename="template_import_guru.xlsx"`)
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Length", fmt.Sprintf("%d", len(data)))
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", data)
}

func (h *Handler) DownloadStudentTemplate(c *gin.Context) {
	data, err := h.service.GenerateStudentTemplate()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "gagal membuat template: "+err.Error())
		return
	}

	c.Header("Content-Disposition", `attachment; filename="template_import_siswa.xlsx"`)
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Length", fmt.Sprintf("%d", len(data)))
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", data)
}

func (h *Handler) ImportTeachers(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "file tidak ditemukan dalam request")
		return
	}

	if fileHeader.Size > 5*1024*1024 {
		response.Error(c, http.StatusBadRequest, "ukuran file maksimal 5MB")
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "gagal membuka file")
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "gagal membaca file")
		return
	}

	result, err := h.service.ImportTeachers(data)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "import guru selesai", result)
}

func (h *Handler) ImportStudents(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "file tidak ditemukan dalam request")
		return
	}

	if fileHeader.Size > 5*1024*1024 {
		response.Error(c, http.StatusBadRequest, "ukuran file maksimal 5MB")
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "gagal membuka file")
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "gagal membaca file")
		return
	}

	result, err := h.service.ImportStudents(data)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "import siswa selesai", result)
}
