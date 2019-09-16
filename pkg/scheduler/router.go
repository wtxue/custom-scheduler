package scheduler

import (
	"net/http"
	"sync"
	"k8s.io/klog"
	"github.com/gin-gonic/gin"
	"k8s.io/client-go/kubernetes"
	schedulerapiv1 "k8s.io/kubernetes/pkg/scheduler/api/v1"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"github.com/xkcp0324/custom-scheduler/pkg/router"
)

// ErrorResponse describes responses when an error occurred
type ErrorResponse struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

type Server struct {
	scheduler Scheduler
	lock      sync.Mutex
}

// StartServer starts a kubernetes scheduler extender http apiserver
func NewServer(kubeCli kubernetes.Interface, mgr manager.Manager) *Server {
	s := NewScheduler(kubeCli, mgr)
	return &Server{scheduler: s}
}

func (svr *Server) filterNode(ctx *gin.Context) {
	svr.lock.Lock()
	defer svr.lock.Unlock()

	args := &schedulerapiv1.ExtenderArgs{}
	if err := ctx.BindJSON(args); err != nil {
		klog.Errorf("filterNode unable to read request body")
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "unable to read request body",
			Error:   err.Error(),
		})
		return
	}

	klog.Infof("filterNode args:%#v", args)
	filterResult, err := svr.scheduler.Filter(args)
	if err != nil {
		klog.Errorf("unable to filter nodes")
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "unable to filter nodes",
			Error:   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, filterResult)
}

func (svr *Server) prioritizeNode(ctx *gin.Context) {
	svr.lock.Lock()
	defer svr.lock.Unlock()

	args := &schedulerapiv1.ExtenderArgs{}
	if err := ctx.BindJSON(args); err != nil {
		klog.Errorf("prioritizeNode unable to read request body")
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "unable to read request body",
			Error:   err.Error(),
		})
		return
	}

	klog.Infof("prioritizeNode args:%+v", args)
	priorityResult, err := svr.scheduler.Priority(args)
	if err != nil {
		klog.Errorf("unable to priority nodes")
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "unable to priority nodes",
			Error:   err.Error(),
		})

		return
	}

	ctx.JSON(http.StatusOK, priorityResult)
}

func (svr *Server) Routes() []*router.Route {
	schedulerRoute := []*router.Route{
		{"POST", "/scheduler/filter", svr.filterNode, ""},
		{"POST", "/scheduler/prioritize", svr.prioritizeNode, ""},
	}

	return schedulerRoute
}
