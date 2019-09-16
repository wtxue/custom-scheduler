package router

import (
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"strconv"
	"sync/atomic"
)

var (
	requestCount         int64
	requestServedMessage = "Request served"
)

func setupContext(c *gin.Context) {
	reqCount := strconv.FormatInt(atomic.AddInt64(&requestCount, 1), 10)
	c.Set("requestcount", reqCount)
	reqID := c.Request.Header.Get("X-Request-Id")
	if reqID == "" {
		reqID = uuid.Must(uuid.NewV4()).String()
	}
	c.Set("requestid", reqID)
	c.Writer.Header().Set("X-Request-Id", reqID)
}
