package util

import (
	"fmt"
	"strings"

	color "github.com/gookit/color"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientv1 "k8s.io/client-go/kubernetes/typed/core/v1"
	metricsapiv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"

	"github.com/makocchi-git/kubectl-free/pkg/constants"
)

// GetSiUnit defines unit for usage (SI prefix)
// If multiple options are selected, returns a biggest unit
func GetSiUnit(b, k, m, g bool) (int64, string) {

	if g {
		return constants.UnitGigaBytes, constants.UnitGigaBytesStr
	}

	if m {
		return constants.UnitMegaBytes, constants.UnitMegaBytesStr
	}

	if k {
		return constants.UnitKiloBytes, constants.UnitKiloBytesStr
	}

	if b {
		return constants.UnitBytes, constants.UnitBytesStr
	}

	// default output is "kilobytes"
	return constants.UnitKiloBytes, constants.UnitKiloBytesStr
}

// GetBinUnit defines unit for usage (Binary prefix)
// If multiple options are selected, returns a biggest unit
func GetBinUnit(b, k, m, g bool) (int64, string) {

	if g {
		return constants.UnitGibiBytes, constants.UnitGibiBytesStr
	}

	if m {
		return constants.UnitMibiBytes, constants.UnitMibiBytesStr
	}

	if k {
		return constants.UnitKibiBytes, constants.UnitKibiBytesStr
	}

	if b {
		return constants.UnitBytes, constants.UnitBytesStr
	}

	// default output is "kibibytes"
	return constants.UnitKibiBytes, constants.UnitKibiBytesStr
}

// JoinTab joins string slice with tab
func JoinTab(s []string) string {
	return strings.Join(s, "\t")
}

// SetPercentageColor returns colored string
//        percentage < warn : Green
// warn < percentage < crit : Yellow
// crit < percentage        : Red
func SetPercentageColor(s *string, p, warn, crit int64) {
	green := color.FgGreen.Render
	yellow := color.FgYellow.Render
	red := color.FgRed.Render

	if p < warn {
		*s = green(*s)
		return
	}

	if p < crit {
		*s = yellow(*s)
		return
	}

	*s = red(*s)
}

// SetNodeStatusColor defined color of node status
func SetNodeStatusColor(status *string, nocolor bool) {

	if nocolor {
		// nothing to do
		return
	}

	if *status != "Ready" {
		// Red
		*status = color.FgRed.Render(*status)
		return
	}

	// Green
	*status = color.FgGreen.Render(*status)

}

// GetPodStatus defined pod status with color
func GetPodStatus(status string, nocolor, emoji bool) string {

	var s string

	switch status {
	case string(v1.PodRunning):
		if emoji {
			s = constants.EmojiPodRunning
		} else {
			s = string(v1.PodRunning)
		}

		if !nocolor {
			Green(&s)
		}
	case string(v1.PodSucceeded):
		if emoji {
			s = constants.EmojiPodSucceeded
		} else {
			s = string(v1.PodSucceeded)
		}

		if !nocolor {
			Green(&s)
		}
	case string(v1.PodPending):
		if emoji {
			s = constants.EmojiPodPending
		} else {
			s = string(v1.PodPending)
		}

		if !nocolor {
			Yellow(&s)
		}
	case string(v1.PodFailed):
		if emoji {
			s = constants.EmojiPodFailed
		} else {
			s = string(v1.PodFailed)
		}

		if !nocolor {
			Red(&s)
		}
	default:
		if emoji {
			s = constants.EmojiPodUnknown
		} else {
			s = "Unknown"
		}

		if !nocolor {
			DefaultColor(&s)
		}
	}

	return s
}

// GetNodes returns node objects
func GetNodes(c clientv1.NodeInterface, args []string, label string) ([]v1.Node, error) {
	nodes := []v1.Node{}

	if len(args) > 0 {
		for _, a := range args {
			n, nerr := c.Get(a, metav1.GetOptions{})
			if nerr != nil {
				return nodes, fmt.Errorf("failed to get node: %v", nerr)
			}
			nodes = append(nodes, *n)
		}
	} else {
		na, naerr := c.List(metav1.ListOptions{LabelSelector: label})
		if naerr != nil {
			return nodes, fmt.Errorf("failed to list nodes: %v", naerr)
		}
		nodes = append(nodes, na.Items...)
	}

	return nodes, nil
}

// GetNodeStatus returns node status
func GetNodeStatus(node v1.Node, emoji bool) (string, error) {
	status := "NotReady"

	for _, condition := range node.Status.Conditions {
		if condition.Type == v1.NodeReady && condition.Status == v1.ConditionTrue {
			status = "Ready"
		}
	}

	if emoji {
		switch status {
		case "Ready":
			status = constants.EmojiReady
		case "NotReady":
			status = constants.EmojiNotReady
		}
	}

	return status, nil
}

// GetPods returns node objects
func GetPods(c clientv1.PodInterface, nodeName string) (*v1.PodList, error) {

	pods, err := c.List(metav1.ListOptions{FieldSelector: "spec.nodeName=" + nodeName})
	if err != nil {
		return pods, fmt.Errorf("failed to get pods: %s", err)
	}

	return pods, nil
}

// GetContainerMetrics returns container metrics usage
func GetContainerMetrics(metrics *metricsapiv1beta1.PodMetricsList, podName, containerName string) (cpu, mem int64) {

	var c int64
	var m int64

	for _, pod := range metrics.Items {
		if pod.ObjectMeta.Name == podName {
			for _, container := range pod.Containers {
				if container.Name == containerName {
					c = container.Usage.Cpu().MilliValue()
					m = container.Usage.Memory().Value()
					break
				}
			}
		}
	}

	// if no metrics found, return 0 0
	return c, m
}

// GetPodResources returns sum of requested/limit resources
func GetPodResources(pods v1.PodList) (int64, int64, int64, int64) {
	var rc, rm, lc, lm int64

	for _, pod := range pods.Items {

		// skip if pod status is not running
		if pod.Status.Phase != v1.PodRunning {
			continue
		}

		for _, container := range pod.Spec.Containers {
			rc += container.Resources.Requests.Cpu().MilliValue()
			lc += container.Resources.Limits.Cpu().MilliValue()
			rm += container.Resources.Requests.Memory().Value()
			lm += container.Resources.Limits.Memory().Value()
		}
	}

	return rc, rm, lc, lm
}

// GetPodCount returns count of pods
func GetPodCount(pods v1.PodList) int {
	return len(pods.Items)
}

// GetContainerCount returns count of containers
func GetContainerCount(pods v1.PodList) int {
	var c int
	for _, pod := range pods.Items {
		c += len(pod.Spec.Containers)
	}
	return c
}

// GetPercentage returns (a*100)/b
func GetPercentage(a, b int64) int64 {
	// avoid 0 divide
	if b == 0 {
		return 0
	}
	return (a * 100) / b
}

// DefaultColor set default color
func DefaultColor(s *string) {
	// add dummy escape code
	*s = color.FgDefault.Render(*s)
}

// Green is coloring string to green
func Green(s *string) {
	*s = color.FgGreen.Render(*s)
}

// Red is coloring string to red
func Red(s *string) {
	*s = color.FgRed.Render(*s)
}

// Yellow is coloring string to yellow
func Yellow(s *string) {
	*s = color.FgYellow.Render(*s)
}
