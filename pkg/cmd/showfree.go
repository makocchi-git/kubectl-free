package cmd

import (
	"fmt"
	"strconv"

	"github.com/makocchi-git/kubectl-free/pkg/util"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// showFree prints requested and allocatable resources
func (o *FreeOptions) showFree(nodes []v1.Node) error {

	// set table header
	if !o.noHeaders {
		o.table.Header = o.freeTableHeaders
	}

	// node loop
	for _, node := range nodes {

		var (
			cpuMetricsUsed  int64
			memMetricsUsed  int64
			cpuMetricsUsedP int64
			memMetricsUsedP int64
		)

		// node name
		nodeName := node.ObjectMeta.Name

		// node status
		nodeStatus, err := util.GetNodeStatus(node, o.emojiStatus)
		if err != nil {
			return err
		}

		util.SetNodeStatusColor(&nodeStatus, o.nocolor)

		// get pods on node
		pods, perr := util.GetPods(o.podClient, nodeName)
		if perr != nil {
			return perr
		}

		// calculate requested resources by pods
		cpuRequested, memRequested, cpuLimited, memLimited := util.GetPodResources(*pods)

		// get cpu allocatable
		cpuAllocatable := node.Status.Allocatable.Cpu().MilliValue()

		// get memoly allocatable
		memAllocatable := node.Status.Allocatable.Memory().Value()

		// get usage
		cpuRequestedP := util.GetPercentage(cpuRequested, cpuAllocatable)
		cpuLimitedP := util.GetPercentage(cpuLimited, cpuAllocatable)
		memRequestedP := util.GetPercentage(memRequested, memAllocatable)
		memLimitedP := util.GetPercentage(memLimited, memAllocatable)

		// get metrics
		if !o.noMetrics && o.metricsNodeClient != nil {
			nodeMetrics, err := o.metricsNodeClient.Get(nodeName, metav1.GetOptions{})
			if err == nil {
				cpuMetricsUsed = nodeMetrics.Usage.Cpu().MilliValue()
				memMetricsUsed = nodeMetrics.Usage.Memory().Value()
				cpuMetricsUsedP = util.GetPercentage(cpuMetricsUsed, cpuAllocatable)
				memMetricsUsedP = util.GetPercentage(memMetricsUsed, memAllocatable)
			}
			// ignore fetching metrics error
		}

		// create table row
		// basic row
		row := []string{
			nodeName,   // node name
			nodeStatus, // node status
		}

		// cpu
		if !o.noMetrics {
			row = append(row, o.toMilliUnitOrDash(cpuMetricsUsed)) // cpu used (from metrics)
		}
		row = append(
			row,
			o.toMilliUnitOrDash(cpuRequested),   // cpu requested
			o.toMilliUnitOrDash(cpuLimited),     // cpu limited
			o.toMilliUnitOrDash(cpuAllocatable), // cpu allocatable
		)
		if !o.noMetrics {
			row = append(row, o.toColorPercent(cpuMetricsUsedP)) // cpu used %
		}
		row = append(
			row,
			o.toColorPercent(cpuRequestedP), // cpu requested %
			o.toColorPercent(cpuLimitedP),   // cpu limited %
		)

		// mem
		if !o.noMetrics {
			row = append(row, o.toUnitOrDash(memMetricsUsed)) // mem used (from metrics)
		}
		row = append(
			row,
			o.toUnitOrDash(memRequested),   // mem requested
			o.toUnitOrDash(memLimited),     // mem limited
			o.toUnitOrDash(memAllocatable), // mem allocatable
		)
		if !o.noMetrics {
			row = append(row, o.toColorPercent(memMetricsUsedP)) // mem used %
		}
		row = append(
			row,
			o.toColorPercent(memRequestedP), // mem requested %
			o.toColorPercent(memLimitedP),   // mem limited %
		)

		// show pod and container (--pod option)
		if o.pod {

			// pod count
			podCount := util.GetPodCount(*pods)

			// container count
			containerCount := util.GetContainerCount(*pods)

			// get pod allocatable
			podAllocatable := node.Status.Allocatable.Pods().Value()

			row = append(
				row,
				fmt.Sprintf("%d", podCount),           // pod used
				strconv.FormatInt(podAllocatable, 10), // pod allocatable
				fmt.Sprintf("%d", containerCount),     // containers
			)
		}

		o.table.AddRow(row)
	}

	o.table.Print()

	return nil
}
