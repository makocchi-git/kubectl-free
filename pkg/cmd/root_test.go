package cmd

import (
	"bytes"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/makocchi-git/kubectl-free/pkg/table"
	"github.com/makocchi-git/kubectl-free/pkg/util"

	color "github.com/gookit/color"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	fake "k8s.io/client-go/kubernetes/fake"
	core "k8s.io/client-go/testing"
	cmdtesting "k8s.io/kubernetes/pkg/kubectl/cmd/testing"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	metricsapiv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
	fakemetrics "k8s.io/metrics/pkg/client/clientset/versioned/fake"
)

// test node object
var testNodes = []v1.Node{
	{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "node1",
			Labels: map[string]string{"hostname": "node1"},
		},
		Status: v1.NodeStatus{
			Allocatable: v1.ResourceList{
				v1.ResourceCPU:    *resource.NewMilliQuantity(4000, resource.DecimalSI),
				v1.ResourceMemory: *resource.NewQuantity(4000, resource.DecimalSI),
				v1.ResourcePods:   *resource.NewQuantity(110, resource.DecimalSI),
			},
			Conditions: []v1.NodeCondition{
				{
					Type:   v1.NodeReady,
					Status: v1.ConditionTrue,
				},
			},
		},
	},
	{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "node2",
			Labels: map[string]string{"hostname": "node1"},
		},
		Status: v1.NodeStatus{
			Allocatable: v1.ResourceList{
				v1.ResourceCPU:    *resource.NewMilliQuantity(8000, resource.DecimalSI),
				v1.ResourceMemory: *resource.NewQuantity(8000, resource.DecimalSI),
				v1.ResourcePods:   *resource.NewQuantity(110, resource.DecimalSI),
			},
			Conditions: []v1.NodeCondition{
				{
					Type:   v1.NodeReady,
					Status: v1.ConditionFalse,
				},
			},
		},
	},
}

// test pod object
var testPods = []v1.Pod{
	{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod1",
			Namespace: "default",
		},
		Status: v1.PodStatus{
			PodIP: "1.2.3.4",
			Phase: v1.PodRunning,
		},
		Spec: v1.PodSpec{
			NodeName: "node1",
			Containers: []v1.Container{
				{
					Name:  "container1",
					Image: "alpine:latest",
					Resources: v1.ResourceRequirements{
						Limits: v1.ResourceList{
							v1.ResourceCPU:    *resource.NewMilliQuantity(2000, resource.DecimalSI),
							v1.ResourceMemory: *resource.NewQuantity(2000, resource.DecimalSI),
						},
						Requests: v1.ResourceList{
							v1.ResourceCPU:    *resource.NewMilliQuantity(1000, resource.DecimalSI),
							v1.ResourceMemory: *resource.NewQuantity(1000, resource.DecimalSI),
						},
					},
				},
			},
		},
	},
	{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod2",
			Namespace: "default",
		},
		Status: v1.PodStatus{
			PodIP: "2.3.4.5",
			Phase: v1.PodRunning,
		},
		Spec: v1.PodSpec{
			NodeName: "node1",
			Containers: []v1.Container{
				{
					Name:  "container2a",
					Image: "nginx:latest",
					Resources: v1.ResourceRequirements{
						Limits: v1.ResourceList{
							v1.ResourceCPU:    *resource.NewMilliQuantity(500, resource.DecimalSI),
							v1.ResourceMemory: *resource.NewQuantity(1000, resource.DecimalSI),
						},
						Requests: v1.ResourceList{
							v1.ResourceCPU:    *resource.NewMilliQuantity(500, resource.DecimalSI),
							v1.ResourceMemory: *resource.NewQuantity(1000, resource.DecimalSI),
						},
					},
				},
				{
					Name:  "container2b",
					Image: "busybox:latest",
				},
			},
		},
	},
	{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod3",
			Namespace: "awesome-ns",
		},
		Status: v1.PodStatus{
			PodIP: "3.4.5.6",
			Phase: v1.PodRunning,
		},
		Spec: v1.PodSpec{
			NodeName: "node1",
			Containers: []v1.Container{
				{
					Name:  "container3a",
					Image: "centos:7",
					Resources: v1.ResourceRequirements{
						Limits: v1.ResourceList{
							v1.ResourceCPU:    *resource.NewMilliQuantity(200, resource.DecimalSI),
							v1.ResourceMemory: *resource.NewQuantity(300, resource.DecimalSI),
						},
						Requests: v1.ResourceList{
							v1.ResourceCPU:    *resource.NewMilliQuantity(200, resource.DecimalSI),
							v1.ResourceMemory: *resource.NewQuantity(300, resource.DecimalSI),
						},
					},
				},
				{
					Name:  "container3b",
					Image: "ubuntu:bionic",
				},
			},
		},
	},
}

