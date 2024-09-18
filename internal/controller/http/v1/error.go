package v1

import (
	"github.com/gin-gonic/gin"
)

type response struct {
	Error  string `json:"error" example:"message"`
	Status bool   `json:"status"`
}

func errorResponse(c *gin.Context, code int, msg string, status bool) {
	c.AbortWithStatusJSON(code, response{Error: msg, Status: status})
}
