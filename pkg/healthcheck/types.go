package healthcheck

import (
	"github.com/gin-gonic/gin"
	"github.com/xkcp0324/custom-scheduler/pkg/router"
)

// Check is a health/readiness check.
type Check func() error

// Handler is an endpoints with additional methods that register health and
// readiness checks. It handles handle "/live" and "/ready" HTTP
// endpoints.
type Handler interface {
	Routes() []*router.Route
	AddLivenessCheck(name string, check Check)
	AddReadinessCheck(name string, check Check)
	LiveEndpoint(ctx *gin.Context)
	ReadyEndpoint(ctx *gin.Context)
}