var testPodMetrics = &metricsapiv1beta1.PodMetricsList{
	Items: []metricsapiv1beta1.PodMetrics{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "pod1",
				Namespace: "default",
				Labels: map[string]string{
					"key": "value",
				},
			},
			Timestamp: metav1.Now(),
			Containers: []metricsapiv1beta1.ContainerMetrics{
				{
					Name: "container1",
					Usage: v1.ResourceList{
						v1.ResourceCPU:    *resource.NewMilliQuantity(10, resource.DecimalSI),
						v1.ResourceMemory: *resource.NewQuantity(10, resource.DecimalSI),
					},
				},
			},
		},
	},
}

var testNodeMetrics = &metricsapiv1beta1.NodeMetricsList{
	Items: []metricsapiv1beta1.NodeMetrics{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "node1",
			},
			Usage: v1.ResourceList{
				v1.ResourceCPU:    *resource.NewMilliQuantity(100, resource.DecimalSI),
				v1.ResourceMemory: *resource.NewQuantity(1024, resource.DecimalSI),
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "node2",
			},
			Usage: v1.ResourceList{
				v1.ResourceCPU:    *resource.NewMilliQuantity(200, resource.DecimalSI),
				v1.ResourceMemory: *resource.NewQuantity(2048, resource.DecimalSI),
			},
		},
	},
}

func TestNewFreeOptions(t *testing.T) {
	streams := genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr}

	expected := &FreeOptions{
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
		noHeaders:          false,
		noMetrics:          false,
	}

	actual := NewFreeOptions(streams)

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("expected(%#v) differ (got: %#v)", expected, actual)
	}
}

func TestNewCmdFree(t *testing.T) {

	cmdtesting.InitTestErrorHandler(t)
	tf := cmdtesting.NewTestFactory()
	defer tf.Cleanup()
	tf.ClientConfigVal = cmdtesting.DefaultClientConfig()

	rootCmd := NewCmdFree(
		tf,
		genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr},
		"v0.0.1",
		"abcd123",
		"1234567890",
	)

	// Version check
	t.Run("version", func(t *testing.T) {
		expected := "Version: v0.0.1, GitCommit: abcd123, BuildDate: 1234567890\n"
		actual, err := executeCommand(rootCmd, "--version")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if actual != expected {
			t.Errorf("expected(%s) differ (got: %s)", expected, actual)
			return
		}
	})

	// Usage
	t.Run("usage", func(t *testing.T) {
		expected := "kubectl free [flags]"
		actual, err := executeCommand(rootCmd, "--help")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if !strings.Contains(actual, expected) {
			t.Errorf("expected(%s) differ (got: %s)", expected, actual)
			return
		}
	})

	// Unknown option
	t.Run("unknown option", func(t *testing.T) {
		expected := "unknown flag: --very-very-bad-option"
		_, err := executeCommand(rootCmd, "--very-very-bad-option")
		if err == nil {
			t.Errorf("unexpected error: should return exit")
			return
		}

		if err.Error() != expected {
			t.Errorf("expected(%s) differ (got: %s)", expected, err.Error())
			return
		}
	})
}

func TestComplete(t *testing.T) {

	cmdtesting.InitTestErrorHandler(t)
	tf := cmdtesting.NewTestFactory()
	defer tf.Cleanup()
	tf.ClientConfigVal = cmdtesting.DefaultClientConfig()

	rootCmd := NewCmdFree(
		tf,
		genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr},
		"v0.0.1",
		"abcd123",
		"1234567890",
	)

	kubeConfigFlags := genericclioptions.NewConfigFlags(true).WithDeprecatedPasswordFlag()
	matchVersionKubeConfigFlags := cmdutil.NewMatchVersionFlags(kubeConfigFlags)

	f := cmdutil.NewFactory(matchVersionKubeConfigFlags)

	t.Run("complete", func(t *testing.T) {

		o := &FreeOptions{
			configFlags: genericclioptions.NewConfigFlags(true),
		}

		if err := o.Complete(f, rootCmd, []string{}); err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
	})

	t.Run("complete allnamespace", func(t *testing.T) {

		o := &FreeOptions{
			configFlags:   genericclioptions.NewConfigFlags(true),
			allNamespaces: true,
		}

		if err := o.Complete(f, rootCmd, []string{}); err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
	})

	t.Run("complete specific namespace", func(t *testing.T) {

		o := &FreeOptions{
			configFlags: genericclioptions.NewConfigFlags(true),
		}
		*o.configFlags.Namespace = "awesome-ns"

		if err := o.Complete(f, rootCmd, []string{}); err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
	})
}

