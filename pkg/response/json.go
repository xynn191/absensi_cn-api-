package response

import "github.com/gin-gonic/gin"

type envelope struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Errors  interface{} `json:"errors,omitempty"`
}

func Success(c *gin.Context, statusCode int, message string, data interface{}) {
	c.JSON(statusCode, envelope{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func Error(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, envelope{
		Success: false,
		Message: message,
	})
}

func ValidationError(c *gin.Context, statusCode int, errors map[string]string) {
	c.JSON(statusCode, envelope{
		Success: false,
		Message: "validation failed",
		Errors:  errors,
	})
}
