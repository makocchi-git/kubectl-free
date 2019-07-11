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
	"k8s.io/cli-runtime/pkg/genericclioptions"
	fake "k8s.io/client-go/kubernetes/fake"
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
		},
	},
	{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "node2",
			Labels: map[string]string{"hostname": "node1"},
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
			PodIP: "1.2.3.4",
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
			PodIP: "2.3.4.5",
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
		header:             "default",
	}

	actual := NewFreeOptions(streams)

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("expected(%#v) differ (got: %#v)", expected, actual)
	}
}

func TestNewCmdFree(t *testing.T) {

	rootCmd := NewCmdFree(
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

	// RunE Validation error 1
	t.Run("RunE validation error 1", func(t *testing.T) {

		c := NewCmdFree(
			genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr},
			"v0.0.1",
			"abcd123",
			"1234567890",
		)

		err := c.ParseFlags([]string{"--crit-threshold=5"})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}

		rerr := c.RunE(c, []string{})
		if rerr == nil {
			t.Errorf("unexpected error: should return error")
			return
		}

		expected := "can not set critical threshold less than warn threshold (warn:25 crit:5)"
		if rerr.Error() != expected {
			t.Errorf("expected(%s) differ (got: %s)", expected, rerr.Error())
			return
		}
	})

	// RunE Validation error 2
	t.Run("RunE validation error 2", func(t *testing.T) {

		c := NewCmdFree(
			genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr},
			"v0.0.1",
			"abcd123",
			"1234567890",
		)

		err := c.ParseFlags([]string{"--header", "hogehoge"})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}

		rerr := c.RunE(c, []string{})
		if rerr == nil {
			t.Errorf("unexpected error: should return error")
			return
		}

		expected := "invalid header option: hogehoge"
		if rerr.Error() != expected {
			t.Errorf("expected(%s) differ (got: %s)", expected, rerr.Error())
			return
		}
	})
}