func TestValidate(t *testing.T) {

	t.Run("validate threshold", func(t *testing.T) {

		o := &FreeOptions{
			critThreshold: 5,
			warnThreshold: 10,
		}

		err := o.Validate()
		expected := "can not set critical threshold less than warn threshold (warn:10 crit:5)"
		if err.Error() != expected {
			t.Errorf("unexpected error: %v", err)
			return
		}
	})

	t.Run("validate success", func(t *testing.T) {

		o := &FreeOptions{
			warnThreshold: 25,
			critThreshold: 50,
		}

		if err := o.Validate(); err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
	})
}

func TestRun(t *testing.T) {

	var tests = []struct {
		description string
		args        []string
		listOption  bool
		expected    []string
	}{
		{
			"default free",
			[]string{},
			false,
			[]string{
				"NAME    STATUS   CPU/use   CPU/req   CPU/lim   CPU/alloc   CPU/use%   CPU/req%   CPU/lim%   MEM/use   MEM/req   MEM/lim   MEM/alloc   MEM/use%   MEM/req%   MEM/lim%",
				"node1   Ready    -         1         2         4           0%         25%        50%        -         1K        2K        4K          0%         25%        50%",
				"",
			},
		},
		{
			"defaul free --list",
			[]string{},
			true,
			[]string{
				"NODE NAME   NAMESPACE   POD NAME   POD AGE     POD IP    POD STATUS   CONTAINER    CPU/use   CPU/req   CPU/lim   MEM/use   MEM/req   MEM/lim",
				"node1       default     pod1       <unknown>   1.2.3.4   Running      container1   -         1         2         -         1K        2K",
				"",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {

			fakeNodeClient := fake.NewSimpleClientset(&testNodes[0])
			fakePodClient := fake.NewSimpleClientset(&testPods[0])
			fakeMetricsNodeClient := prepareTestNodeMetricsClient()
			fakeMetricsPodClient := prepareTestPodMetricsClient()

			buffer := &bytes.Buffer{}
			o := &FreeOptions{
				nocolor:           true,
				table:             table.NewOutputTable(buffer),
				list:              test.listOption,
				nodeClient:        fakeNodeClient.CoreV1().Nodes(),
				podClient:         fakePodClient.CoreV1().Pods("default"),
				metricsPodClient:  fakeMetricsPodClient.MetricsV1beta1().PodMetricses("default"),
				metricsNodeClient: fakeMetricsNodeClient.MetricsV1beta1().NodeMetricses(),
			}

			o.prepareFreeTableHeader()
			o.prepareListTableHeader()

			if err := o.Run(test.args); err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			e := strings.Join(test.expected, "\n")
			if buffer.String() != e {
				t.Errorf("expected(%s) differ (got: %s)", e, buffer.String())
				return
			}
		})
	}
}

func TestPrepareFreeTableHeader(t *testing.T) {

	colorStatus := "STATUS"
	colorCPUreqP := "CPU/req%"
	colorCPUlimP := "CPU/lim%"
	colorMEMreqP := "MEM/req%"
	colorMEMlimP := "MEM/lim%"
	util.DefaultColor(&colorStatus)
	util.DefaultColor(&colorCPUreqP)
	util.DefaultColor(&colorCPUlimP)
	util.DefaultColor(&colorMEMreqP)
	util.DefaultColor(&colorMEMlimP)

	var tests = []struct {
		description string
		listPod     bool
		nocolor     bool
		noheader    bool
		nometrics   bool
		expected    []string
	}{
		{
			"default header",
			false,
			true,
			false,
			true,
			[]string{
				"NAME",
				"STATUS",
				"CPU/req",
				"CPU/lim",
				"CPU/alloc",
				"CPU/req%",
				"CPU/lim%",
				"MEM/req",
				"MEM/lim",
				"MEM/alloc",
				"MEM/req%",
				"MEM/lim%",
			},
		},
		{
			"default header with metrics",
			false,
			true,
			false,
			false,
			[]string{
				"NAME",
				"STATUS",
				"CPU/use",
				"CPU/req",
				"CPU/lim",
				"CPU/alloc",
				"CPU/use%",
				"CPU/req%",
				"CPU/lim%",
				"MEM/use",
				"MEM/req",
				"MEM/lim",
				"MEM/alloc",
				"MEM/use%",
				"MEM/req%",
				"MEM/lim%",
			},
		},
		{
			"default header with --pod",
			true,
			true,
			false,
			true,
			[]string{
				"NAME",
				"STATUS",
				"CPU/req",
				"CPU/lim",
				"CPU/alloc",
				"CPU/req%",
				"CPU/lim%",
				"MEM/req",
				"MEM/lim",
				"MEM/alloc",
				"MEM/req%",
				"MEM/lim%",
				"PODS",
				"PODS/alloc",
				"CONTAINERS",
			},
		},
		{
			"default header with color",
			false,
			false,
			false,
			true,
			[]string{
				"NAME",
				colorStatus,
				"CPU/req",
				"CPU/lim",
				"CPU/alloc",
				colorCPUreqP,
				colorCPUlimP,
				"MEM/req",
				"MEM/lim",
				"MEM/alloc",
				colorMEMreqP,
				colorMEMlimP,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			o := &FreeOptions{
				pod:       test.listPod,
				noHeaders: test.noheader,
				noMetrics: test.nometrics,
				nocolor:   test.nocolor,
			}
			o.prepareFreeTableHeader()

			if !reflect.DeepEqual(o.freeTableHeaders, test.expected) {
				t.Errorf(
					"[%s] expected(%v) differ (got: %v)",
					test.description,
					test.expected,
					o.freeTableHeaders,
				)
				return
			}
		})
	}
}

func TestPrepareListTableHeader(t *testing.T) {

	colorStatus := "POD STATUS"
	util.DefaultColor(&colorStatus)

	var tests = []struct {
		description string
		listImage   bool
		nocolor     bool
		noheader    bool
		nometrics   bool
		expected    []string
	}{
		{
			"default header",
			false,
			true,
			false,
			true,
			[]string{
				"NODE NAME",
				"NAMESPACE",
				"POD NAME",
				"POD AGE",
				"POD IP",
				"POD STATUS",
				"CONTAINER",
				"CPU/req",
				"CPU/lim",
				"MEM/req",
				"MEM/lim",
			},
		},
		{
			"default header with metrics",
			false,
			true,
			false,
			false,
			[]string{
				"NODE NAME",
				"NAMESPACE",
				"POD NAME",
				"POD AGE",
				"POD IP",
				"POD STATUS",
				"CONTAINER",
				"CPU/use",
				"CPU/req",
				"CPU/lim",
				"MEM/use",
				"MEM/req",
				"MEM/lim",
			},
		},
		{
			"default header with --list-image",
			true,
			true,
			false,
			true,
			[]string{
				"NODE NAME",
				"NAMESPACE",
				"POD NAME",
				"POD AGE",
				"POD IP",
				"POD STATUS",
				"CONTAINER",
				"CPU/req",
				"CPU/lim",
				"MEM/req",
				"MEM/lim",
				"IMAGE",
			},
		},
		{
			"default header with color",
			false,
			false,
			false,
			true,
			[]string{
				"NODE NAME",
				"NAMESPACE",
				"POD NAME",
				"POD AGE",
				"POD IP",
				colorStatus,
				"CONTAINER",
				"CPU/req",
				"CPU/lim",
				"MEM/req",
				"MEM/lim",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			o := &FreeOptions{
				listContainerImage: test.listImage,
				noHeaders:          test.noheader,
				noMetrics:          test.nometrics,
				nocolor:            test.nocolor,
			}
			o.prepareListTableHeader()

			if !reflect.DeepEqual(o.listTableHeaders, test.expected) {
				t.Errorf(
					"[%s] expected(%v) differ (got: %v)",
					test.description,
					test.expected,
					o.listTableHeaders,
				)
				return
			}
		})
	}

}

func TestToUnit(t *testing.T) {

	var tests = []struct {
		description string
		input       int64
		binPrefix   bool
		withoutunit bool
		expected    string
	}{
		{"si prefix without unit", 12345, false, true, "12"},
		{"si prefix with unit", 6000, false, false, "6K"},
		{"binary prefix without unit", 12345, true, true, "12"},
		{"binary prefix with unit", 6000, true, false, "5Ki"},
		{"0 case", 0, true, false, "0Ki"},
	}

	o := &FreeOptions{
		bytes: false,
		kByte: true,
		mByte: false,
		gByte: false,
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			o.withoutUnit = test.withoutunit
			o.binPrefix = test.binPrefix
			actual := o.toUnit(test.input)
			if actual != test.expected {
				t.Errorf(
					"[%s] expected(%s) differ (got: %s)",
					test.description,
					test.expected,
					actual,
				)
				return
			}
		})
	}
}

