package scheduler

import (
	"github.com/xkcp0324/custom-scheduler/pkg/scheduler/predicates"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"

	"k8s.io/klog"
	schedulerapiv1 "k8s.io/kubernetes/pkg/scheduler/api/v1"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"github.com/xkcp0324/custom-scheduler/pkg/observe"
)

// Scheduler is an interface for external processes to influence scheduling
// decisions made by kubernetes. This is typically needed for resources not directly
// managed by kubernetes.
type Scheduler interface {
	// Filter based on extender-implemented predicate functions. The filtered list is
	// expected to be a subset of the supplied list.
	Filter(*schedulerapiv1.ExtenderArgs) (*schedulerapiv1.ExtenderFilterResult, error)

	// Prioritize based on extender-implemented priority functions. The returned scores & weight
	// are used to compute the weighted score for an extender. The weighted scores are added to
	// the scores computed  by kubernetes scheduler. The total scores are used to do the host selection.
	Priority(*schedulerapiv1.ExtenderArgs) (schedulerapiv1.HostPriorityList, error)
}

type scheduler struct {
	// component => predicates
	predicates map[string][]predicates.Predicate
}

// NewScheduler returns a Scheduler
func NewScheduler(kubeCli kubernetes.Interface, mgr manager.Manager) Scheduler {
	cacher := mgr.GetCache()
	_, err := cacher.GetInformerForKind(corev1.SchemeGroupVersion.WithKind("Pod"))
	if err != nil {
		klog.Errorf("cacher get informer err:%+v", err)
		return nil
	}

	// priorityClass, err := cacher.GetInformerForKind(corev1.SchemeGroupVersion.WithKind("PriorityClass"))
	// if err != nil {
	// 	klog.Errorf("cacher get informer err:%+v", err)
	// 	return nil
	// }
	//
	// kubeCli.SchedulingV1beta1().PriorityClasses().List()

	// podLister := corelisters.NewPodLister(podInformer.GetIndexer())
	predicatesByComponent := map[string][]predicates.Predicate{
		"ha": {
			predicates.NewHA(kubeCli, mgr),
		},
	}

	return &scheduler{
		predicates: predicatesByComponent,
	}
}

// Filter selects a set of nodes from *schedulerapiv1.ExtenderArgs.Nodes when this is a pd or tikv pod
// otherwise, returns the original nodes.
func (s *scheduler) Filter(args *schedulerapiv1.ExtenderArgs) (*schedulerapiv1.ExtenderFilterResult, error) {
	pod := args.Pod
	ns := pod.GetNamespace()
	podName := pod.GetName()
	kubeNodes := args.Nodes.Items

	var instanceName string
	var exist bool
	if instanceName, exist = pod.Labels[observe.ObserveMustLabelAppName]; !exist {
		klog.Warningf("can't find instanceName in pod labels: %s/%s", ns, podName)
		return &schedulerapiv1.ExtenderFilterResult{
			Nodes: args.Nodes,
		}, nil
	}

	predicatesByComponent, ok := s.predicates["ha"]
	if !ok {
		return &schedulerapiv1.ExtenderFilterResult{
			Nodes: args.Nodes,
		}, nil
	}

	klog.Infof("scheduling pod: %s/%s", ns, podName)
	var err error
	for _, predicate := range predicatesByComponent {
		klog.Infof("entering predicate: %s, nodes: %v", predicate.Name(), predicates.GetNodeNames(kubeNodes))
		kubeNodes, err = predicate.Filter(instanceName, pod, kubeNodes)
		if err != nil {
			return nil, err
		}
		klog.Infof("leaving predicate: %s, nodes: %v", predicate.Name(), predicates.GetNodeNames(kubeNodes))
	}

	result := &schedulerapiv1.ExtenderFilterResult{
		Nodes: &corev1.NodeList{Items: kubeNodes},
	}

	return result, nil
}

// We don't pass `prioritizeVerb` to kubernetes scheduler extender's config file, this method will not be called.
func (s *scheduler) Priority(args *schedulerapiv1.ExtenderArgs) (schedulerapiv1.HostPriorityList, error) {
	result := schedulerapiv1.HostPriorityList{}

	var score int
	klog.Infof("Priority args:%+v", args)
	if args.Nodes != nil {
		predicatesByComponent, ok := s.predicates["ha"]
		if ok {
			for _, predicate := range predicatesByComponent {
				ret, err := predicate.Priority(args.Pod, args.Nodes.Items)
				if err == nil {
					return ret, nil
				}
			}
		}
		for _, node := range args.Nodes.Items {
			result = append(result, schedulerapiv1.HostPriority{
				Host:  node.Name,
				Score: score,
			})
		}
	}

	return result, nil
}

var _ Scheduler = &scheduler{}
