package rounter

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Options(c *gin.Context) { // options
	c.Writer.WriteHeader(http.StatusNoContent)
}
