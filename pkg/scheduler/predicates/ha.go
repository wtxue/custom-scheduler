package predicates

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog"
	schedulerapiv1 "k8s.io/kubernetes/pkg/scheduler/api/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sync"
	"github.com/xkcp0324/custom-scheduler/pkg/observe"
)



type ha struct {
	lock     sync.Mutex
	kubeCli  kubernetes.Interface
	mgr      manager.Manager
	recorder record.EventRecorder
}

// NewHA returns a Predicate
func NewHA(kubeCli kubernetes.Interface, mgr manager.Manager) Predicate {
	h := &ha{
		kubeCli: kubeCli,
		mgr:     mgr,
	}

	return h
}

func (h *ha) Name() string {
	return "HighAvailability"
}

func (h *ha) Filter(instanceName string, pod *corev1.Pod, nodes []corev1.Node) ([]corev1.Node, error) {
	return nodes, nil
}

func (h *ha) Priority(pod *corev1.Pod, nodes []corev1.Node) (schedulerapiv1.HostPriorityList, error) {
	h.lock.Lock()
	defer h.lock.Unlock()

	result := schedulerapiv1.HostPriorityList{}
	ns := pod.GetNamespace()
	if len(nodes) == 0 {
		return nil, fmt.Errorf("kube nodes is empty")
	}

	instanceName, ok := pod.Labels[observe.ObserveMustLabelAppName]
	if !ok {
		klog.Errorf("no find pod label: %s", observe.ObserveMustLabelAppName)
		return nil, fmt.Errorf("no find pod label")
	}

	if len(nodes) == 1 {
		result = append(result, schedulerapiv1.HostPriority{
			Host:  nodes[0].Name,
			Score: 0,
		})
		return result, nil
	}

	// podList, err := h.podLister.Pods(ns).List(labels.SelectorFromSet(labels.Set{"app": instanceName}))
	// if err != nil {
	// 	klog.Errorf("list pod err: %+v", err)
	// 	return nil, err
	// }

	cl := h.mgr.GetClient()
	podList := &corev1.PodList{}

	// err := cl.List(context.Background(), client.InNamespace(ns).MatchingLabels(map[string]string{"app": instanceName}), podList)
	err := cl.List(context.Background(), podList, client.InNamespace(ns), client.MatchingLabels{"app": instanceName})
	if err != nil {
		klog.Errorf("list pod err: %+v", err)
		return nil, err
	}

	for _, node := range nodes {
		result = append(result, schedulerapiv1.HostPriority{
			Host:  node.Name,
			Score: 100,
		})
	}

	for _, pod := range podList.Items {
		nodeName := pod.Spec.NodeName
		if nodeName == "" {
			continue
		}

		for i := range result {
			if result[i].Host == nodeName {
				if result[i].Score > 0 {
					result[i].Score--
				}
			}
		}
	}

	klog.V(3).Infof("result: %+v", result)
	return result, nil
}
