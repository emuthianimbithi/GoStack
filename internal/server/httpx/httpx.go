package httpx

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ErrorCode is a typed string for error codes.
type ErrorCode string

const (
	CodeBadRequest          ErrorCode = "BAD_REQUEST"
	CodeUnauthorized        ErrorCode = "UNAUTHORIZED"
	CodeForbidden           ErrorCode = "FORBIDDEN"
	CodeNotFound            ErrorCode = "NOT_FOUND"
	CodeInternalError       ErrorCode = "INTERNAL_ERROR"
	CodeMethodNotAllowed    ErrorCode = "METHOD_NOT_ALLOWED"
	CodeConflict            ErrorCode = "CONFLICT"
	CodeUnprocessableEntity ErrorCode = "UNPROCESSABLE_ENTITY"
	CodeTooManyRequests     ErrorCode = "TOO_MANY_REQUESTS"
	CodeServiceUnavailable  ErrorCode = "SERVICE_UNAVAILABLE"
	CodeNotImplemented      ErrorCode = "NOT_IMPLEMENTED"
)

// APIError describes the error payload.
type APIError struct {
	Code    ErrorCode   `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// APIResponse is the standard envelope for all responses.
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *APIError   `json:"error,omitempty"`
}

// JSON is the generic success response helper.
func JSON(c *gin.Context, status int, data interface{}) {
	c.JSON(status, APIResponse{
		Success: true,
		Data:    data,
	})
}

// Ok is a convenience wrapper for HTTP 200 with data.
func Ok(c *gin.Context, data interface{}) {
	JSON(c, http.StatusOK, data)
}

// NoContent sends HTTP 204.
func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// JSONError is the generic error response helper.
func JSONError(c *gin.Context, status int, code ErrorCode, message string, details interface{}) {
	c.AbortWithStatusJSON(status, APIResponse{
		Success: false,
		Error: &APIError{
			Code:    code,
			Message: message,
			Details: details,
		},
	})
}

// Convenience error helpers:

func BadRequest(c *gin.Context, message string, details interface{}) {
	JSONError(c, http.StatusBadRequest, CodeBadRequest, message, details)
}

func Unauthorized(c *gin.Context, message string) {
	JSONError(c, http.StatusUnauthorized, CodeUnauthorized, message, nil)
}

func Forbidden(c *gin.Context, message string) {
	JSONError(c, http.StatusForbidden, CodeForbidden, message, nil)
}

func NotFound(c *gin.Context, message string) {
	JSONError(c, http.StatusNotFound, CodeNotFound, message, nil)
}

func InternalError(c *gin.Context, message string) {
	JSONError(c, http.StatusInternalServerError, CodeInternalError, message, nil)
}

func MethodNotAllowed(c *gin.Context, message string) {
	JSONError(c, http.StatusMethodNotAllowed, CodeMethodNotAllowed, message, nil)
}

func Conflict(c *gin.Context, message string, details interface{}) {
	JSONError(c, http.StatusConflict, CodeConflict, message, details)
}

func UnprocessableEntity(c *gin.Context, message string, details interface{}) {
	JSONError(c, http.StatusUnprocessableEntity, CodeUnprocessableEntity, message, details)
}

func TooManyRequests(c *gin.Context, message string) {
	JSONError(c, http.StatusTooManyRequests, CodeTooManyRequests, message, nil)
}

func ServiceUnavailable(c *gin.Context, message string) {
	JSONError(c, http.StatusServiceUnavailable, CodeServiceUnavailable, message, nil)
}

func NotImplemented(c *gin.Context, message string) {
	JSONError(c, http.StatusNotImplemented, CodeNotImplemented, message, nil)
}

// ---------- Pagination ----------

type PageMeta struct {
	Page     int   `json:"page"`
	PageSize int   `json:"pageSize"`
	Total    int64 `json:"total"`
}

type PagedResponse struct {
	Items interface{} `json:"items"`
	PageMeta
}

func GetPageParams(c *gin.Context) (page int, pageSize int) {
	page = 1
	pageSize = 20

	if v := c.Query("page"); v != "" {
		if p, err := strconv.Atoi(v); err == nil && p > 0 {
			page = p
		}
	}

	if v := c.Query("pageSize"); v != "" {
		if s, err := strconv.Atoi(v); err == nil {
			if s < 1 {
				s = 1
			}
			if s > 200 {
				s = 200
			}
			pageSize = s
		}
	}
	return
}

func Paginate(c *gin.Context, base *gorm.DB, dest interface{}) (PageMeta, error) {
	page, pageSize := GetPageParams(c)

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return PageMeta{}, err
	}

	offset := (page - 1) * pageSize
	if err := base.Limit(pageSize).Offset(offset).Find(dest).Error; err != nil {
		return PageMeta{}, err
	}

	return PageMeta{
		Page:     page,
		PageSize: pageSize,
		Total:    total,
	}, nil
}

func WrapPaged(items interface{}, meta PageMeta) PagedResponse {
	return PagedResponse{
		Items: items,
		PageMeta: PageMeta{
			Page:     meta.Page,
			PageSize: meta.PageSize,
			Total:    meta.Total,
		},
	}
}
