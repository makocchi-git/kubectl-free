package cmd

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/makocchi-git/kubectl-free/pkg/table"
	"github.com/makocchi-git/kubectl-free/pkg/util"

	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	clientv1 "k8s.io/client-go/kubernetes/typed/core/v1"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	"k8s.io/kubernetes/pkg/kubectl/util/templates"

	// Initialize all known client auth plugins.
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

var (
	// DfLong defines long description
	freeLong = templates.LongDesc(`
		Show various requested resources on Kubernetes nodes.
	`)

	// DfExample defines command examples
	freeExample = templates.Examples(`
		# Show pod resource usage of Kubernetes nodes (default namespace is "default").
		kubectl free

		# Show pod resource usage of Kubernetes nodes (all namespaces).
		kubectl free --all-namespaces

		# Show pod resource usage of Kubernetes nodes with number of pods and containers.
		kubectl free --pod

		# Using label selector.
		kubectl free -l key=value

		# Print raw(bytes) usage.
		kubectl free --bytes --without-unit

		# Using binary prefix unit (GiB, MiB, etc)
		kubectl free -g -B

		# List resources of containers in pods on nodes.
		kubectl free --list

		# List resources of containers in pods on nodes with image information.
		kubectl free --list --list-image

		# Print container even if that has no resources/limits.
		kubectl free --list --list-all

		# Do you like emoji? 😃
		kubectl free --emoji
		kubectl free --list --emoji
	`)
)

// FreeOptions is struct of df options
type FreeOptions struct {
	configFlags *genericclioptions.ConfigFlags
	genericclioptions.IOStreams

	// general options
	labelSelector string
	table         *table.OutputTable
	header        string
	pod           bool
	emojiStatus   bool
	allNamespaces bool

	// unit options
	bytes       bool
	kByte       bool
	mByte       bool
	gByte       bool
	withoutUnit bool
	binPrefix   bool

	// color output options
	nocolor       bool
	warnThreshold int64
	critThreshold int64

	// list options
	list               bool
	listContainerImage bool
	listAll            bool

	// k8s clients
	nodeClient clientv1.NodeInterface
	podClient  clientv1.PodInterface

	// table headers
	freeTableHeaders []string
	listTableHeaders []string
}

// NewFreeOptions is an instance of FreeOptions
func NewFreeOptions(streams genericclioptions.IOStreams) *FreeOptions {
	return &FreeOptions{
		configFlags:        genericclioptions.NewConfigFlags(true),
		bytes:              false,
		kByte:              false,
		mByte:              false,
		gByte:              false,
		withoutUnit:        false,
		binPrefix:          false,
		nocolor:            false,
		warnThreshold:      25,
		critThreshold:      50,
		IOStreams:          streams,
		labelSelector:      "",
		list:               false,
		listContainerImage: false,
		listAll:            false,
		pod:                false,
		emojiStatus:        false,
		table:              table.NewOutputTable(os.Stdout),
		header:             "default",
		allNamespaces:      false,
	}
}

// NewCmdFree is a cobra command wrapping
func NewCmdFree(f cmdutil.Factory, streams genericclioptions.IOStreams, version, commit, date string) *cobra.Command {
	o := NewFreeOptions(streams)

	cmd := &cobra.Command{
		Use:     "kubectl free",
		Short:   "Show various requested resources on Kubernetes nodes.",
		Long:    freeLong,
		Example: freeExample,
		Version: version,
		Run: func(c *cobra.Command, args []string) {
			cmdutil.CheckErr(o.Complete(f, c, args))
			cmdutil.CheckErr(o.Validate())
			cmdutil.CheckErr(o.Run(args))
		},
	}

	// bool options
	cmd.Flags().BoolVarP(&o.bytes, "bytes", "b", o.bytes, `Use 1-byte (1-Byte) blocks rather than the default.`)
	cmd.Flags().BoolVarP(&o.kByte, "kilobytes", "k", o.kByte, `Use 1024-byte (1-Kbyte) blocks rather than the default.`)
	cmd.Flags().BoolVarP(&o.mByte, "megabytes", "m", o.mByte, `Use 1048576-byte (1-Mbyte) blocks rather than the default.`)
	cmd.Flags().BoolVarP(&o.gByte, "gigabytes", "g", o.gByte, `Use 1073741824-byte (1-Gbyte) blocks rather than the default.`)
	cmd.Flags().BoolVarP(&o.binPrefix, "binary-prefix", "B", o.binPrefix, `Use 1024 for basic unit calculation instead of 1000. (print like "KiB")`)
	cmd.Flags().BoolVarP(&o.withoutUnit, "without-unit", "", o.withoutUnit, `Do not print size with unit string.`)
	cmd.Flags().BoolVarP(&o.nocolor, "no-color", "", o.nocolor, `Print without ansi color.`)
	cmd.Flags().BoolVarP(&o.pod, "pod", "p", o.pod, `Show pod count and limit.`)
	cmd.Flags().BoolVarP(&o.list, "list", "", o.list, `Show container list on node.`)
	cmd.Flags().BoolVarP(&o.listContainerImage, "list-image", "", o.listContainerImage, `Show pod list on node with container image.`)
	cmd.Flags().BoolVarP(&o.listAll, "list-all", "", o.listAll, `Show pods even if they have no requests/limit`)
	cmd.Flags().BoolVarP(&o.emojiStatus, "emoji", "", o.emojiStatus, `Let's smile!! 😃 😭`)
	cmd.Flags().BoolVarP(&o.allNamespaces, "all-namespaces", "", o.allNamespaces, `If present, list pod resources(limits) across all namespaces. Namespace in current context is ignored even if specified with --namespace.`)

	// int64 options
	cmd.Flags().Int64VarP(&o.warnThreshold, "warn-threshold", "", o.warnThreshold, `Threshold of warn(yellow) color for USED column.`)
	cmd.Flags().Int64VarP(&o.critThreshold, "crit-threshold", "", o.critThreshold, `Threshold of critical(red) color for USED column.`)

	// string option
	cmd.Flags().StringVarP(&o.labelSelector, "selector", "l", o.labelSelector, `Selector (label query) to filter on.`)
	cmd.Flags().StringVarP(&o.header, "header", "", o.header, `header style. ["default", "verbose", "none"] are allowed.`)

	o.configFlags.AddFlags(cmd.Flags())

	// add the klog flags
	cmd.PersistentFlags().AddGoFlagSet(flag.CommandLine)

	// version command template
	cmd.SetVersionTemplate("Version: " + version + ", GitCommit: " + commit + ", BuildDate: " + date + "\n")

	return cmd
}