func TestToUnitOrDash(t *testing.T) {

	var tests = []struct {
		description string
		input       int64
		binPrefix   bool
		withoutunit bool
		expected    string
	}{
		{"si prefix without unit", 12345, false, true, "12"},
		{"si prefix with unit", 6000, false, false, "6K"},
		{"binary prefix without unit", 12345, true, true, "12"},
		{"binary prefix with unit", 6000, true, false, "5Ki"},
		{"0 case", 0, true, false, "-"},
	}

	o := &FreeOptions{
		bytes: false,
		kByte: true,
		mByte: false,
		gByte: false,
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			o.withoutUnit = test.withoutunit
			o.binPrefix = test.binPrefix
			actual := o.toUnitOrDash(test.input)
			if actual != test.expected {
				t.Errorf(
					"[%s] expected(%s) differ (got: %s)",
					test.description,
					test.expected,
					actual,
				)
				return
			}
		})
	}
}

func TestToMilliUnitOrDash(t *testing.T) {

	var tests = []struct {
		description string
		i           int64
		withoutunit bool
		expected    string
	}{
		{"return dash", 0, false, "-"},
		{"return unit", 1500, false, "1500m"},
		{"return no unit", 1500, true, "1500"},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			o := &FreeOptions{
				withoutUnit: test.withoutunit,
			}
			actual := o.toMilliUnitOrDash(test.i)
			if actual != test.expected {
				t.Errorf(
					"[%s] expected(%s) differ (got: %s)",
					test.description,
					test.expected,
					actual,
				)
				return
			}
		})
	}
}