func TestPrepare(t *testing.T) {

	t.Run("prepare", func(t *testing.T) {

		o := &FreeOptions{
			configFlags: genericclioptions.NewConfigFlags(true),
		}

		if err := o.Prepare(); err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
	})

	t.Run("prepare allnamespace", func(t *testing.T) {

		o := &FreeOptions{
			configFlags:   genericclioptions.NewConfigFlags(true),
			allNamespaces: true,
		}

		if err := o.Prepare(); err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
	})

	t.Run("prepare specific namespace", func(t *testing.T) {

		o := &FreeOptions{
			configFlags: genericclioptions.NewConfigFlags(true),
		}
		*o.configFlags.Namespace = "awesome-ns"

		if err := o.Prepare(); err != nil {
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

	t.Run("validate header option", func(t *testing.T) {

		o := &FreeOptions{
			header: "foobar",
		}

		err := o.Validate()
		expected := "invalid header option: foobar"
		if err.Error() != expected {
			t.Errorf("unexpected error: %v", err)
			return
		}
	})

	t.Run("validate success", func(t *testing.T) {

		o := &FreeOptions{
			header:        "default",
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
		list        bool
		expected    []string
		expectedErr error
	}{
		{
			"default free",
			[]string{},
			false,
			[]string{
				"node1   NotReady   1     4     25%   1K    4K    25%",
				"",
			},
			nil,
		},
		{
			"default free --list",
			[]string{},
			true,
			[]string{
				"node1   pod1         Running   default   container1   1     2     1K    2K",
				"",
			},
			nil,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {

			fakeNodeClient := fake.NewSimpleClientset(&testNodes[0])
			fakePodClient := fake.NewSimpleClientset(&testPods[0])

			buffer := &bytes.Buffer{}
			o := &FreeOptions{
				nocolor:    true,
				table:      table.NewOutputTable(buffer),
				list:       test.list,
				nodeClient: fakeNodeClient.CoreV1().Nodes(),
				podClient:  fakePodClient.CoreV1().Pods("default"),
				header:     "none",
			}

			if err := o.Run(test.args); !reflect.DeepEqual(err, test.expectedErr) {
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
	colorCPUp := "CPU/%"
	colorMEMp := "MEM/%"
	util.DefaultColor(&colorStatus)
	util.DefaultColor(&colorCPUp)
	util.DefaultColor(&colorMEMp)

	var tests = []struct {
		description string
		listPod     bool
		nocolor     bool
		header      string
		expected    []string
	}{
		{
			"default header",
			false,
			true,
			"default",
			[]string{
				"NAME",
				"STATUS",
				"CPU/req",
				"CPU/alloc",
				"CPU/%",
				"MEM/req",
				"MEM/alloc",
				"MEM/%",
			},
		},
		{
			"verbose header",
			false,
			true,
			"verbose",
			[]string{
				"NAME",
				"STATUS",
				"CPU requested",
				"CPU allocatable",
				"CPU %USED",
				"Memory requested",
				"Memory allocatable",
				"Memory %USED",
			},
		},
		{
			"none header",
			false,
			true,
			"none",
			[]string{},
		},
		{
			"default header with --pod",
			true,
			true,
			"default",
			[]string{
				"NAME",
				"STATUS",
				"CPU/req",
				"CPU/alloc",
				"CPU/%",
				"MEM/req",
				"MEM/alloc",
				"MEM/%",
				"PODS",
				"PODS/alloc",
				"CONTAINERS",
			},
		},
		{
			"verbose header with --pod",
			true,
			true,
			"verbose",
			[]string{
				"NAME",
				"STATUS",
				"CPU requested",
				"CPU allocatable",
				"CPU %USED",
				"Memory requested",
				"Memory allocatable",
				"Memory %USED",
				"PODS",
				"PODS allocation",
				"CONTAINERS",
			},
		},
		{
			"none header with --pod",
			true,
			true,
			"none",
			[]string{},
		},
		{
			"default header with color",
			false,
			false,
			"default",
			[]string{
				"NAME",
				colorStatus,
				"CPU/req",
				"CPU/alloc",
				colorCPUp,
				"MEM/req",
				"MEM/alloc",
				colorMEMp,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			o := &FreeOptions{
				pod:     test.listPod,
				header:  test.header,
				nocolor: test.nocolor,
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
		header      string
		expected    []string
	}{
		{
			"default header",
			false,
			true,
			"default",
			[]string{
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
			},
		},
		{
			"verbose header",
			false,
			true,
			"verbose",
			[]string{
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
			},
		},
		{
			"none header",
			false,
			true,
			"none",
			[]string{},
		},
		{
			"default header with --list-image",
			true,
			true,
			"default",
			[]string{
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
				"IMAGE",
			},
		},
		{
			"verbose header with --list-image",
			true,
			true,
			"verbose",
			[]string{
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
				"IMAGE",
			},
		},
		{
			"none header with --list-image",
			true,
			true,
			"none",
			[]string{},
		},
		{
			"default header with color",
			false,
			false,
			"default",
			[]string{
				"NODE NAME",
				"POD",
				"POD IP",
				colorStatus,
				"NAMESPACE",
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
				header:             test.header,
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

func TestShowFree(t *testing.T) {

	var tests = []struct {
		description string
		pod         bool
		namespace   string
		expected    []string
		expectedErr error
	}{
		{
			"default free",
			false,
			"default",
			[]string{
				"node1   NotReady   1     4     25%   1K    4K    25%",
				"",
			},
			nil,
		},
		{
			"default free --pod",
			true,
			"default",
			[]string{
				"node1   NotReady   1     4     25%   1K    4K    25%   1     110   1",
				"",
			},
			nil,
		},
		{
			"awesome-ns free",
			true,
			"awesome-ns",
			[]string{
				"node1   NotReady   200m   4     5%    0K    4K    7%    1     110   2",
				"",
			},
			nil,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {

			fakeNodeClient := fake.NewSimpleClientset(&testNodes[0])
			fakePodClient := fake.NewSimpleClientset(&testPods[0], &testPods[2])

			buffer := &bytes.Buffer{}
			o := &FreeOptions{
				nocolor:    true,
				table:      table.NewOutputTable(buffer),
				list:       false,
				pod:        test.pod,
				nodeClient: fakeNodeClient.CoreV1().Nodes(),
				podClient:  fakePodClient.CoreV1().Pods(test.namespace),
				header:     "none",
			}

			if err := o.showFree([]v1.Node{testNodes[0]}); err != nil {
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

	t.Run("Allnamespace", func(t *testing.T) {

		fakeNodeClient := fake.NewSimpleClientset(&testNodes[0])
		fakePodClient := fake.NewSimpleClientset(&testPods[0], &testPods[1], &testPods[2])

		buffer := &bytes.Buffer{}
		o := &FreeOptions{
			nocolor:       true,
			table:         table.NewOutputTable(buffer),
			list:          false,
			pod:           false,
			nodeClient:    fakeNodeClient.CoreV1().Nodes(),
			podClient:     fakePodClient.CoreV1().Pods(""),
			allNamespaces: true,
			header:        "none",
		}

		if err := o.showFree([]v1.Node{testNodes[0]}); err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}

		expected := []string{
			"node1   NotReady   1700m   4     42%   2K    4K    57%",
			"",
		}
		e := strings.Join(expected, "\n")
		if buffer.String() != e {
			t.Errorf("expected(%s) differ (got: %s)", e, buffer.String())
			return
		}

	})
}

func TestListPodsOnNode(t *testing.T) {

	var tests = []struct {
		description   string
		listContainer bool
		listAll       bool
		expected      []string
	}{
		{
			"list container: false, list all: false",
			false,
			false,
			[]string{
				"node2   pod2   1.2.3.4   Running   default   container2a   500m   500m   1K    1K",
				"",
			},
		},
		{
			"list container: true, list all: false",
			true,
			false,
			[]string{
				"node2   pod2   1.2.3.4   Running   default   container2a   500m   500m   1K    1K    nginx:latest",
				"",
			},
		},
		{
			"list container: false, list all: true",
			false,
			true,
			[]string{
				"node2   pod2   1.2.3.4   Running   default   container2a   500m   500m   1K    1K",
				"node2   pod2   1.2.3.4   Running   default   container2b   -      -      -     -",
				"",
			},
		},
		{
			"list container: true, list all: true",
			true,
			true,
			[]string{
				"node2   pod2   1.2.3.4   Running   default   container2a   500m   500m   1K    1K    nginx:latest",
				"node2   pod2   1.2.3.4   Running   default   container2b   -      -      -     -     busybox:latest",
				"",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			buffer := &bytes.Buffer{}
			fakeClient := fake.NewSimpleClientset(&testPods[1])
			o := &FreeOptions{
				table:              table.NewOutputTable(buffer),
				podClient:          fakeClient.CoreV1().Pods(""),
				header:             "none",
				nocolor:            true,
				listContainerImage: test.listContainer,
				listAll:            test.listAll,
			}

			if err := o.listPodsOnNode([]v1.Node{testNodes[1]}); err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			expected := strings.Join(test.expected, "\n")
			actual := buffer.String()
			if actual != expected {
				t.Errorf("[%s] expected(%s) differ (got: %s)", test.description, expected, actual)
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