// Complete prepares k8s clients
func (o *FreeOptions) Complete(f cmdutil.Factory, cmd *cobra.Command, args []string) error {

	// get k8s client
	client, err := f.KubernetesClientSet()
	if err != nil {
		return err
	}

	// node client
	o.nodeClient = client.CoreV1().Nodes()

	// pod client
	if o.allNamespaces {
		// --all-namespace flag
		o.podClient = client.CoreV1().Pods(v1.NamespaceAll)
	} else {
		if *o.configFlags.Namespace == "" {
			// default namespace is "default"
			o.podClient = client.CoreV1().Pods(v1.NamespaceDefault)
		} else {
			// targeted namespace (--namespace flag)
			o.podClient = client.CoreV1().Pods(*o.configFlags.Namespace)
		}
	}

	// prepare table header
	o.prepareFreeTableHeader()
	o.prepareListTableHeader()

	return nil
}

// Validate ensures that all required arguments and flag values are provided
func (o *FreeOptions) Validate() error {

	// validate threshold
	if err := util.ValidateThreshold(o.warnThreshold, o.critThreshold); err != nil {
		return err
	}

	// validate header option
	if err := util.ValidateHeaderOpt(o.header); err != nil {
		return err
	}

	return nil
}

// Run printing disk usage of images
func (o *FreeOptions) Run(args []string) error {

	// get nodes
	nodes, err := util.GetNodes(o.nodeClient, args, o.labelSelector)
	if err != nil {
		return nil
	}

	// list pods and return
	if o.list {
		if err := o.listPodsOnNode(nodes); err != nil {
			return err
		}
		return nil
	}

	// print cpu/mem/pod resource usage
	if err := o.showFree(nodes); err != nil {
		return err
	}

	return nil
}

// prepareFreeTableHeader defines table headers for free usage
func (o *FreeOptions) prepareFreeTableHeader() {

	fth := []string{}

	switch o.header {
	case "default", "d":
		fth = []string{
			"NAME",
			"STATUS",
			"CPU/req",
			"CPU/alloc",
			"CPU/%",
			"MEM/req",
			"MEM/alloc",
			"MEM/%",
		}

		if o.pod {
			fth = append(fth, "PODS", "PODS/alloc", "CONTAINERS")
		}

	case "verbose", "v":
		fth = []string{
			"NAME",
			"STATUS",
			"CPU requested",
			"CPU allocatable",
			"CPU %USED",
			"Memory requested",
			"Memory allocatable",
			"Memory %USED",
		}

		if o.pod {
			fth = append(fth, "PODS", "PODS allocation", "CONTAINERS")
		}

	}

	if !o.nocolor && len(fth) > 0 {
		// hack: avoid breaking column by escape char
		util.DefaultColor(&fth[1]) // STATUS
		util.DefaultColor(&fth[4]) // CPU/%
		util.DefaultColor(&fth[7]) // MEM/%
	}

	o.freeTableHeaders = fth
}

// prepareListTableHeader defines table headers for --list
func (o *FreeOptions) prepareListTableHeader() {

	lth := []string{}

	switch o.header {
	case "default", "d":
		lth = []string{
			"NODE NAME",
			"POD",
			"POD IP",
			"POD STATUS",
			"NAMESPACE",
			"CONTAINER",
			"CPU/req",
			"CPU/lim",
			"MEM/req",
			"MEM/lim",
		}

		if o.listContainerImage {
			lth = append(lth, "IMAGE")
		}

	case "verbose", "v":
		lth = []string{
			"NODE NAME",
			"POD",
			"POD IP",
			"POD STATUS",
			"NAMESPACE",
			"CONTAINER",
			"CPU requested",
			"CPU limit",
			"MEM requested",
			"MEM limit",
		}

		if o.listContainerImage {
			lth = append(lth, "IMAGE")
		}
	}

	if !o.nocolor && len(lth) > 0 {
		// hack: avoid breaking column by escape char
		util.DefaultColor(&lth[3]) // POD STATUS
	}

	o.listTableHeaders = lth
}

