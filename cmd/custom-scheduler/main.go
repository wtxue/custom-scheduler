package main

import (
	"os"
	"flag"
	"k8s.io/klog"

	logf "sigs.k8s.io/controller-runtime/pkg/log"
	opt "github.com/xkcp0324/custom-scheduler/cmd/custom-scheduler/option"
	"github.com/xkcp0324/custom-scheduler/pkg/version"
	"github.com/xkcp0324/custom-scheduler/pkg/router"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
	"github.com/xkcp0324/custom-scheduler/pkg/scheduler"
	"github.com/xkcp0324/custom-scheduler/pkg/healthcheck"
	"k8s.io/klog/klogr"
)


func main() {
	var options opt.Options
	options.BindFlags()
	klog.InitFlags(nil)
	flag.Parse()

	// logf.SetLogger(zap.Logger(false))
	logf.SetLogger(klogr.New())
	loggger := logf.Log.WithName("entrypoint")

	// print version and exist
	if options.PrintVersion {
		klog.Infof("Welcome to custom Scheduler.")
		klog.Infof("%#v\n", version.GetVersion())
		return
	}

	cfg, err := config.GetConfig()
	if err != nil {
		loggger.Error(err, "unable to set up client config")
		os.Exit(1)
	}

	loggger.Info("setting up manager")
	managerOptions := manager.Options{
		MetricsBindAddress: "0",
	}

	if options.LeaderElection {
		managerOptions.LeaderElection = true
		managerOptions.LeaderElectionID = options.LeaderElectionID
	}

	mgr, err := manager.New(cfg, managerOptions)
	if err != nil {
		loggger.Error(err, "unable to set up controller manager")
		os.Exit(1)
	}

	kubeCli, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("failed to get kubernetes Clientset: %v", err)
	}

	routerOptions := &router.RouterOptions{
		IsGinLogEnabled:  true,
		IsMetricsEnabled: true,
		IsPprofEnabled:   true,
		Addr:             options.BindAddressPort,
		MetricsPath:      "metrics",
		MetricsSubsystem: "custom_scheduler",
	}

	healthHander := healthcheck.NewHealthHandler()
	healthHander.AddLivenessCheck("goroutine_threshold",
		healthcheck.GoroutineCountCheck(options.GoroutineThreshold))

	rt := router.NewRouter(routerOptions)
	rt.AddRoutes("rt", router.DefaultRoutes())
	rt.AddRoutes("scheduler", scheduler.NewServer(kubeCli, mgr).Routes())
	rt.AddRoutes("health", healthHander.Routes())

	loggger.Info("adding gin http server")
	err = mgr.Add(rt)
	if err != nil {
		loggger.Error(err, "Unable to add gin runnableServer")
		os.Exit(1)
	}

	loggger.Info("Starting the Cmd.")
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		loggger.Error(err, "unable to run the manager")
		os.Exit(1)
	}
}
