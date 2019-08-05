package cmd

import (
	"time"

	"github.com/makocchi-git/kubectl-free/pkg/util"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/duration"
	metricsapiv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

func (o *FreeOptions) showPodsOnNode(nodes []v1.Node) error {

	// set table header
	if !o.noHeaders {
		o.table.Header = o.listTableHeaders
	}

	// get pod metrics
	var podMetrics *metricsapiv1beta1.PodMetricsList
	if !o.noMetrics && o.metricsPodClient != nil {
		podMetrics, _ = o.metricsPodClient.List(metav1.ListOptions{})
	}

	// node loop
	for _, node := range nodes {

		// node name
		nodeName := node.ObjectMeta.Name

		// get pods on node
		pods, perr := util.GetPods(o.podClient, nodeName)
		if perr != nil {
			return perr
		}

		// node loop
		for _, pod := range pods.Items {

			var containerCPUUsed int64
			var containerMEMUsed int64

			// pod information
			podName := pod.ObjectMeta.Name
			podNamespace := pod.ObjectMeta.Namespace
			podIP := pod.Status.PodIP
			podStatus := util.GetPodStatus(string(pod.Status.Phase), o.nocolor, o.emojiStatus)
			podCreationTime := pod.ObjectMeta.CreationTimestamp.UTC()
			podCreationTimeDiff := time.Since(podCreationTime)
			podAge := "<unknown>"
			if !podCreationTime.IsZero() {
				podAge = duration.HumanDuration(podCreationTimeDiff)
			}

			// container loop
			for _, container := range pod.Spec.Containers {
				containerName := container.Name
				containerImage := container.Image
				cCpuRequested := container.Resources.Requests.Cpu().MilliValue()
				cCpuLimit := container.Resources.Limits.Cpu().MilliValue()
				cMemRequested := container.Resources.Requests.Memory().Value()
				cMemLimit := container.Resources.Limits.Memory().Value()

				if !o.noMetrics && podMetrics != nil {
					containerCPUUsed, containerMEMUsed = util.GetContainerMetrics(podMetrics, podName, containerName)
				}

				// skip if the requested/limit resources are not set
				if !o.listAll {
					if cCpuRequested == 0 && cCpuLimit == 0 && cMemRequested == 0 && cMemLimit == 0 {
						continue
					}
				}

				row := []string{
					nodeName,      // node name
					podNamespace,  // namespace
					podName,       // pod name
					podAge,        // pod age
					podIP,         // pod ip
					podStatus,     // pod status
					containerName, // container name
				}

				if !o.noMetrics {
					row = append(row, o.toMilliUnitOrDash(containerCPUUsed))
				}

				row = append(
					row,
					o.toMilliUnitOrDash(cCpuRequested), // container CPU requested
					o.toMilliUnitOrDash(cCpuLimit),     // container CPU limit
				)

				if !o.noMetrics {
					row = append(row, o.toUnitOrDash(containerMEMUsed))
				}

				row = append(
					row,
					o.toUnitOrDash(cMemRequested), // Memory requested
					o.toUnitOrDash(cMemLimit),     // Memory limit
				)

				if o.listContainerImage {
					row = append(row, containerImage)
				}

				o.table.AddRow(row)
			}
		}
	}
	o.table.Print()

	return nil
}