func TestToColorPercent(t *testing.T) {

	var tests = []struct {
		description string
		p           int64
		warn        int64
		crit        int64
		nocolor     bool
		expected    string
	}{
		{"p < w", 10, 25, 50, false, color.Green.Sprint("10%")},
		{"w < p < c", 30, 25, 50, false, color.Yellow.Sprint("30%")},
		{"c < p", 80, 25, 50, false, color.Red.Sprint("80%")},
		{"p < w with nocolor", 10, 25, 50, true, "10%"},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			o := &FreeOptions{
				warnThreshold: test.warn,
				critThreshold: test.crit,
				nocolor:       test.nocolor,
			}
			actual := o.toColorPercent(test.p)
			actualB := []byte(actual)
			expectedB := []byte(test.expected)
			if !bytes.Equal(actualB, expectedB) {
				t.Errorf(
					"[%s] expected(%v) differ (got: %v)",
					test.description,
					expectedB,
					actualB,
				)
				return
			}
		})
	}

}

// Test Helper
func executeCommand(root *cobra.Command, args ...string) (output string, err error) {
	_, output, err = executeCommandC(root, args...)
	return output, err
}

func executeCommandC(root *cobra.Command, args ...string) (c *cobra.Command, output string, err error) {
	buf := new(bytes.Buffer)
	root.SetOutput(buf)
	root.SetArgs(args)

	c, err = root.ExecuteC()

	return c, buf.String(), err
}

func prepareTestNodeMetricsClient() *fakemetrics.Clientset {
	fakeMetricsClient := &fakemetrics.Clientset{}
	fakeMetricsClient.AddReactor("get", "nodes", func(action core.Action) (handled bool, ret runtime.Object, err error) {
		return true, &testNodeMetrics.Items[0], nil
	})
	return fakeMetricsClient
}

func prepareTestPodMetricsClient() *fakemetrics.Clientset {
	fakeMetricsClient := &fakemetrics.Clientset{}
	fakeMetricsClient.AddReactor("list", "pods", func(action core.Action) (handled bool, ret runtime.Object, err error) {
		return true, testPodMetrics, nil
	})
	return fakeMetricsClient
}
