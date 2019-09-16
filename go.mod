module github.com/xkcp0324/custom-scheduler

go 1.13

require (
	github.com/DeanThompson/ginpprof v0.0.0-20190408063150-3be636683586
	github.com/gin-gonic/gin v1.4.0
	github.com/gofrs/uuid v3.2.0+incompatible
	github.com/prometheus/client_golang v1.1.0
	k8s.io/api v0.0.0-20190409021203-6e4e0e4f393b
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	k8s.io/klog v0.4.0
	k8s.io/kubernetes v1.14.6
	sigs.k8s.io/controller-runtime v0.2.1
)

replace (
	k8s.io/kubernetes => k8s.io/kubernetes v1.14.6
	sigs.k8s.io/controller-runtime => sigs.k8s.io/controller-runtime v0.2.1
)
