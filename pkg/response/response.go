package response

import "github.com/gin-gonic/gin"

type Status string

const (
	StatusSuccess Status = "success"
	StatusFail    Status = "fail"
	StatusError   Status = "error"
)

type Response struct {
	Status  Status `json:"status"`
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

// Success sends a 2xx response with data.
func Success(c *gin.Context, httpCode int, message string, data any) {
	c.JSON(httpCode, Response{
		Status:  StatusSuccess,
		Code:    httpCode,
		Message: message,
		Data:    data,
	})
}

// Fail sends a 4xx response for client-side problems (e.g. validation errors).
// data can carry field-level error details.
func Fail(c *gin.Context, httpCode int, message string, data any) {
	c.JSON(httpCode, Response{
		Status:  StatusFail,
		Code:    httpCode,
		Message: message,
		Data:    data,
	})
}

// Error sends a 5xx response for server-side errors. data is always null.
func Error(c *gin.Context, httpCode int, message string) {
	c.JSON(httpCode, Response{
		Status:  StatusError,
		Code:    httpCode,
		Message: message,
		Data:    nil,
	})
}
