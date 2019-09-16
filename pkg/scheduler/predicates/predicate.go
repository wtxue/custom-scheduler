package predicates

import (
	corev1 "k8s.io/api/core/v1"
	schedulerapiv1 "k8s.io/kubernetes/pkg/scheduler/api/v1"
	"sort"
)

// Predicate is an interface as extender-implemented predicate functions
type Predicate interface {
	// Name return the predicate name
	Name() string

	// Filter function receives a set of nodes and returns a set of candidate nodes.
	Filter(string, *corev1.Pod, []corev1.Node) ([]corev1.Node, error)

	// Priority function receives a set of HostPriorityList.
	Priority(*corev1.Pod, []corev1.Node) (schedulerapiv1.HostPriorityList, error)
}

func getNodeFromNames(nodes []corev1.Node, nodeNames []string) []corev1.Node {
	var retNodes []corev1.Node
	for _, node := range nodes {
		for _, nodeName := range nodeNames {
			if node.GetName() == nodeName {
				retNodes = append(retNodes, node)
				break
			}
		}
	}
	return retNodes
}

func GetNodeNames(nodes []corev1.Node) []string {
	nodeNames := make([]string, 0)
	for _, node := range nodes {
		nodeNames = append(nodeNames, node.GetName())
	}
	sort.Strings(nodeNames)
	return nodeNames
}
