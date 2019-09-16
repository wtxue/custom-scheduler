package config

import (
	"flag"
	"fmt"
	"k8s.io/klog"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

const (
	// LeaderLockName is the name of lock for leader election
	LeaderLockName = "custom-scheduler-lock"
)

// Options contains all the options for captain
type Options struct {
	BindAddressPort string

	// Max num Goroutine
	GoroutineThreshold int

	// PrintVersion print the version and exist
	PrintVersion bool

	// Options contains some useful options
	manager.Options
}

func (opt *Options) setDefaults() {
	opt.LeaderElectionID = LeaderLockName
}

// BindFlags init flags and options
func (opt *Options) BindFlags() {
	opt.setDefaults()

	flag.BoolVar(&opt.PrintVersion, "version", false, "Print version")
	flag.IntVar(&opt.GoroutineThreshold, "goroutine-threshold", 200, "check the max goroutine num")
	flag.BoolVar(&opt.LeaderElection, "enable-leader-election", false, "Enable leader election")
	flag.StringVar(&opt.BindAddressPort, "bind-address-port", ":8080", "Setup bind address for metrics and scheduler endpoint")
}

// FixKlogFlags copy flags between glog and klog
func FixKlogFlags() {
	// klog.SetOutput(os.Stdout)

	// This code sinppet copyed from klog. About glog/klog, we have a few options
	// 1. replace glog by klog in go mod --> pass kog.Infof to helm not working
	// 2. use this, it worked, but i don't known why. Fuck klog! Fuck glog!
	klogFlags := flag.NewFlagSet("klog", flag.ExitOnError)
	klog.InitFlags(klogFlags)

	// Sync the glog and klog flags.
	flag.CommandLine.VisitAll(func(f1 *flag.Flag) {
		f2 := klogFlags.Lookup(f1.Name)
		if f2 != nil {
			value := f1.Value.String()
			// default is invalid for parser...
			if f1.Name != "log_backtrace_at" || value != ":0" {
				if err := f2.Value.Set(value); err != nil {
					fmt.Printf("init klog flag %s:%s error: %s\n", f1.Name, value, err.Error())
				}
			}
		}
	})

	// why i need to set this????
	klog.SetOutput(os.Stdout)
}