// showFree prints requested and allocatable resources
func (o *FreeOptions) showFree(nodes []v1.Node) error {

	// set table header
	o.table.AddHeader(o.freeTableHeaders)

	// node loop
	for _, node := range nodes {

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
		cpuRequested, memRequested, _, _ := util.GetPodResources(*pods)

		// get cpu allocatable
		cpuAllocatable := node.Status.Allocatable.Cpu().MilliValue()

		// get memoly allocatable
		memAllocatable := node.Status.Allocatable.Memory().Value()

		// get usage
		cpuUsed := util.GetPercentage(cpuRequested, cpuAllocatable)
		memUsed := util.GetPercentage(memRequested, memAllocatable)

		row := []string{
			nodeName,                            // node name
			nodeStatus,                          // node status
			o.toMilliUnitOrDash(cpuRequested),   // cpu requested
			o.toMilliUnitOrDash(cpuAllocatable), // cpu allocatable
			o.toColorPercent(cpuUsed),           // cpu used %
			o.toUnitOrDash(memRequested),        // mem requested
			o.toUnitOrDash(memAllocatable),      // mem allocatable
			o.toColorPercent(memUsed),           // mem used %
		}

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

func (o *FreeOptions) listPodsOnNode(nodes []v1.Node) error {

	// set table header
	o.table.AddHeader(o.listTableHeaders)

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

			// pod name
			podName := pod.ObjectMeta.Name
			podNamespace := pod.ObjectMeta.Namespace
			podIP := pod.Status.PodIP
			podStatus := util.GetPodStatus(string(pod.Status.Phase), o.nocolor, o.emojiStatus)

			// container loop
			for _, container := range pod.Spec.Containers {
				containerName := container.Name
				containerImage := container.Image
				cCpuRequested := container.Resources.Requests.Cpu().MilliValue()
				cCpuLimit := container.Resources.Limits.Cpu().MilliValue()
				cMemRequested := container.Resources.Requests.Memory().Value()
				cMemLimit := container.Resources.Limits.Memory().Value()

				// skip if the requested/limit resources are not set
				if !o.listAll {
					if cCpuRequested == 0 && cCpuLimit == 0 && cMemRequested == 0 && cMemLimit == 0 {
						continue
					}
				}

				row := []string{
					nodeName,                           // node name
					podName,                            // pod name
					podIP,                              // pod ip
					podStatus,                          // pod status
					podNamespace,                       // namespace
					containerName,                      // container name
					o.toMilliUnitOrDash(cCpuRequested), // container CPU requested
					o.toMilliUnitOrDash(cCpuLimit),     // container CPU limit
					o.toUnitOrDash(cMemRequested),      // Memory requested
					o.toUnitOrDash(cMemLimit),          // Memory limit
				}

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

// toUnit calculate and add unit for int64
func (o *FreeOptions) toUnit(i int64) string {

	var unitbytes int64
	var unitstr string

	if o.binPrefix {
		unitbytes, unitstr = util.GetBinUnit(o.bytes, o.kByte, o.mByte, o.gByte)
	} else {
		unitbytes, unitstr = util.GetSiUnit(o.bytes, o.kByte, o.mByte, o.gByte)
	}

	// -H adds human readable unit
	unit := ""
	if !o.withoutUnit {
		unit = unitstr
	}

	return strconv.FormatInt(i/unitbytes, 10) + unit
}

// toUnitOrDash returns "-" if "i" is 0, otherwise returns toUnit()
func (o *FreeOptions) toUnitOrDash(i int64) string {

	if i == 0 {
		return "-"
	}

	return o.toUnit(i)
}

// toMilliUnitOrDash returns "-" if "i" is 0, otherwise returns MilliQuantity
func (o *FreeOptions) toMilliUnitOrDash(i int64) string {

	if i == 0 {
		return "-"
	}

	if o.withoutUnit {
		// return raw value
		return strconv.FormatInt(i, 10)
	}

	return resource.NewMilliQuantity(i, resource.DecimalSI).String()
}

// toColorPercent returns colored strings
//        percentage < warn : Green
// warn < percentage < crit : Yellow
// crit < percentage        : Red
func (o *FreeOptions) toColorPercent(i int64) string {
	p := strconv.FormatInt(i, 10) + "%"

	if o.nocolor {
		// nothing to do
		return p
	}

	switch {
	case i < o.warnThreshold:
		// percentage < warn : Green
		util.Green(&p)
	case i < o.critThreshold:
		// warn < percentage < crit : Yellow
		util.Yellow(&p)
	default:
		// crit < percentage : Red
		util.Red(&p)
	}

	return p
}
